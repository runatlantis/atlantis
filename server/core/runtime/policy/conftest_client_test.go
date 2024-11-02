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
		Source: valid.LocalPolicySet,
		Path:   policySetPath1,
		Name:   policySetName1,
	}

	policySet2 := valid.PolicySet{
		Source: valid.LocalPolicySet,
		Path:   policySetPath2,
		Name:   policySetName2,
	}

	ctx := command.ProjectContext{
		PolicySets: valid.PolicySets{
			PolicySets: []valid.PolicySet{
				policySet1,
				policySet2,
			},
		},
		ProjectName: "testproj",
		Workspace:   "default",
		Log:         log,
	}

	t.Run("success", func(t *testing.T) {
		var extraArgs []string

		expectedOutput := "Success"
		expectedResult := `[{"PolicySetName":"policy1","PolicyOutput":"Success","Passed":true,"ReqApprovals":0,"CurApprovals":0},{"PolicySetName":"policy2","PolicyOutput":"Success","Passed":true,"ReqApprovals":0,"CurApprovals":0}]`

		expectedArgsPolicy1 := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color"}
		expectedArgsPolicy2 := []string{executablePath, "test", "-p", localPolicySetPath2, filepath.Join(workdir, "testproj-default.json"), "--no-color"}

		When(mockResolver.Resolve(policySet1)).ThenReturn(localPolicySetPath1, nil)
		When(mockResolver.Resolve(policySet2)).ThenReturn(localPolicySetPath2, nil)

		When(mockExec.CombinedOutput(expectedArgsPolicy1, envs, workdir)).ThenReturn(expectedOutput, nil)
		When(mockExec.CombinedOutput(expectedArgsPolicy2, envs, workdir)).ThenReturn(expectedOutput, nil)

		result, err := subject.Run(ctx, executablePath, envs, workdir, extraArgs)

		fmt.Println(result)

		Ok(t, errors.Unwrap(err))

		Assert(t, result == expectedResult, "result is expected")

	})

	t.Run("success extra args", func(t *testing.T) {
		extraArgs := []string{"--all-namespaces"}

		expectedOutput := "Success"
		expectedResult := `[{"PolicySetName":"policy1","PolicyOutput":"","Passed":true,"ReqApprovals":0,"CurApprovals":0},{"PolicySetName":"policy2","PolicyOutput":"","Passed":true,"ReqApprovals":0,"CurApprovals":0}]`

		expectedArgsPolicy1 := []string{executablePath, "test", "-p", localPolicySetPath1, filepath.Join(workdir, "testproj-default.json"), "--no-color"}
		expectedArgsPolicy2 := []string{executablePath, "test", "-p", localPolicySetPath2, filepath.Join(workdir, "testproj-default.json"), "--no-color"}

		When(mockResolver.Resolve(policySet1)).ThenReturn(localPolicySetPath1, nil)
		When(mockResolver.Resolve(policySet2)).ThenReturn(localPolicySetPath2, nil)

		When(mockExec.CombinedOutput(expectedArgsPolicy1, envs, workdir)).ThenReturn(expectedOutput, nil)
		When(mockExec.CombinedOutput(expectedArgsPolicy2, envs, workdir)).ThenReturn(expectedOutput, nil)

		result, err := subject.Run(ctx, executablePath, envs, workdir, extraArgs)

		fmt.Println(result)

		Ok(t, errors.Unwrap(err))

		Assert(t, result == expectedResult, "result is expected")

	})

	t.Run("error resolving one policy source", func(t *testing.T) {
		var extraArgs []string

		expectedOutput := "Success"
		expectedResult := `[{"PolicySetName":"policy1","PolicyOutput":"Success","Passed":true,"ReqApprovals":0,"CurApprovals":0}]`

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
		expectedResult := `[{"PolicySetName":"policy1","PolicyOutput":"FAIL - <redacted plan file> - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions","Passed":false,"ReqApprovals":0,"CurApprovals":0},{"PolicySetName":"policy2","PolicyOutput":"Success","Passed":true,"ReqApprovals":0,"CurApprovals":0}]`

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
		expectedResult := `[{"PolicySetName":"policy1","PolicyOutput":"FAIL - <redacted plan file> - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions","Passed":false,"ReqApprovals":0,"CurApprovals":0},{"PolicySetName":"policy2","PolicyOutput":"FAIL - <redacted plan file> - failure\n1 tests, 0 passed, 0 warnings, 1 failure, 0 exceptions","Passed":false,"ReqApprovals":0,"CurApprovals":0}]`

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
}
