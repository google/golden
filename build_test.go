// Copyright 2017 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package golden

import (
	"bytes"
	"fmt"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"text/template"
)

func TestSortedKeys(t *testing.T) {
	var tests = []struct {
		in  map[string]bool
		out []string
	}{
		{
			in:  map[string]bool{"a": true, "c": true, "b": true},
			out: []string{"a", "b", "c"},
		},
		{
			in:  map[string]bool{"a": true},
			out: []string{"a"},
		},
		{
			in:  map[string]bool{},
			out: []string{},
		},
	}
	for _, test := range tests {
		got := sortedKeys(test.in)
		if !reflect.DeepEqual(got, test.out) {
			t.Errorf("sortedKeys(%v); got %v want %v", test.in, got, test.out)
		}
	}
}

func TestGetFullPath(t *testing.T) {
	type test struct {
		function func(string) (string, error)
		in       string
		out      string
		err      string
	}
	var environments = []struct {
		goPath    string
		fakeFiles []string
		tests     []test
	}{
		{
			goPath:    "",
			fakeFiles: []string{},
			tests: []test{
				{
					function: getFullPathForRead,
					in:       "github.com/google/foobar/hi.txt",
					err:      "GOPATH is empty",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/foobar/hi.txt",
					err:      "GOPATH is empty",
				},
			},
		},
		{
			goPath: "{{.TempDir}}/p1",
			fakeFiles: []string{
				"{{.TempDir}}/p1/src/github.com/google/foobar/hi.txt",
			},
			tests: []test{
				{
					function: getFullPathForRead,
					in:       "github.com/google/foobar/hi.txt",
					out:      "{{.TempDir}}/p1/src/github.com/google/foobar/hi.txt",
				},
				{
					function: getFullPathForRead,
					in:       "github.com/google/foobar/nosuchfile.txt",
					err:      "github.com/google/foobar/nosuchfile.txt: file not found in GOPATH",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/foobar/hi.txt",
					out:      "{{.TempDir}}/p1/src/github.com/google/foobar/hi.txt",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/foobar/new.txt",
					out:      "{{.TempDir}}/p1/src/github.com/google/foobar/new.txt",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/newdir/new.txt",
					out:      "{{.TempDir}}/p1/src/github.com/google/newdir/new.txt",
				},
			},
		},
		{
			goPath: "{{.TempDir}}/p1{{.PathListSeparator}}{{.TempDir}}/p2",
			fakeFiles: []string{
				"{{.TempDir}}/p1/src/github.com/google/foobar/a.txt",
				"{{.TempDir}}/p2/src/github.com/google/foobar/b.txt",
				"{{.TempDir}}/p1/src/github.com/google/foobar/d.txt",
				"{{.TempDir}}/p2/src/github.com/google/foobar/d.txt",
				"{{.TempDir}}/p1/src/github.com/google/baz/",
				"{{.TempDir}}/p2/src/github.com/google/bar/",
			},
			tests: []test{
				{
					function: getFullPathForRead,
					in:       "github.com/google/foobar/a.txt",
					out:      "{{.TempDir}}/p1/src/github.com/google/foobar/a.txt",
				},
				{
					function: getFullPathForRead,
					in:       "github.com/google/foobar/b.txt",
					out:      "{{.TempDir}}/p2/src/github.com/google/foobar/b.txt",
				},
				{
					function: getFullPathForRead,
					in:       "github.com/google/foobar/c.txt",
					err:      "github.com/google/foobar/c.txt: file not found in GOPATH",
				},
				{
					function: getFullPathForRead,
					in:       "github.com/google/foobar/d.txt",
					out:      "{{.TempDir}}/p1/src/github.com/google/foobar/d.txt",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/foobar/a.txt",
					out:      "{{.TempDir}}/p1/src/github.com/google/foobar/a.txt",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/foobar/b.txt",
					out:      "{{.TempDir}}/p2/src/github.com/google/foobar/b.txt",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/foobar/c.txt",
					err:      "there are multiple suitable directories in the GOPATH: [{{.TempDir}}/p1/src/github.com/google/foobar/c.txt {{.TempDir}}/p2/src/github.com/google/foobar/c.txt]",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/foobar/d.txt",
					err:      "there are multiple files in the GOPATH with the same relative path \"github.com/google/foobar/d.txt\": [{{.TempDir}}/p1/src/github.com/google/foobar/d.txt {{.TempDir}}/p2/src/github.com/google/foobar/d.txt]",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/baz/a.txt",
					out:      "{{.TempDir}}/p1/src/github.com/google/baz/a.txt",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/bar/a.txt",
					out:      "{{.TempDir}}/p2/src/github.com/google/bar/a.txt",
				},
				{
					function: getFullPathForWrite,
					in:       "github.com/google/nosuchdir/a.txt",
					err:      "none of these directories in the GOPATH exist: [{{.TempDir}}/p1/src/github.com/google/nosuchdir {{.TempDir}}/p2/src/github.com/google/nosuchdir]",
				},
			},
		},
	}
	for envIndex, env := range environments {
		// Add scope for defer statements inside...
		func() {
			tempDir, err := ioutil.TempDir("", "goldendata_test")
			if err != nil {
				t.Fatalf("Unable to create temporary directory: %v", err)
			}
			defer os.RemoveAll(tempDir)
			expandTemplate := func(tmplStr string) string {
				tmpl, err := template.New("").Parse(tmplStr)
				if err != nil {
					t.Fatalf("Error parsing template %q: %v", tmplStr, err)
				}
				buf := bytes.NewBuffer([]byte{})
				err = tmpl.Execute(buf, struct{ TempDir, PathListSeparator string }{
					TempDir:           tempDir,
					PathListSeparator: string(filepath.ListSeparator),
				})
				if err != nil {
					t.Fatalf("Error executing template: %v", err)
				}
				return buf.String()
			}

			originalGoPath := build.Default.GOPATH
			build.Default.GOPATH = expandTemplate(env.goPath)
			defer func() {
				build.Default.GOPATH = originalGoPath
			}()
			for _, pathTemplate := range env.fakeFiles {
				fullPath := expandTemplate(pathTemplate)
				fullDir := path.Dir(fullPath)
				if err := os.MkdirAll(fullDir, 0755); err != nil {
					t.Fatalf("Error making 'fake' directory: %v", err)
				}
				if !strings.HasSuffix(fullPath, "/") {
					if err := ioutil.WriteFile(fullPath, []byte(""), 644); err != nil {
						t.Fatalf("Cannot write fake file: %v", err)
					}
				}
			}
			for testIndex, test := range env.tests {
				testID := fmt.Sprintf("env #%v test #%v", envIndex, testIndex)
				want := expandTemplate(test.out)
				got, err := test.function(test.in)
				if got != want {
					t.Errorf("%v: %q: got %q want %q", testID, test.in, got, want)
				}
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				wantErrStr := expandTemplate(test.err)
				if errStr != wantErrStr {
					t.Errorf("%v: %q error: got %q want %q", testID, test.in, errStr, wantErrStr)
				}
			}
		}()
	}
}
