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
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestCompareEquals(t *testing.T) {
	got := Compare("It reads many bits\nIt exchanges many bits\nIt writes many bits\n",
		"github.com/google/golden/testdata/haiku.txt.golden")
	want := ""
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestCompareNotEquals(t *testing.T) {
	got := Compare("It reads many bits\nIt exchanges twenty bits\nIt writes many bits\n",
		"github.com/google/golden/testdata/haiku.txt.golden")
	want := `Actual data differs from golden data; run "go test -update_golden" to update
--- github.com/google/golden/testdata/haiku.txt.golden
+++ github.com/google/golden/testdata/haiku.txt.actual
@@ -1,4 +1,4 @@
 It reads many bits
-It exchanges many bits
+It exchanges twenty bits
 It writes many bits
 
`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestUpdateGolden(t *testing.T) {
	dir, err := ioutil.TempDir("", "goldendata_test")
	if err != nil {
		t.Fatalf("Unable to create temporary directory: %v", err)
	}
	defer os.RemoveAll(dir)
	if err := os.MkdirAll(path.Join(dir, "src/fake/testdata"), 0700); err != nil {
		t.Fatalf("Unable to create testdata directory: %v", err)
	}
	originalGoPath := build.Default.GOPATH
	build.Default.GOPATH = dir
	defer func() {
		build.Default.GOPATH = originalGoPath
	}()
	*updateGolden = true
	defer func() { *updateGolden = false }()
	{
		want := ""
		got := Compare("This is the new contents", "fake/testdata/haiku.txt.golden")
		if got != want {
			t.Errorf("compare output: got %q, want %q", got, want)
		}
	}
	{
		got, err := ioutil.ReadFile(path.Join(dir, "src/fake/testdata/haiku.txt.golden"))
		if err != nil {
			t.Error(err)
		}
		want := "This is the new contents"
		if string(got) != want {
			t.Errorf("written contents: got %q, want %q", string(got), want)
		}
	}
}
