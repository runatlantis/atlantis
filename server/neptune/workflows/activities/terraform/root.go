package terraform

import (
	"path/filepath"
	"strings"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/execute"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
)

// Root is the definition of a root
type Root struct {
	Name string

	// Path is the relative path from the repo
	Path         string
	TfVersion    string
	Apply        execute.Job
	Plan         PlanJob
	Trigger      Trigger
	Rerun        bool
	TrackedFiles []string
}

func (r Root) GetTrackedFilesRelativeToRepo() []string {
	var trackedFilesRelToRepoRoot []string
	for _, wm := range r.TrackedFiles {
		wm = strings.TrimSpace(wm)
		// An exclusion uses a '!' at the beginning. If it's there, we need
		// to remove it, then add in the project path, then add it back.
		exclusion := false
		if wm != "" && wm[0] == '!' {
			wm = wm[1:]
			exclusion = true
		}

		// Prepend project dir to when modified patterns because the patterns
		// are relative to the project dirs but our list of modified files is
		// relative to the repo root.
		wmRelPath := filepath.Join(r.Path, wm)
		if exclusion {
			wmRelPath = "!" + wmRelPath
		}
		trackedFilesRelToRepoRoot = append(trackedFilesRelToRepoRoot, wmRelPath)
	}
	return trackedFilesRelToRepoRoot
}

func (r Root) WithPlanApprovalOverride(a PlanApproval) Root {
	r.Plan.Approval = a
	return r
}

type Trigger string

const (
	MergeTrigger  Trigger = "merge"
	ManualTrigger Trigger = "manual"
)

// LocalRoot is a root that exists locally on disk
type LocalRoot struct {
	Root Root
	// Path on disk
	Path string
	Repo github.Repo
}

func (r *LocalRoot) RelativePathFromRepo() string {
	return r.Root.Path
}

func BuildLocalRoot(root Root, repo github.Repo, path string) *LocalRoot {
	return &LocalRoot{
		Root: root,
		Repo: repo,
		Path: path,
	}
}
