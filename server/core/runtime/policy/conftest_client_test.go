// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime/cache/mocks"
	models_mocks "github.com/runatlantis/atlantis/server/core/runtime/models/mocks"
	conftest_mocks "github.com/runatlantis/atlantis/server/core/runtime/policy/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestConfTestVersionDownloader(t *testing.T) {

	version, _ := version.NewVersion("0.25.0")
	destPath := "some/path"
	platform := getPlatform()
	fullURL := fmt.Sprintf("https://github.com/open-policy-agent/conftest/releases/download/v0.25.0/conftest_0.25.0_%s.tar.gz?checksum=file:https://github.com/open-policy-agent/conftest/releases/download/v0.25.0/checksums.txt", platform)

	RegisterMockTestingT(t)

	mockDownloader := conftest_mocks.NewMockDownloader()

	subject := ConfTestVersionDownloader{
		downloader: mockDownloader,
	}

	t.Run("success", func(t *testing.T) {

		binPath, err := subject.downloadConfTestVersion(version, destPath)

		mockDownloader.VerifyWasCalledOnce().GetAny(Eq(destPath), Eq(fullURL))

		Ok(t, err)

		Assert(t, binPath.Resolve() == filepath.Join(destPath, "conftest"), "expected binpath")
	})

	t.Run("error", func(t *testing.T) {

		When(mockDownloader.GetAny(Eq(destPath), Eq(fullURL))).ThenReturn(errors.New("err"))
		_, err := subject.downloadConfTestVersion(version, destPath)

		Assert(t, err != nil, "err is expected")
	})
}

func TestEnsureExecutorVersion(t *testing.T) {

	defaultVersion, _ := version.NewVersion("1.0")
	expectedPath := "some/path"

	RegisterMockTestingT(t)

	mockCache := mocks.NewMockExecutionVersionCache()
	mockExec := models_mocks.NewMockExec()
	log := logging.NewNoopLogger(t)

	t.Run("no specified version or default version without conftest command", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache: mockCache,
			Exec:         mockExec,
		}

		When(mockExec.LookPath(Any[string]())).ThenReturn("", errors.New("not found"))
		_, err := subject.EnsureExecutorVersion(log, nil)

		Assert(t, err != nil, "expected error finding version")
	})

	t.Run("no specified version or default version with conftest command", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache: mockCache,
			Exec:         mockExec,
		}
		When(mockExec.LookPath(Any[string]())).ThenReturn(expectedPath, nil)
		path, err := subject.EnsureExecutorVersion(log, nil)
		Ok(t, err)
		Assert(t, path == expectedPath, "path is expected")
	})

	t.Run("use default version", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache:           mockCache,
			DefaultConftestVersion: defaultVersion,
		}

		When(mockCache.Get(defaultVersion)).ThenReturn(expectedPath, nil)

		path, err := subject.EnsureExecutorVersion(log, nil)

		Ok(t, err)

		Assert(t, path == expectedPath, "path is expected")
	})

	t.Run("use specified version", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache:           mockCache,
			DefaultConftestVersion: defaultVersion,
		}

		versionInput, _ := version.NewVersion("2.0")

		When(mockCache.Get(versionInput)).ThenReturn(expectedPath, nil)

		path, err := subject.EnsureExecutorVersion(log, versionInput)

		Ok(t, err)

		Assert(t, path == expectedPath, "path is expected")
	})

	t.Run("cache error", func(t *testing.T) {
		subject := &ConfTestExecutorWorkflow{
			VersionCache:           mockCache,
			DefaultConftestVersion: defaultVersion,
		}

		versionInput, _ := version.NewVersion("2.0")

		When(mockCache.Get(versionInput)).ThenReturn(expectedPath, errors.New("some err"))

		_, err := subject.EnsureExecutorVersion(log, versionInput)

		Assert(t, err != nil, "path is expected")
	})
}

func TestRun(t *testing.T) {

	RegisterMockTestingT(t)
	mockResolver := conftest_mocks.NewMockSourceResolver()
	mockExec := models_mocks.NewMockExec()

	subject := &ConfTestExecutorWorkflow{
		SourceResolver: mockResolver,
		Exec:           mockExec,
	}
	log := logging.NewNoopLogger(t)

	policySetName1 := "policy1"
	policySetPath1 := "/some/path"
	localPolicySetPath1 := "/tmp/some/path"

	policySetName2 := "policy2"
	policySetPath2 := "/some/path2"
	localPolicySetPath2 := "/tmp/some/path2"
	executablePath := "/usr/bin/conftest"
	envs := map[string]string{
		"key": "val",
	}
	workdir := t.TempDir()

	policySet1 := valid.PolicySet{
		Source:          valid.LocalPolicySet,
		Path:            policySetPath1,
		Name:            policySetName1,
		PolicyItemRegex: valid.DefaultPolicyItemRegex,
	}

	policySet2 := valid.PolicySet{
		Source:          valid.LocalPolicySet,
		Path:            policySetPath2,
		Name:            policySetName2,
		PolicyItemRegex: valid.DefaultPolicyItemRegex,
	}

	ctx := command.ProjectContext{
		PolicySets: valid.PolicySets{
			PolicySets:      []valid.PolicySet{policySet1, policySet2},
			PolicyItemRegex: valid.DefaultPolicyItemRegex,
		},
		ProjectName: "testproj",
		Workspace:   "default",
		Log:         log,
	}

	t.Run("success", func(t *testing.T) {
		var extraArgs []string

		expectedOutput := "Success"
		h := models.HashPolicyItem("Success")
		expectedResult := fmt.Sprintf(`[{"PolicySetName":"policy1","PolicyOutput":"Success","Passed":true,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"},{"PolicySetName":"policy2","PolicyOutput":"Success","Passed":true,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"}]`, h, h)

		expectedArgsPolicy1 := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color"}
		expectedArgsPolicy2 := []string{executablePath, "test", "-p", localPolicySetPath2, filepath.Join(workdir, "testproj-default.json"), "--no-color"}

		When(mockResolver.Resolve(policySet1)).ThenReturn(localPolicySetPath1, nil)
		When(mockResolver.Resolve(policySet2)).ThenReturn(localPolicySetPath2, nil)

		When(mockExec.CombinedOutput(expectedArgsPolicy1, envs, workdir)).ThenReturn(expectedOutput, nil)
		When(mockExec.CombinedOutput(expectedArgsPolicy2, envs, workdir)).ThenReturn(expectedOutput, nil)

		result, err := subject.Run(ctx, executablePath, envs, workdir, extraArgs)

		Ok(t, errors.Unwrap(err))

		Assert(t, result == expectedResult, "result is expected")

	})

	t.Run("success extra args", func(t *testing.T) {
		extraArgs := []string{"--all-namespaces"}

		expectedOutput := "Success"
		h := models.HashPolicyItem("Success")
		expectedResult := fmt.Sprintf(`[{"PolicySetName":"policy1","PolicyOutput":"Success","Passed":true,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"},{"PolicySetName":"policy2","PolicyOutput":"Success","Passed":true,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"}]`, h, h)

		expectedArgsPolicy1 := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color", "--all-namespaces"}
		expectedArgsPolicy2 := []string{executablePath, "test", "-p", localPolicySetPath2, filepath.Join(workdir, "testproj-default.json"), "--no-color", "--all-namespaces"}

		When(mockResolver.Resolve(policySet1)).ThenReturn(localPolicySetPath1, nil)
		When(mockResolver.Resolve(policySet2)).ThenReturn(localPolicySetPath2, nil)

		When(mockExec.CombinedOutput(expectedArgsPolicy1, envs, workdir)).ThenReturn(expectedOutput, nil)
		When(mockExec.CombinedOutput(expectedArgsPolicy2, envs, workdir)).ThenReturn(expectedOutput, nil)

		result, err := subject.Run(ctx, executablePath, envs, workdir, extraArgs)

		Ok(t, errors.Unwrap(err))

		Assert(t, result == expectedResult, "result is expected")

	})

	t.Run("error resolving one policy source", func(t *testing.T) {
		var extraArgs []string

		expectedOutput := "Success"
		expectedResult := fmt.Sprintf(`[{"PolicySetName":"policy1","PolicyOutput":"Success","Passed":true,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"}]`, models.HashPolicyItem("Success"))

		expectedArgsPolicy1 := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color"}
		expectedArgsPolicy2 := []string{executablePath, "test", "-p", localPolicySetPath2, filepath.Join(workdir, "testproj-default.json"), "--no-color"}

		When(mockResolver.Resolve(policySet1)).ThenReturn(localPolicySetPath1, nil)
		When(mockResolver.Resolve(policySet2)).ThenReturn("", errors.New("err"))

		When(mockExec.CombinedOutput(expectedArgsPolicy1, envs, workdir)).ThenReturn(expectedOutput, nil)
		When(mockExec.CombinedOutput(expectedArgsPolicy2, envs, workdir)).ThenReturn(expectedOutput, nil)

		result, err := subject.Run(ctx, executablePath, envs, workdir, extraArgs)

		Ok(t, errors.Unwrap(err))

		Assert(t, result == expectedResult, "result is expected")

	})

	t.Run("error resolving both policy sources", func(t *testing.T) {
		var extraArgs []string

		expectedResult := ""
		expectedArgsPolicy1 := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color"}

		When(mockResolver.Resolve(policySet1)).ThenReturn("", errors.New("err"))
		When(mockResolver.Resolve(policySet2)).ThenReturn("", errors.New("err"))

		When(mockExec.CombinedOutput(expectedArgsPolicy1, envs, workdir)).ThenReturn(expectedResult, nil)

		result, err := subject.Run(ctx, executablePath, envs, workdir, extraArgs)

		Ok(t, err)

		Assert(t, result == "", "result is expected")

	})

	t.Run("error running one cmd", func(t *testing.T) {
		var extraArgs []string

		expectedOutputPolicy1 := fmt.Sprintf("FAIL - %s - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions", filepath.Join(workdir, "testproj-default.json"))
		expectedOutputPolicy2 := "Success"
		sanitizedPolicy1 := "FAIL - <redacted plan file> - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions"
		hFail := models.HashPolicyItem(sanitizedPolicy1)
		hSuccess := models.HashPolicyItem("Success")
		expectedResult := fmt.Sprintf(`[{"PolicySetName":"policy1","PolicyOutput":"FAIL - <redacted plan file> - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions","Passed":false,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"},{"PolicySetName":"policy2","PolicyOutput":"Success","Passed":true,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"}]`, hFail, hSuccess)

		expectedArgsPolicy1 := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color"}
		expectedArgsPolicy2 := []string{executablePath, "test", "-p", localPolicySetPath2, filepath.Join(workdir, "testproj-default.json"), "--no-color"}

		When(mockResolver.Resolve(policySet1)).ThenReturn(localPolicySetPath1, nil)
		When(mockResolver.Resolve(policySet2)).ThenReturn(localPolicySetPath2, nil)

		When(mockExec.CombinedOutput(expectedArgsPolicy1, envs, workdir)).ThenReturn(expectedOutputPolicy1, errors.New("exit status code 1"))
		When(mockExec.CombinedOutput(expectedArgsPolicy2, envs, workdir)).ThenReturn(expectedOutputPolicy2, nil)

		result, err := subject.Run(ctx, executablePath, envs, workdir, extraArgs)

		Equals(t, result, expectedResult)
		Assert(t, err != nil, "error is expected")

	})

	t.Run("error running both cmds", func(t *testing.T) {
		var extraArgs []string

		expectedOutput := fmt.Sprintf("FAIL - %s - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions", filepath.Join(workdir, "testproj-default.json"))
		sanitizedOutput := "FAIL - <redacted plan file> - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions"
		hBoth := models.HashPolicyItem(sanitizedOutput)
		expectedResult := fmt.Sprintf(`[{"PolicySetName":"policy1","PolicyOutput":"FAIL - <redacted plan file> - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions","Passed":false,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"},{"PolicySetName":"policy2","PolicyOutput":"FAIL - <redacted plan file> - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions","Passed":false,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"}]`, hBoth, hBoth)

		expectedArgsPolicy1 := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color"}
		expectedArgsPolicy2 := []string{executablePath, "test", "-p", localPolicySetPath2, filepath.Join(workdir, "testproj-default.json"), "--no-color"}

		When(mockResolver.Resolve(policySet1)).ThenReturn(localPolicySetPath1, nil)
		When(mockResolver.Resolve(policySet2)).ThenReturn(localPolicySetPath2, nil)

		When(mockExec.CombinedOutput(expectedArgsPolicy1, envs, workdir)).ThenReturn(expectedOutput, errors.New("exit status code 1"))
		When(mockExec.CombinedOutput(expectedArgsPolicy2, envs, workdir)).ThenReturn(expectedOutput, errors.New("exit status code 1"))

		result, err := subject.Run(ctx, executablePath, envs, workdir, extraArgs)

		Equals(t, result, expectedResult)
		Assert(t, err != nil, "error is expected")

	})

	t.Run("parse error should fail policy", func(t *testing.T) {
		var extraArgs []string

		parseErrorOutput := "Error: running test: load: loading policies: load: 2 errors occurred during loading:"
		hParseErr := models.HashPolicyItem(parseErrorOutput)
		expectedResult := fmt.Sprintf(`[{"PolicySetName":"policy1","PolicyOutput":"%s","Passed":false,"ReqApprovalCount":0,"Approvals":null,"Hashes":["%s"],"PolicyItemRegex":"(?s).+"}]`, parseErrorOutput, hParseErr)

		expectedArgsPolicy := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color"}

		When(mockResolver.Resolve(policySet1)).ThenReturn(localPolicySetPath1, nil)
		When(mockExec.CombinedOutput(expectedArgsPolicy, envs, workdir)).ThenReturn(parseErrorOutput, errors.New("exit status code 1"))

		ctxSinglePolicy := command.ProjectContext{
			PolicySets: valid.PolicySets{
				PolicySets:      []valid.PolicySet{policySet1},
				PolicyItemRegex: valid.DefaultPolicyItemRegex,
			},
			ProjectName: "testproj",
			Workspace:   "default",
			Log:         log,
		}

		result, err := subject.Run(ctxSinglePolicy, executablePath, envs, workdir, extraArgs)

		Equals(t, result, expectedResult)
		Assert(t, err != nil, "error is expected")

	})
}

func TestConftestTestCommandArgs_build(t *testing.T) {
	// Arguments are handed to conftest as a vector rather than spliced into a
	// shell command line, so anything that used to rely on the shell splitting
	// a value on whitespace has to be split here instead. Two configurations
	// depend on that: a policy set `path` with extra flags appended (there is
	// no list form for `path`, so this is the only way to express it), and the
	// `extra_args` form used by the policy checking documentation.
	cases := []struct {
		description string
		policyPaths []string
		extraArgs   []string
		expArgs     []string
	}{
		{
			description: "plain path is untouched",
			policyPaths: []string{"/home/atlantis/conftest_policies"},
			expArgs: []string{"conftest", "test", "-p", "/home/atlantis/conftest_policies",
				"/tmp/default.json", "--no-color"},
		},
		{
			description: "path carrying an extra flag is split",
			policyPaths: []string{"/atlantis-data/terraform-policy/policy --namespace=standards"},
			expArgs: []string{"conftest", "test", "-p", "/atlantis-data/terraform-policy/policy",
				"--namespace=standards", "/tmp/default.json", "--no-color"},
		},
		{
			// This is the form used in policy-checking.md.
			description: "extra_args entry carrying a flag and its value is split",
			policyPaths: []string{"/home/atlantis/conftest_policies"},
			extraArgs:   []string{"-p /home/atlantis/conftest_policies/", "--all-namespaces"},
			expArgs: []string{"conftest", "test", "-p", "/home/atlantis/conftest_policies",
				"/tmp/default.json", "--no-color", "-p", "/home/atlantis/conftest_policies/",
				"--all-namespaces"},
		},
		{
			description: "quoting expresses a genuine space in a path",
			policyPaths: []string{`"/home/atlantis/my policies"`},
			expArgs: []string{"conftest", "test", "-p", "/home/atlantis/my policies",
				"/tmp/default.json", "--no-color"},
		},
		{
			// Splitting is not shell interpretation: metacharacters become
			// ordinary arguments and are never executed.
			description: "shell metacharacters become literal arguments",
			policyPaths: []string{"/policies; id"},
			expArgs: []string{"conftest", "test", "-p", "/policies;", "id",
				"/tmp/default.json", "--no-color"},
		},
		{
			description: "unbalanced quote falls back to the value as a single argument",
			policyPaths: []string{"/home/o'brien/policies"},
			expArgs: []string{"conftest", "test", "-p", "/home/o'brien/policies",
				"/tmp/default.json", "--no-color"},
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			policyArgs := make([]Arg, 0, len(c.policyPaths))
			for _, p := range c.policyPaths {
				policyArgs = append(policyArgs, NewPolicyArg(p))
			}
			args := ConftestTestCommandArgs{
				PolicyArgs: policyArgs,
				ExtraArgs:  c.extraArgs,
				InputFile:  "/tmp/default.json",
				Command:    "conftest",
			}

			got, err := args.build()

			Ok(t, err)
			Equals(t, c.expArgs, got)
		})
	}
}

func TestConftestTestCommandArgs_buildExpandsEnvVars(t *testing.T) {
	// The previous shell-based command line expanded environment variables in
	// administrator-configured values. Expansion is done here now, without a
	// shell.
	args := ConftestTestCommandArgs{
		PolicyArgs: []Arg{NewPolicyArg("$POLICY_DIR --namespace=$POLICY_NS")},
		ExtraArgs:  []string{"--data $POLICY_DATA"},
		// The input file is named after the project and workspace, which come
		// from a pull request, so it is never expanded.
		InputFile: "/tmp/$NOT_EXPANDED.json",
		Command:   "conftest",
		Env:       map[string]string{"POLICY_DIR": "/policies", "POLICY_NS": "standards", "POLICY_DATA": "/data"},
	}

	got, err := args.build()

	Ok(t, err)
	Equals(t, []string{"conftest", "test", "-p", "/policies", "--namespace=standards",
		"/tmp/$NOT_EXPANDED.json", "--no-color", "--data", "/data"}, got)
}

func TestSplitConfiguredArgPreservesExpansionIntent(t *testing.T) {
	t.Setenv("CONFTEST_PROCESS_ONLY", "process-value")
	t.Setenv("CONFTEST_PRECEDENCE", "process-value")

	env := map[string]string{
		"CONFTEST_VALUE":        "configured-value",
		"CONFTEST_WITH_SPACE":   "configured value",
		"CONFTEST_COMPLEX":      `space ; $(id) $CONFTEST_VALUE`,
		"CONFTEST_PREFIX_VALUE": "joined-secret",
		"CONFTEST_PRECEDENCE":   "workflow-value",
	}
	tests := []struct {
		description string
		value       string
		expected    []string
	}{
		{
			description: "unquoted reference expands",
			value:       `$CONFTEST_VALUE`,
			expected:    []string{"configured-value"},
		},
		{
			description: "double quoted reference expands without splitting its value",
			value:       `"$CONFTEST_WITH_SPACE"`,
			expected:    []string{"configured value"},
		},
		{
			description: "unquoted reference does not split its expanded value",
			value:       `--data $CONFTEST_WITH_SPACE`,
			expected:    []string{"--data", "configured value"},
		},
		{
			description: "single quoted reference remains literal",
			value:       `'$CONFTEST_VALUE'`,
			expected:    []string{"$CONFTEST_VALUE"},
		},
		{
			description: "escaped unquoted reference remains literal",
			value:       `\$CONFTEST_VALUE`,
			expected:    []string{"$CONFTEST_VALUE"},
		},
		{
			description: "escaped double quoted reference remains literal",
			value:       `"\$CONFTEST_VALUE"`,
			expected:    []string{"$CONFTEST_VALUE"},
		},
		{
			description: "mixed quoting preserves each segment's expansion intent",
			value:       `prefix-'$CONFTEST_VALUE'-"$CONFTEST_VALUE"-\$CONFTEST_VALUE`,
			expected:    []string{"prefix-$CONFTEST_VALUE-configured-value-$CONFTEST_VALUE"},
		},
		{
			description: "quote boundary cannot synthesize an expandable variable name",
			value:       `'$CONFTEST_PREFIX_'VALUE`,
			expected:    []string{"$CONFTEST_PREFIX_VALUE"},
		},
		{
			description: "expanded value is not parsed or expanded again",
			value:       `$CONFTEST_COMPLEX`,
			expected:    []string{`space ; $(id) $CONFTEST_VALUE`},
		},
		{
			description: "process environment remains available",
			value:       `$CONFTEST_PROCESS_ONLY`,
			expected:    []string{"process-value"},
		},
		{
			description: "process environment matches child precedence",
			value:       `$CONFTEST_PRECEDENCE`,
			expected:    []string{"process-value"},
		},
		{
			description: "malformed quoting is passed through without expansion",
			value:       `'$CONFTEST_VALUE`,
			expected:    []string{"'$CONFTEST_VALUE"},
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			Equals(t, test.expected, splitConfiguredArg(test.value, env))
		})
	}
}

func TestConftestTestCommandArgs_buildProtectsBothConfiguredSources(t *testing.T) {
	args := ConftestTestCommandArgs{
		PolicyArgs: []Arg{NewPolicyArg(`'$CONFTEST_SECRET'`)},
		ExtraArgs:  []string{`--namespace=\$CONFTEST_SECRET`},
		InputFile:  "/tmp/$CONFTEST_SECRET.json",
		Command:    "conftest",
		Env:        map[string]string{"CONFTEST_SECRET": "sensitive-value"},
	}

	got, err := args.build()

	Ok(t, err)
	Equals(t, []string{"conftest", "test", "-p", "$CONFTEST_SECRET",
		"/tmp/$CONFTEST_SECRET.json", "--no-color", "--namespace=$CONFTEST_SECRET"}, got)
}
