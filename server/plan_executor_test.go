package server

import (
	"github.com/hootsuite/atlantis/models"
	. "github.com/hootsuite/atlantis/testing_util"
	"testing"
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

func TestGenerateOutputFilename(t *testing.T) {
	runTestGetOutputFilename(t, "should handle root", ".", "env", ".tfplan.env")
	runTestGetOutputFilename(t, "should handle empty environment", ".", "", ".tfplan")
	runTestGetOutputFilename(t, "should prepend underscore on relative paths", "a/b", "", "_a_b.tfplan")
	runTestGetOutputFilename(t, "should prepend underscore on relative paths and env", "a/b", "env", "_a_b.tfplan.env")
}

func runTestGetOutputFilename(t *testing.T, testDescrip string, path string, env string, expected string) {
	t.Log(testDescrip)
	outputFileName := p.GenerateOutputFilename(models.NewProject("owner/repo", path), env)
	Equals(t, expected, outputFileName)
}
