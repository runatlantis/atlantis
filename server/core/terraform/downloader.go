package terraform

import (
	"context"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"github.com/opentofu/tofudl"
)

//go:generate pegomock generate --package mocks -o mocks/mock_downloader.go Downloader

// Downloader is for downloading terraform versions.
type Downloader interface {
	Install(ctx context.Context, dir string, downloadURL string, v *version.Version) (string, error)
}

type TofuDownloader struct{}

func (d *TofuDownloader) Install(ctx context.Context, dir string, _downloadURL string, v *version.Version) (string, error) {
	// Initialize the downloader:
	dl, err := tofudl.New()
	if err != nil {
		return "", err
	}

	binary, err := dl.Download(ctx, tofudl.DownloadOptVersion(tofudl.Version(v.String())))
	if err != nil {
		return "", err
	}

	// Write out the tofu binary to the disk:
	file := filepath.Join(dir, "tofu"+v.String())
	if err := os.WriteFile(file, binary, 0755); /* #nosec G306 */ err != nil {
		return "", err
	}

	return file, nil
}

type TerraformDownloader struct{}

func (d *TerraformDownloader) Install(ctx context.Context, dir string, downloadURL string, v *version.Version) (string, error) {
	installer := install.NewInstaller()
	execPath, err := installer.Install(ctx, []src.Installable{
		&releases.ExactVersion{
			Product:    product.Terraform,
			Version:    v,
			InstallDir: dir,
			ApiBaseURL: downloadURL,
		},
	})
	if err != nil {
		return "", err
	}

	// hc-install installs terraform binary as just "terraform".
	// We need to rename it to terraform{version} to be consistent with current naming convention.
	newPath := filepath.Join(dir, "terraform"+v.String())
	if err := os.Rename(execPath, newPath); err != nil {
		return "", err
	}

	return newPath, nil
}
