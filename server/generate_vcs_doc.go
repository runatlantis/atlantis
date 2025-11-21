//go:build ignore

//go:generate go run ./generate_vcs_doc.go

// This script updates documentation around supported VCS features
// It is meant to be run via `make go-generate` or directly via `go generate server/generate_vcs_doc.go`

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/models"
)

// Output path relative to this script
var filename = filepath.Join("..", "runatlantis.io", "docs", "per-vcs-features.md")

var allVCSs = []models.VCSHostType{
	models.Github,
	models.Gitlab,
	models.BitbucketCloud,
	models.BitbucketServer,
	models.AzureDevops,
	models.Gitea,
}

func getMarkdown() string {
	features := server.GetVCSFeatures()

	// Markdown table header
	out := "# VCS Features\n\n"
	out += "Below are the available features and which VCS providers support them.\n\n"
	out += "Some are configurable by the user with flags, others are handled internally. Some are unimplemented because of inherent deficiencies with the VCS, whereas most are just due to lack of developer support.\n\n"

	for _, f := range features {
		out += fmt.Sprintf("### %s\n", f.Name)
		if f.UserConfigField != "" {
			out += fmt.Sprintf("[`--%s`](/docs/server-configuration.html#%s)\n\n", f.UserConfigField, f.UserConfigField)
		}
		out += fmt.Sprintf("%s\n\n", f.Description)

		out += "| *VCS* | *Supported* |\n"
		out += "|---|---------|\n"

		for _, vcs := range allVCSs {
			if f.IsSupportedBy(vcs) {
				out += fmt.Sprintf("| %s | ✔ |\n", vcs.String())
			} else {
				out += fmt.Sprintf("| %s | ✘ |\n", vcs.String())
			}
		}
		out += "\n"
	}

	return out
}

func main() {

	content := getMarkdown()

	err := os.WriteFile(filename, []byte(content), 0o644)
	if err != nil {
		log.Fatal(err)
	}
}
