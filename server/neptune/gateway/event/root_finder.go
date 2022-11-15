// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package event

import (
	"context"
	"github.com/runatlantis/atlantis/server/logging"
	"os"
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/core/config/valid"

	"github.com/docker/docker/pkg/fileutils"
	"github.com/pkg/errors"
)

// RepoRootFinder implements rootFinder.
type RepoRootFinder struct {
	Logger logging.Logger
}

func (f *RepoRootFinder) FindRoots(ctx context.Context, config valid.RepoCfg, absRepoDir string, modifiedFiles []string) ([]valid.Project, error) {
	// TODO: rename struct roots
	var roots []valid.Project
	for _, root := range config.Projects {

		// Check if root's directory exists
		rootAbsDirectory := filepath.Join(absRepoDir, root.Dir)
		_, err := os.Stat(rootAbsDirectory)
		if err != nil {
			f.Logger.WarnContext(ctx, "unable to find directory for root", map[string]interface{}{
				"err": err,
				"dir": rootAbsDirectory,
			})
			continue
		}

		var whenModifiedRelToRepoRoot []string
		for _, wm := range root.Autoplan.WhenModified {
			wm = strings.TrimSpace(wm)
			// An exclusion uses a '!' at the beginning. If it's there, we need
			// to remove it, then add in the root path, then add it back.
			exclusion := false
			if wm != "" && wm[0] == '!' {
				wm = wm[1:]
				exclusion = true
			}

			// Prepend root dir to when modified patterns because the patterns
			// are relative to the root dirs but our list of modified files is
			// relative to the repo root.
			wmRelPath := filepath.Join(root.Dir, wm)
			if exclusion {
				wmRelPath = "!" + wmRelPath
			}
			whenModifiedRelToRepoRoot = append(whenModifiedRelToRepoRoot, wmRelPath)
		}
		pm, err := fileutils.NewPatternMatcher(whenModifiedRelToRepoRoot)
		if err != nil {
			return nil, errors.Wrapf(err, "matching modified files with patterns: %v", root.Autoplan.WhenModified)
		}

		// If any of the modified files matches the pattern then this root is
		// considered modified.
		for _, file := range modifiedFiles {
			match, err := pm.Matches(file)
			if err != nil {
				continue
			}
			if match {
				roots = append(roots, root)
				break
			}
		}
	}
	return roots, nil
}
