package server

import (
	"testing"

	. "github.com/hootsuite/atlantis/testing_util"
)

var p PlanExecutor

func TestModifiedProjects(t *testing.T) {
	runTest(t, "should handle no files modified", []string{}, []string{})
	runTest(t, "should handle files at root", []string{"root.tf"}, []string{"."})
	runTest(t, "should de-duplicate files at root", []string{"root1.tf", "root2.tf"}, []string{"."})
	runTest(t, "should handle sub directories", []string{"sub/dir/file.tf"}, []string{"sub/dir"})
	runTest(t, "should de-duplicate files in sub directories", []string{"sub/dir/file.tf", "sub/dir/file2.tf"}, []string{"sub/dir"})
	runTest(t, "should handle nested sub directories", []string{"root.tf", "sub/child.tf", "sub/sub/child.tf"}, []string{".", "sub", "sub/sub"})
	runTest(t, "should use parent of env/ dirs", []string{"env/env.tf"}, []string{"."})
	runTest(t, "should use parent of env/ dirs in sub dirs", []string{"sub/env/env.tf"}, []string{"sub"})
}

func runTest(t *testing.T, testDescrip string, filesChanged []string, expectedPaths []string) {
	projects := p.ModifiedProjects("owner/repo", filesChanged)
	for i, p := range projects {
		t.Log(testDescrip)
		Equals(t, expectedPaths[i], p.Path)
	}
}
