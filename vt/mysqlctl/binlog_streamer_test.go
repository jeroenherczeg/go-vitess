// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mysqlctl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/youtube/vitess/go/sync2"
)

func TestPosParse(t *testing.T) {
	line := "#131018 22:21:47 server id 41983  end_log_pos 286  group_id 7   Query   thread_id=3     exec_time=0     error_code=0"
	values := posRE.FindStringSubmatch(line)
	if len(values) != 4 {
		t.Fatalf("want 4, got %v", len(values))
	}
	if values[1] != "41983" {
		t.Errorf("want 41983, got %v", values[1])
	}
	if values[2] != "286" {
		t.Errorf("want 286, got %v", values[2])
	}
	if values[3] != "7" {
		t.Errorf("want 7, got %v", values[3])
	}

	line = "#131018 22:22:19 server id 41983  end_log_pos 372       Rotate to vt-0000041983-bin.000003  pos: 4"
	values = rotateRE.FindStringSubmatch(line)
	if len(values) != 3 {
		t.Fatalf("want 3, got %v", len(values))
	}
	if values[1] != "vt-0000041983-bin.000003" {
		t.Errorf("want vt-0000041983-bin.000003, got %v", values[1])
	}
	if values[2] != "4" {
		t.Errorf("want 4, got %v", values[2])
	}

	// Check for match even if there's a tab after the file name.
	line = "#131018 22:22:19 server id 41983  end_log_pos 372       Rotate to vt-0000041983-bin.000003\t  pos: 4"
	values = rotateRE.FindStringSubmatch(line)
	if len(values) != 3 {
		t.Fatalf("want 3, got %v", len(values))
	}
	if values[1] != "vt-0000041983-bin.000003" {
		t.Errorf("want vt-0000041983-bin.000003, got %v", values[1])
	}
	if values[2] != "4" {
		t.Errorf("want 4, got %v", values[2])
	}

	line = "DELIMITER /*!*/;\n"
	values = delimRE.FindStringSubmatch(line)
	if len(values) != 2 {
		t.Fatalf("want 3, got %v", len(values))
	}
	if values[1] != "/*!*/;" {
		t.Errorf("want /*!*/;, got %v", values[1])
	}
}

func TestFileInfo(t *testing.T) {
	fname := path.Join(os.TempDir(), "binlog_streamer.test")
	writer, err := os.Create(fname)
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(fname)
	var file fileInfo
	err = file.Init(fname, 0)
	if err != nil {
		t.Fatal(err)
	}
	ch := make(chan []byte, 10)
	var running = sync2.AtomicInt32(1)
	go func() {
		for {
			file.WaitForChange(&running)
			b := make([]byte, 128)
			n, err := file.handle.Read(b)
			if err != nil {
				ch <- []byte(err.Error())
			}
			file.Set(file.pos + int64(n))
			ch <- b[:n]
		}
	}()

	want := "Message1"
	writer.WriteString(want)
	writer.Sync()
	got := string(<-ch)
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}

	want = "Message2"
	writer.WriteString(want)
	writer.Sync()
	got = string(<-ch)
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}

	time.Sleep(200 * time.Millisecond)
	want = "Message3"
	writer.WriteString(want)
	writer.Sync()
	got = string(<-ch)
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}

	want = "EOF"
	running.Set(-1)
	got = string(<-ch)
	if want != got {
		t.Errorf("want %v, got %v", want, got)
	}
}

type fakeReader struct {
	toSend []byte
	err    error
}

func (fkr *fakeReader) Read(p []byte) (n int, err error) {
	if len(p) < len(fkr.toSend) {
		copy(p, fkr.toSend)
		fkr.toSend = fkr.toSend[len(p):]
		return len(p), nil
	}
	copy(p, fkr.toSend)
	n = len(fkr.toSend)
	fkr.toSend = nil
	return n, fkr.err
}

func TestReadEvent(t *testing.T) {
	// Error with 0 bytes
	fkreader := &fakeReader{
		toSend: nil,
		err:    fmt.Errorf("err1"),
	}
	reader := bufio.NewReaderSize(fkreader, 5)
	bls := &BinlogStreamer{}
	out, err := bls.readEvent(reader)
	if out != nil {
		t.Errorf("want nil, got %s", out)
	}
	want := "read error: err1"
	if err.Error() != want {
		t.Errorf("want %s, got %v", want, err)
	}

	// Error before reading \n
	longstr := "0123456789ABCDEFGH"
	fkreader = &fakeReader{
		toSend: []byte(longstr),
		err:    fmt.Errorf("err1"),
	}
	reader = bufio.NewReaderSize(fkreader, 5)
	bls = &BinlogStreamer{}
	out, err = bls.readEvent(reader)
	if string(out) != longstr {
		t.Errorf("want %s, got %s", longstr, out)
	}
	want = "read error: err1"
	if err.Error() != want {
		t.Errorf("want %s, got %v", want, err)
	}

	// '#' comment
	longstr = "#0123456789ABCDEFGH\n"
	fkreader = &fakeReader{
		toSend: []byte(longstr),
	}
	reader = bufio.NewReaderSize(fkreader, 5)
	bls = &BinlogStreamer{}
	out, err = bls.readEvent(reader)
	if string(out) != longstr[:len(longstr)-1] {
		t.Errorf("want %s, got %s", longstr[:len(longstr)-1], out)
	}
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}

	// '/*' comment
	longstr = "/*0123456789ABCDEFGH\n"
	fkreader = &fakeReader{
		toSend: []byte(longstr),
	}
	reader = bufio.NewReaderSize(fkreader, 5)
	bls = &BinlogStreamer{}
	out, err = bls.readEvent(reader)
	if string(out) != longstr[:len(longstr)-1] {
		t.Errorf("want %s, got %s", longstr[:len(longstr)-1], out)
	}
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}

	// delimeter
	longstr = "0123456789ABCDEFGH/*!*/;\n"
	fkreader = &fakeReader{
		toSend: []byte(longstr),
	}
	reader = bufio.NewReaderSize(fkreader, 5)
	bls = &BinlogStreamer{delim: []byte("/*!*/;")}
	out, err = bls.readEvent(reader)
	if string(out) != longstr[:len(longstr)-len(bls.delim)-1] {
		t.Errorf("want %s, got %s", longstr[:len(longstr)-len(bls.delim)-1], out)
	}
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}

	// \n before delimiter
	longstr = "0123456789ABCDEFGH\n/*!*/;\n"
	fkreader = &fakeReader{
		toSend: []byte(longstr),
	}
	reader = bufio.NewReaderSize(fkreader, 5)
	bls = &BinlogStreamer{delim: []byte("/*!*/;")}
	out, err = bls.readEvent(reader)
	if string(out) != longstr[:len(longstr)-len(bls.delim)-2] {
		t.Errorf("want %s, got %s", longstr[:len(longstr)-len(bls.delim)-2], out)
	}
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}

	// \n in middle of event
	longstr = "01234567\n89ABCDEFGH\n/*!*/;\n"
	fkreader = &fakeReader{
		toSend: []byte(longstr),
	}
	reader = bufio.NewReaderSize(fkreader, 5)
	bls = &BinlogStreamer{delim: []byte("/*!*/;")}
	out, err = bls.readEvent(reader)
	if string(out) != longstr[:len(longstr)-len(bls.delim)-2] {
		t.Errorf("want %s, got %s", longstr[:len(longstr)-len(bls.delim)-2], out)
	}
	if err != nil {
		t.Errorf("want nil, got %v", err)
	}
}

type transaction struct {
	Statements []string
	Position   BinlogPosition
}

func TestStream(t *testing.T) {
	env := setup("cat $3", 0)
	defer cleanup(env)

	var transactions []transaction

	out, err := ioutil.ReadFile("test/expected.json")
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(out, &transactions)
	if err != nil {
		t.Fatal(err)
	}

	curTransaction := 0
	bls := NewBinlogStreamer("db", "test/vt-0000041983-bin")
	err = bls.Stream("vt-0000041983-bin.000001", 0, func(reply interface{}) error {
		tx := reply.(*BinlogTransaction)
		for i, stmt := range tx.Statements {
			if transactions[curTransaction].Statements[i] != string(stmt) {
				t.Errorf("want %s, got %s", transactions[curTransaction].Statements[i], stmt)
			}
		}
		if transactions[curTransaction].Position != tx.Position {
			t.Errorf("want %#v, got %#v", transactions[curTransaction].Position, tx.Position)
		}
		curTransaction++
		if curTransaction == len(transactions) {
			bls.Stop()
		}
		// Uncomment the following lines to produce a different set of
		// expected outputs
		/*
			fmt.Printf("{\n\"Statements\": [\n")
			for i := 0; i < len(tx.Statements); i++ {
				if i == len(tx.Statements)-1 {
					fmt.Printf("%#v\n", string(tx.Statements[i]))
				} else {
					fmt.Printf("%#v,\n", string(tx.Statements[i]))
				}
			}
			fmt.Printf("],\n")
			fmt.Printf("\"Position\": {\"GroupId\": %d, \"ServerId\": %d}\n},\n", tx.Position.GroupId, tx.Position.ServerId)
		*/
		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
