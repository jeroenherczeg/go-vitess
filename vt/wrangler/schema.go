// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package wrangler

import (
	"bytes"
	"fmt"
	"html/template"
	"sort"
	"sync"
	"time"

	"golang.org/x/net/context"

	log "github.com/golang/glog"
	"github.com/youtube/vitess/go/vt/concurrency"
	"github.com/youtube/vitess/go/vt/mysqlctl/tmutils"
	"github.com/youtube/vitess/go/vt/schemamanager"
	"github.com/youtube/vitess/go/vt/topo"
	"github.com/youtube/vitess/go/vt/topo/topoproto"

	tabletmanagerdatapb "github.com/youtube/vitess/go/vt/proto/tabletmanagerdata"
	topodatapb "github.com/youtube/vitess/go/vt/proto/topodata"
)

// GetSchema uses an RPC to get the schema from a remote tablet
func (wr *Wrangler) GetSchema(ctx context.Context, tabletAlias *topodatapb.TabletAlias, tables, excludeTables []string, includeViews bool) (*tabletmanagerdatapb.SchemaDefinition, error) {
	ti, err := wr.ts.GetTablet(ctx, tabletAlias)
	if err != nil {
		return nil, err
	}

	return wr.tmc.GetSchema(ctx, ti.Tablet, tables, excludeTables, includeViews)
}

// ReloadSchema forces the remote tablet to reload its schema.
func (wr *Wrangler) ReloadSchema(ctx context.Context, tabletAlias *topodatapb.TabletAlias) error {
	ti, err := wr.ts.GetTablet(ctx, tabletAlias)
	if err != nil {
		return err
	}

	return wr.tmc.ReloadSchema(ctx, ti.Tablet)
}

// helper method to asynchronously diff a schema
func (wr *Wrangler) diffSchema(ctx context.Context, masterSchema *tabletmanagerdatapb.SchemaDefinition, masterTabletAlias, alias *topodatapb.TabletAlias, excludeTables []string, includeViews bool, wg *sync.WaitGroup, er concurrency.ErrorRecorder) {
	defer wg.Done()
	log.Infof("Gathering schema for %v", topoproto.TabletAliasString(alias))
	slaveSchema, err := wr.GetSchema(ctx, alias, nil, excludeTables, includeViews)
	if err != nil {
		er.RecordError(err)
		return
	}

	log.Infof("Diffing schema for %v", topoproto.TabletAliasString(alias))
	tmutils.DiffSchema(topoproto.TabletAliasString(masterTabletAlias), masterSchema, topoproto.TabletAliasString(alias), slaveSchema, er)
}

// ValidateSchemaShard will diff the schema from all the tablets in the shard.
func (wr *Wrangler) ValidateSchemaShard(ctx context.Context, keyspace, shard string, excludeTables []string, includeViews bool) error {
	si, err := wr.ts.GetShard(ctx, keyspace, shard)
	if err != nil {
		return err
	}

	// get schema from the master, or error
	if !si.HasMaster() {
		return fmt.Errorf("No master in shard %v/%v", keyspace, shard)
	}
	log.Infof("Gathering schema for master %v", topoproto.TabletAliasString(si.MasterAlias))
	masterSchema, err := wr.GetSchema(ctx, si.MasterAlias, nil, excludeTables, includeViews)
	if err != nil {
		return err
	}

	// read all the aliases in the shard, that is all tablets that are
	// replicating from the master
	aliases, err := wr.ts.FindAllTabletAliasesInShard(ctx, keyspace, shard)
	if err != nil {
		return err
	}

	// then diff with all slaves
	er := concurrency.AllErrorRecorder{}
	wg := sync.WaitGroup{}
	for _, alias := range aliases {
		if topoproto.TabletAliasEqual(alias, si.MasterAlias) {
			continue
		}

		wg.Add(1)
		go wr.diffSchema(ctx, masterSchema, si.MasterAlias, alias, excludeTables, includeViews, &wg, &er)
	}
	wg.Wait()
	if er.HasErrors() {
		return fmt.Errorf("Schema diffs: %v", er.Error().Error())
	}
	return nil
}

// ValidateSchemaKeyspace will diff the schema from all the tablets in
// the keyspace.
func (wr *Wrangler) ValidateSchemaKeyspace(ctx context.Context, keyspace string, excludeTables []string, includeViews bool) error {
	// find all the shards
	shards, err := wr.ts.GetShardNames(ctx, keyspace)
	if err != nil {
		return err
	}

	// corner cases
	if len(shards) == 0 {
		return fmt.Errorf("No shards in keyspace %v", keyspace)
	}
	sort.Strings(shards)
	if len(shards) == 1 {
		return wr.ValidateSchemaShard(ctx, keyspace, shards[0], excludeTables, includeViews)
	}

	// find the reference schema using the first shard's master
	si, err := wr.ts.GetShard(ctx, keyspace, shards[0])
	if err != nil {
		return err
	}
	if !si.HasMaster() {
		return fmt.Errorf("No master in shard %v/%v", keyspace, shards[0])
	}
	referenceAlias := si.MasterAlias
	log.Infof("Gathering schema for reference master %v", topoproto.TabletAliasString(referenceAlias))
	referenceSchema, err := wr.GetSchema(ctx, referenceAlias, nil, excludeTables, includeViews)
	if err != nil {
		return err
	}

	// then diff with all other tablets everywhere
	er := concurrency.AllErrorRecorder{}
	wg := sync.WaitGroup{}

	// first diff the slaves in the reference shard 0
	aliases, err := wr.ts.FindAllTabletAliasesInShard(ctx, keyspace, shards[0])
	if err != nil {
		return err
	}

	for _, alias := range aliases {
		if topoproto.TabletAliasEqual(alias, si.MasterAlias) {
			continue
		}

		wg.Add(1)
		go wr.diffSchema(ctx, referenceSchema, referenceAlias, alias, excludeTables, includeViews, &wg, &er)
	}

	// then diffs all tablets in the other shards
	for _, shard := range shards[1:] {
		si, err := wr.ts.GetShard(ctx, keyspace, shard)
		if err != nil {
			er.RecordError(err)
			continue
		}

		if !si.HasMaster() {
			er.RecordError(fmt.Errorf("No master in shard %v/%v", keyspace, shard))
			continue
		}

		aliases, err := wr.ts.FindAllTabletAliasesInShard(ctx, keyspace, shard)
		if err != nil {
			er.RecordError(err)
			continue
		}

		for _, alias := range aliases {
			wg.Add(1)
			go wr.diffSchema(ctx, referenceSchema, referenceAlias, alias, excludeTables, includeViews, &wg, &er)
		}
	}
	wg.Wait()
	if er.HasErrors() {
		return fmt.Errorf("Schema diffs: %v", er.Error().Error())
	}
	return nil
}

// PreflightSchema will try a schema change on the remote tablet.
func (wr *Wrangler) PreflightSchema(ctx context.Context, tabletAlias *topodatapb.TabletAlias, changes []string) ([]*tabletmanagerdatapb.SchemaChangeResult, error) {
	ti, err := wr.ts.GetTablet(ctx, tabletAlias)
	if err != nil {
		return nil, err
	}
	return wr.tmc.PreflightSchema(ctx, ti.Tablet, changes)
}

// ApplySchemaKeyspace applies a schema change to an entire keyspace.
// Takes a keyspace lock to do this.
// First we will validate the Preflight works the same on all shard masters
// and fail if not (unless force is specified).
func (wr *Wrangler) ApplySchemaKeyspace(ctx context.Context, keyspace, change string, allowLongUnavailability bool, waitSlaveTimeout time.Duration) (err error) {
	// lock the keyspace
	ctx, unlock, lockErr := wr.ts.LockKeyspace(ctx, keyspace, "ApplySchemaKeyspace")
	if lockErr != nil {
		return lockErr
	}
	defer unlock(&err)

	// apply the schema change
	executor := schemamanager.NewTabletExecutor(wr.tmc, wr.ts)
	if allowLongUnavailability {
		executor.AllowBigSchemaChange()
	}
	return schemamanager.Run(
		ctx,
		schemamanager.NewPlainController(change, keyspace),
		executor,
	)
}

// CopySchemaShardFromShard copies the schema from a source shard to the specified destination shard.
// For both source and destination it picks the master tablet. See also CopySchemaShard.
func (wr *Wrangler) CopySchemaShardFromShard(ctx context.Context, tables, excludeTables []string, includeViews bool, sourceKeyspace, sourceShard, destKeyspace, destShard string) error {
	sourceShardInfo, err := wr.ts.GetShard(ctx, sourceKeyspace, sourceShard)
	if err != nil {
		return err
	}

	return wr.CopySchemaShard(ctx, sourceShardInfo.MasterAlias, tables, excludeTables, includeViews, destKeyspace, destShard)
}

// CopySchemaShard copies the schema from a source tablet to the
// specified shard.  The schema is applied directly on the master of
// the destination shard, and is propogated to the replicas through
// binlogs.
func (wr *Wrangler) CopySchemaShard(ctx context.Context, sourceTabletAlias *topodatapb.TabletAlias, tables, excludeTables []string, includeViews bool, destKeyspace, destShard string) error {
	destShardInfo, err := wr.ts.GetShard(ctx, destKeyspace, destShard)
	if err != nil {
		return err
	}

	diffs, err := wr.compareSchemas(ctx, sourceTabletAlias, destShardInfo.MasterAlias, tables, excludeTables, includeViews)
	if err != nil {
		return fmt.Errorf("CopySchemaShard failed because schemas could not be compared initially: %v", err)
	}
	if diffs == nil {
		// Return early because dest has already the same schema as source.
		return nil
	}

	sourceSd, err := wr.GetSchema(ctx, sourceTabletAlias, tables, excludeTables, includeViews)
	if err != nil {
		return err
	}
	createSQL := tmutils.SchemaDefinitionToSQLStrings(sourceSd)
	destTabletInfo, err := wr.ts.GetTablet(ctx, destShardInfo.MasterAlias)
	if err != nil {
		return err
	}
	for i, sqlLine := range createSQL {
		err = wr.applySQLShard(ctx, destTabletInfo, sqlLine, i == len(createSQL)-1)
		if err != nil {
			return err
		}
	}

	// Although the copy was successful, we have to verify it to catch the case
	// where the database already existed on the destination, but with different
	// options e.g. a different character set.
	// In that case, MySQL would have skipped our CREATE DATABASE IF NOT EXISTS
	// statement. We want to fail early in this case because vtworker SplitDiff
	// fails in case of such an inconsistency as well.
	diffs, err = wr.compareSchemas(ctx, sourceTabletAlias, destShardInfo.MasterAlias, tables, excludeTables, includeViews)
	if err != nil {
		return fmt.Errorf("CopySchemaShard failed because schemas could not be compared finally: %v", err)
	}
	if diffs != nil {
		return fmt.Errorf("CopySchemaShard was not successful because the schemas between the two tablets %v and %v differ: %v", sourceTabletAlias, destShardInfo.MasterAlias, diffs)
	}
	return nil
}

// compareSchemas returns nil if the schema of the two tablets referenced by
// "sourceAlias" and "destAlias" are identical. Otherwise, the difference is
// returned as []string.
func (wr *Wrangler) compareSchemas(ctx context.Context, sourceAlias, destAlias *topodatapb.TabletAlias, tables, excludeTables []string, includeViews bool) ([]string, error) {
	sourceSd, err := wr.GetSchema(ctx, sourceAlias, tables, excludeTables, includeViews)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema from tablet %v. err: %v", sourceAlias, err)
	}
	destSd, err := wr.GetSchema(ctx, destAlias, tables, excludeTables, includeViews)
	if err != nil {
		return nil, fmt.Errorf("failed to get schema from tablet %v. err: %v", destAlias, err)
	}
	return tmutils.DiffSchemaToArray("source", sourceSd, "dest", destSd), nil
}

// applySQLShard applies a given SQL change on a given tablet alias. It allows executing arbitrary
// SQL statements, but doesn't return any results, so it's only useful for SQL statements
// that would be run for their effects (e.g., CREATE).
// It works by applying the SQL statement on the shard's master tablet with replication turned on.
// Thus it should be used only for changes that can be applied on a live instance without causing issues;
// it shouldn't be used for anything that will require a pivot.
// The SQL statement string is expected to have {{.DatabaseName}} in place of the actual db name.
func (wr *Wrangler) applySQLShard(ctx context.Context, tabletInfo *topo.TabletInfo, change string, reloadSchema bool) error {
	filledChange, err := fillStringTemplate(change, map[string]string{"DatabaseName": tabletInfo.DbName()})
	if err != nil {
		return fmt.Errorf("fillStringTemplate failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	// Need to make sure that we enable binlog, since we're only applying the statement on masters.
	_, err = wr.tmc.ExecuteFetchAsDba(ctx, tabletInfo.Tablet, filledChange, 0, false, reloadSchema)
	return err
}

// fillStringTemplate returns the string template filled
func fillStringTemplate(tmpl string, vars interface{}) (string, error) {
	myTemplate := template.Must(template.New("").Parse(tmpl))
	data := new(bytes.Buffer)
	if err := myTemplate.Execute(data, vars); err != nil {
		return "", err
	}
	return data.String(), nil
}
