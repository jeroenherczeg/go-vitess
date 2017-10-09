/*
Copyright 2017 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sqlparser

import (
	"io"
	"strings"
	"testing"
)

// TestParseNextValid concatenates all the valid SQL test cases and check it can read
// them as one long string.
func TestParseNextValid(t *testing.T) {
	// Test only the first N queries, because after that we hit errors
	// TODO Remove this limit
	testLimit := len(validSQL)

	var sql string
	for _, tcase := range validSQL[:testLimit] {
		sql += strings.TrimSuffix(tcase.input, ";") + ";"
	}

	tokens := NewStringTokenizer(sql)
	for i, tcase := range validSQL[:testLimit] {
		input := tcase.input + ";"
		want := tcase.output
		if want == "" {
			want = tcase.input
		}

		tree, err := ParseNext(tokens)
		if err != nil {
			t.Fatalf("[%d] ParseNext(%q) err: %q, want nil", i, input, err)
			continue
		}

		if got := String(tree); got != want {
			t.Fatalf("[%d] ParseNext(%q) = %q, want %q", i, input, got, want)
		}
	}

	// Read one more and it should be EOF
	if tree, err := ParseNext(tokens); err != io.EOF {
		t.Errorf("ParseNext(tokens) = (%q, %v) want io.EOF", String(tree), err)
	}
}

// TestParseNextEdgeCases tests various ParseNext edge cases
func TestParseNextEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{{
		name:  "Trailing ;",
		input: "select 1 from a; update a set b = 2;",
		want:  []string{"select 1 from a", "update a set b = 2"},
	}, {
		name:  "No trailing ;",
		input: "select 1 from a; update a set b = 2",
		want:  []string{"select 1 from a", "update a set b = 2"},
	}, {
		name:  "Trailing whitespace",
		input: "select 1 from a; update a set b = 2    ",
		want:  []string{"select 1 from a", "update a set b = 2"},
	}, {
		name:  "Trailing whitespace and ;",
		input: "select 1 from a; update a set b = 2   ;   ",
		want:  []string{"select 1 from a", "update a set b = 2"},
	}, {
		name:  "Handle ForceEOF statements",
		input: "set character set utf8; select 1 from a",
		want:  []string{"set ", "select 1 from a"},
	}, {
		name:  "Semicolin inside a string",
		input: "set character set ';'; select 1 from a",
		want:  []string{"set ", "select 1 from a"},
	}, {
		name:  "Partial DDL",
		input: "create table a; select 1 from a",
		want:  []string{"create table a", "select 1 from a"},
	}, {
		name:  "Partial DDL",
		input: "create table a ignore me this is garbage; select 1 from a",
		want:  []string{"create table a", "select 1 from a"},
	}}

	for _, test := range tests {
		tokens := NewStringTokenizer(test.input)

		for i, want := range test.want {
			tree, err := ParseNext(tokens)
			if err != nil {
				t.Fatalf("[%d] ParseNext(%q) err = %q, want nil", i, test.input, err)
				continue
			}

			if got := String(tree); got != want {
				t.Fatalf("[%d] ParseNext(%q) = %q, want %q", i, test.input, got, want)
			}
		}

		// Read one more and it should be EOF
		if tree, err := ParseNext(tokens); err != io.EOF {
			t.Errorf("ParseNext(%q) = (%q, %v) want io.EOF", test.input, String(tree), err)
		}

		// And again, one more should be EOF
		if tree, err := ParseNext(tokens); err != io.EOF {
			t.Errorf("ParseNext(%q) = (%q, %v) want io.EOF", test.input, String(tree), err)
		}
	}
}

// TODO Add Tests to check the resyncing after errors.
// "sel ga blah ab a; select 1 from a;""
// "; select 1"
// "   ; select 1"
