// Copyright 2016, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package throttlerclienttest contains the testsuite against which each
// RPC implementation of the throttlerclient interface must be tested.
package throttlerclienttest

// NOTE: This file is not test-only code because it is referenced by
// tests in other packages and therefore it has to be regularly
// visible.

// NOTE: This code is in its own package such that its dependencies
// (e.g.  zookeeper) won't be drawn into production binaries as well.

import (
	"reflect"
	"testing"

	"golang.org/x/net/context"

	"github.com/youtube/vitess/go/vt/throttler"
	"github.com/youtube/vitess/go/vt/throttler/throttlerclient"
)

// TestSuite runs the test suite on the given throttlerclient and throttlerserver.
func TestSuite(t *testing.T, c throttlerclient.Client) {
	tf := &testFixture{}
	if err := tf.setUp(); err != nil {
		t.Fatal(err)
	}
	defer tf.tearDown()

	tf.setMaxRate(t, c)

	// TODO(mberlin): Add a test for panic handling.
}

var throttlerNames = []string{"t1", "t2"}

type testFixture struct {
	throttlers []*throttler.Throttler
}

func (tf *testFixture) setUp() error {
	for _, name := range throttlerNames {
		t, err := throttler.NewThrottler(name, "TPS", 1 /* threadCount */, 1, throttler.ReplicationLagModuleDisabled)
		if err != nil {
			return err
		}
		tf.throttlers = append(tf.throttlers, t)
	}
	return nil
}

func (tf *testFixture) tearDown() {
	for _, t := range tf.throttlers {
		t.Close()
	}
}

func (tf *testFixture) setMaxRate(t *testing.T, client throttlerclient.Client) {
	got, err := client.SetMaxRate(context.Background(), 23)
	if err != nil {
		t.Fatalf("Cannot execute remote command: %v", err)
	}

	if !reflect.DeepEqual(got, throttlerNames) {
		t.Fatalf("rate was not updated on all registered throttlers. got = %v, want = %v", got, throttlerNames)
	}
}
