package terraform_test

import (
	"context"
	"os"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/cmd"
	"github.com/runatlantis/atlantis/server/core/terraform"
)

func TestTerraformInstall(t *testing.T) {
	d := &terraform.TerraformDownloader{}
	RegisterMockTestingT(t)
	binDir := t.TempDir()

	v, _ := version.NewVersion("1.8.1")

	newPath, err := d.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("Binary not found at %s", newPath)
	}
}

func TestOpenTofuInstall(t *testing.T) {
	d := &terraform.TofuDownloader{}
	RegisterMockTestingT(t)
	binDir := t.TempDir()

	v, _ := version.NewVersion("1.8.0")

	newPath, err := d.Install(context.Background(), binDir, cmd.DefaultTFDownloadURL, v)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Errorf("Binary not found at %s", newPath)
	}
}
