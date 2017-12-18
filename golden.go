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

// Package golden provides a function for comparing arbitrary text data to
// a golden file, and allowing a person running a test to automatically set the
// golden data to the "actual" data by setting a flag when running the unit
// test.
//
// This library is most useful when the data being compared is large, and
// correctness can be easily determined by having a human look at a diff.  Some
// examples are the output of an HTML template, or the output of a large
// computation formatted as a text protobuf.
//
// Expected usage:
//
//     func AUnitTest(t *testing.T) {
//       got := proto.MarshalTextString(code_under_test.ComputeTediousData(...))
//       if diff := golden.Compare(got, ".../testdata/data.txt.golden"); diff != "" {
//         t.Error(diff)
//       }
//     }
//
// When the user runs this test and the actual data differs from the golden
// file, they will see the following error message:
//
//     Actual data differs from golden data; run "go test -update_golden" to update
//     --- .../testdata/data.txt.golden
//     +++ .../testdata/data.txt.actual
//       blah: ""
//     - ultimate_answer: 41
//     + ultimate_answer: 42
//       foo: "bar"
//       baz: "blah"
//
// The user will inspect the diff. Let's say that this diff is due to a simple
// code change, and the new value is definitely correct. The user can then
// decide to overwrite the golden data with the actual data by re-running the
// test, passing the -update_golden flag:
//
//     $ go test -update_golden
//
// This time the test will succeed, and the golden data will be overwritten
// with the actual data. Code reviewers will notice in diffs that the golden
// data has been modified, and can easily compare the output of the code before
// and after the change.
package golden

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/pmezard/go-difflib/difflib"
)

// Compare compares the actual parameter to the contents of goldenFile and
// returns an empty string if they match. If they don't match, it returns a
// unified diff string documenting the differences.
//
// If the -update_golden flag is set, this function will overwrite the
// contents of goldenFile with the actual value. This is useful for updating
// the golden data automatically.
//
// goldenFile is a path relative to os.Getenv("GOROOT").
func Compare(actual string, goldenFile string) string {
	if shouldUpdateGolden() {
		fullPath, err := getFullPathForWrite(goldenFile)
		if err != nil {
			log.Fatalf("Error while getting path for writes: %v", err)
		}
		if err := ioutil.WriteFile(fullPath, []byte(actual), 0660); err != nil {
			log.Fatal(err)
		}
		return ""
	}

	fullPath, err := getFullPathForRead(goldenFile)
	if err != nil {
		log.Fatalf("Error while getting path for reads: %v", err)
	}

	expected, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Fatalf("Error while reading golden file: %v", err)
	}
	if string(expected) == actual {
		return ""
	}
	udiff := difflib.UnifiedDiff{
		A:        difflib.SplitLines(string(expected)),
		FromFile: goldenFile,
		B:        difflib.SplitLines(actual),
		ToFile:   strings.TrimSuffix(goldenFile, ".golden") + ".actual",
		Context:  3,
	}
	diffstr, err := difflib.GetUnifiedDiffString(udiff)
	if err != nil {
		log.Fatalf("Error computing unified diff with golden file: %v", err)
	}
	return fmt.Sprintf("Actual data differs from golden data; run %q to update\n%v", formatUpdateCommand(), diffstr)
}
