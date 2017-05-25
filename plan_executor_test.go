package main

import (
	"reflect"
	"testing"
)

func TestDetermineExecPaths(t *testing.T) {
	runTest(t, "should handle empty plan paths", "/repoPath", []string{}, []ExecutionPath{})
	runTest(t, "should extract path and append to repoPath", "/repoPath", []string{"a/b/c.ext"}, []ExecutionPath{{"/repoPath/a/b", "a/b"}})
	runTest(t, "should return repoPath if at root", "/repoPath", []string{"a.ext"}, []ExecutionPath{{"/repoPath", "."}})
	runTest(t, "should handle repoPath with trailing slash", "/repoPath/", []string{"a.ext"}, []ExecutionPath{{"/repoPath", "."}})
	runTest(t, "should set plan dir one level up from env/ directories", "/repoPath/", []string{"env/a.ext"}, []ExecutionPath{{"/repoPath", "."}})
	runTest(t, "should set plan dir one level up from env/ directories and deduplicate plans.", "/repoPath/", []string{"env/a.ext", "b.ext"}, []ExecutionPath{{"/repoPath", "."}})
	runTest(t, "should de-depluciate", "/repoPath/", []string{"a/b/c.ext", "a/b/d.ext"}, []ExecutionPath{{"/repoPath/a/b", "a/b"}})
}

func runTest(t *testing.T, testDescrip string, repoPath string, filesChanged []string, expected []ExecutionPath) {
	p := PlanExecutor{}
	plans := p.DetermineExecPaths(repoPath, filesChanged)
	if !reflect.DeepEqual(expected, plans) {
		t.Errorf("%s: expected %v, got %v", testDescrip, expected, plans)
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	runTestGetOutputFilename(t, "should handle empty plan path", "/repoPath", NewExecutionPath("", ""), "env", ".tfplan.env")
	runTestGetOutputFilename(t, "should handle empty environment", "/repoPath", NewExecutionPath("", ""), "", ".tfplan")
	runTestGetOutputFilename(t, "should prepend underscore on relative paths", "/repoPath", NewExecutionPath("", "a/b"), "", "_a_b.tfplan")
	runTestGetOutputFilename(t, "should work with relative path and environment", "/repoPath", NewExecutionPath("", "a/b"), "env", "_a_b.tfplan.env")
	runTestGetOutputFilename(t, "should exec path at root", "/a/b", NewExecutionPath("", "."), "env", ".tfplan.env")
}

func runTestGetOutputFilename(t *testing.T, testDescrip string, repoPath string, tfPlanPath ExecutionPath, tfEnvName string, expected string) {
	p := PlanExecutor{}
	outputFileName := p.GenerateOutputFilename(repoPath, tfPlanPath, tfEnvName)
	if !reflect.DeepEqual(expected, outputFileName) {
		t.Errorf("%s: expected %v, got %v", testDescrip, expected, outputFileName)
	}
}
