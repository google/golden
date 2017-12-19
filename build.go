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

// Build system specific logic. Forks that use custom build systems can modify
// this file.

package golden

import (
	"flag"
	"fmt"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"sort"
)

var (
	// This flag is ONLY for use in tests.
	updateGolden = flag.Bool("update_golden", false, "Whether to update the golden files if they differ.")
)

func getFullPathForRead(relPath string) (string, error) {
	goPaths := filepath.SplitList(build.Default.GOPATH)
	if len(goPaths) == 0 {
		return "", fmt.Errorf("GOPATH is empty")
	}
	err := error(nil)
	for _, p := range goPaths {
		fullPath := path.Join(p, "src", relPath)
		_, err = os.Stat(fullPath)
		if err == nil {
			return fullPath, nil
		}
	}

	if os.IsNotExist(err) {
		return "", fmt.Errorf("%v: file not found in GOPATH", relPath)

	}
	return "", err
}

func sortedKeys(m map[string]bool) []string {
	result := make([]string, len(m))
	i := 0
	for k, _ := range m {
		result[i] = k
		i++
	}
	sort.Strings(result)
	return result
}

func getFullPathForWrite(relPath string) (string, error) {
	goPaths := filepath.SplitList(build.Default.GOPATH)
	if len(goPaths) == 0 {
		return "", fmt.Errorf("GOPATH is empty")
	}
	if len(goPaths) == 1 {
		// If there is only a single GOPATH, just use it
		return path.Join(goPaths[0], "src", relPath), nil
	}
	existingFiles := map[string]bool{}
	possibleDirectories := map[string]bool{}
	filesWithExistingDir := map[string]bool{}
	for _, p := range goPaths {
		fullPath := path.Join(p, "src", relPath)
		_, err := os.Stat(fullPath)
		if err == nil {
			existingFiles[fullPath] = true
		}
		fullDir := filepath.Dir(fullPath)
		possibleDirectories[fullDir] = true
		_, err = os.Stat(fullDir)
		if err == nil {
			filesWithExistingDir[fullPath] = true
		}
	}
	if len(existingFiles) > 1 {
		return "", fmt.Errorf("there are multiple files in the GOPATH with the same relative path %q: %v", relPath, sortedKeys(existingFiles))
	}

	if len(existingFiles) == 1 {
		for fullPath, _ := range existingFiles {
			return fullPath, nil
		}
	}
	if len(filesWithExistingDir) > 1 {
		return "", fmt.Errorf("there are multiple suitable directories in the GOPATH: %v", sortedKeys(filesWithExistingDir))
	}

	if len(filesWithExistingDir) == 1 {
		for fullPath, _ := range filesWithExistingDir {
			return fullPath, nil
		}
	}
	return "", fmt.Errorf("none of these directories in the GOPATH exist: %v", sortedKeys(possibleDirectories))
}

func shouldUpdateGolden() bool {
	return *updateGolden
}

func formatUpdateCommand() string {
	return "go test -update_golden"
}

func enableUpdateGoldenForTest(tmpdir string) func() {
	originalGoPath := build.Default.GOPATH
	originalUpdateGolden := *updateGolden

	build.Default.GOPATH = tmpdir
	*updateGolden = true

	restoreFunc := func() {
		build.Default.GOPATH = originalGoPath
		*updateGolden = originalUpdateGolden
	}
	return restoreFunc
}
