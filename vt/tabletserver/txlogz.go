// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tabletserver

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"time"

	log "github.com/golang/glog"
	"github.com/youtube/vitess/go/acl"
	"github.com/youtube/vitess/go/vt/callerid"

	querypb "github.com/youtube/vitess/go/vt/proto/query"
	vtpb "github.com/youtube/vitess/go/vt/proto/vtrpc"
)

var (
	txlogzHeader = []byte(`
		<thead>
			<tr>
				<th>Transaction id</th>
				<th>Effective caller</th>
				<th>Immediate caller</th>
				<th>Start</th>
				<th>End</th>
				<th>Duration</th>
				<th>Decision</th>
				<th>Statements</th>
			</tr>
		</thead>
	`)
	txlogzFuncMap = template.FuncMap{
		"stampMicro":         func(t time.Time) string { return t.Format(time.StampMicro) },
		"getEffectiveCaller": func(e *vtpb.CallerID) string { return callerid.GetPrincipal(e) },
		"getImmediateCaller": func(i *querypb.VTGateCallerID) string { return callerid.GetUsername(i) },
	}
	txlogzTmpl = template.Must(template.New("example").Funcs(txlogzFuncMap).Parse(`
		<tr class="{{.ColorLevel}}">
			<td>{{.TransactionID}}</td>
			<td>{{.EffectiveCallerID | getEffectiveCaller}}</td>
			<td>{{.ImmediateCallerID | getImmediateCaller}}</td>
			<td>{{.StartTime | stampMicro}}</td>
			<td>{{.EndTime | stampMicro}}</td>
			<td>{{.Duration}}</td>
			<td>{{.Conclusion}}</td>
			<td>
				{{ range .Queries }}
					{{.}}<br>
				{{ end}}
			</td>
		</tr>`))
)

func init() {
	http.HandleFunc("/txlogz", txlogzHandler)
}

// txlogzHandler serves a human readable snapshot of the
// current transaction log.
// Endpoint: /txlogz?timeout=%d&limit=%d
// timeout: the txlogz will keep dumping transactions until timeout
// limit: txlogz will keep dumping transcations until it hits the limit
func txlogzHandler(w http.ResponseWriter, req *http.Request) {
	if err := acl.CheckAccessHTTP(req, acl.DEBUGGING); err != nil {
		acl.SendError(w, err)
		return
	}

	timeout, limit := parseTimeoutLimitParams(req)
	ch := TxLogger.Subscribe("txlogz")
	defer TxLogger.Unsubscribe(ch)
	startHTMLTable(w)
	defer endHTMLTable(w)
	w.Write(txlogzHeader)

	tmr := time.NewTimer(timeout)
	defer tmr.Stop()
	for i := 0; i < limit; i++ {
		select {
		case out := <-ch:
			txc, ok := out.(*TxConnection)
			if !ok {
				err := fmt.Errorf("Unexpected value in %s: %#v (expecting value of type %T)", TxLogger.Name(), out, &TxConnection{})
				io.WriteString(w, `<tr class="error">`)
				io.WriteString(w, err.Error())
				io.WriteString(w, "</tr>")
				log.Error(err)
				continue
			}
			duration := txc.EndTime.Sub(txc.StartTime).Seconds()
			var level string
			if duration < 0.1 {
				level = "low"
			} else if duration < 1.0 {
				level = "medium"
			} else {
				level = "high"
			}
			tmplData := struct {
				*TxConnection
				Duration   float64
				ColorLevel string
			}{txc, duration, level}
			if err := txlogzTmpl.Execute(w, tmplData); err != nil {
				log.Errorf("txlogz: couldn't execute template: %v", err)
			}
		case <-tmr.C:
			return
		}
	}
}
