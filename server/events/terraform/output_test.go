package terraform

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
	. "github.com/runatlantis/atlantis/testing"

)

func TestNewOutput(t *testing.T) {
	// Create tmp folder to hold the mocked tf output files
	tmp, cleanup := TempDir(t)
	defer cleanup()

	_, err := NewOutput(tmp)
	Ok(t, err)
}

func TestList_FilesExists(t *testing.T) {
	// Create tmp folder to hold the mocked tf output files
	tmp, cleanup := TempDir(t)
	defer cleanup()

	client, err := NewOutput(tmp)
	Ok(t, err)

	tfOutputFileNames := []string{
		"20201121175848-init-default-home_test_.atlantis_repos_test_test_2_default",
		"20201121175849-plan-default-home_test_.atlantis_repos_test_test_2_default",
		"20201121175850-apply-default-home_test_.atlantis_repos_test_test_2_default",
	}

	var tfOutputs []TfOutput

	// Create the mocked files and parse the file name.
	for _, tfOutputFileName := range tfOutputFileNames {
		_, err := os.Create(filepath.Join(tmp, tfOutputFileName))
		Ok(t, err)

		tfOutput, err := client.ParseFileName(tfOutputFileName)
		Ok(t, err)

		tfOutputs = append(tfOutputs, tfOutput)
	}

	outputs, err := client.List()
	Ok(t, err)

	// Sort both slices with the same criteria
	sort.Slice(tfOutputs, func(i, j int) bool {
		return tfOutputs[i].CreatedAt.Before(tfOutputs[j].CreatedAt)
	})

	sort.Slice(outputs, func(i, j int) bool {
		return outputs[i].CreatedAt.Before(outputs[j].CreatedAt)
	})

	Equals(t, tfOutputs, outputs)
}
