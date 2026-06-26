// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package i18n_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/i18n"
	. "github.com/runatlantis/atlantis/testing"
)

func TestNormalizeLanguageCode(t *testing.T) {
	Equals(t, "en", i18n.NormalizeLanguageCode(""))
	Equals(t, "es", i18n.NormalizeLanguageCode("es-MX"))
	Equals(t, "en", i18n.NormalizeLanguageCode(" EN "))
}

func TestValidateLanguage(t *testing.T) {
	Ok(t, i18n.ValidateLanguage("en"))
	Ok(t, i18n.ValidateLanguage("es-MX"))
	ErrEquals(t, `unsupported language "de": supported languages are en, es`, i18n.ValidateLanguage("de"))
}

func TestTranslator_BuiltInCatalog(t *testing.T) {
	translator, err := i18n.NewTranslator(i18n.TranslatorConfig{LanguageCode: "es"})
	Ok(t, err)
	Equals(t, "Aplicar", translator.CommandTitle("apply"))
	Equals(t, "Solicitud de extracción", translator.PullRequestLabel())
	Equals(t, "Solicitud de fusión", translator.MergeRequestLabel())
}

func TestTranslator_CustomCatalogOverrides(t *testing.T) {
	customCatalogPath := filepath.Join(t.TempDir(), "custom-language.yaml")
	err := os.WriteFile(customCatalogPath, []byte(`
pull_request_label: Pull Request (custom)
merge_request_label: Merge Request (custom)
command_titles:
  plan: Plan (custom)
`), 0o600)
	Ok(t, err)

	translator, err := i18n.NewTranslator(i18n.TranslatorConfig{
		LanguageCode: "de",
		CatalogPath:  customCatalogPath,
	})
	Ok(t, err)
	Equals(t, "Plan (custom)", translator.CommandTitle("plan"))
	Equals(t, "Apply", translator.CommandTitle("apply")) // fallback from built-in en
	Equals(t, "Pull Request (custom)", translator.PullRequestLabel())
	Equals(t, "Merge Request (custom)", translator.MergeRequestLabel())
}

func TestValidateCustomCatalog(t *testing.T) {
	Ok(t, i18n.ValidateCustomCatalog(""))

	tmpDir := t.TempDir()
	validPath := filepath.Join(tmpDir, "valid-language.yaml")
	err := os.WriteFile(validPath, []byte("command_titles:\n  apply: Anwenden\n"), 0o600)
	Ok(t, err)
	Ok(t, i18n.ValidateCustomCatalog(validPath))

	invalidPath := filepath.Join(tmpDir, "invalid-language.yaml")
	err = os.WriteFile(invalidPath, []byte(":\n"), 0o600)
	Ok(t, err)
	err = i18n.ValidateCustomCatalog(invalidPath)
	if err == nil {
		t.Fatalf("expected invalid custom catalog error")
	}

	unknownFieldPath := filepath.Join(tmpDir, "unknown-field-language.yaml")
	err = os.WriteFile(unknownFieldPath, []byte("unknown: value\n"), 0o600)
	Ok(t, err)
	err = i18n.ValidateCustomCatalog(unknownFieldPath)
	if err == nil {
		t.Fatalf("expected unknown custom catalog field error")
	}

	mergeKeyPath := filepath.Join(tmpDir, "merge-key-language.yaml")
	err = os.WriteFile(mergeKeyPath, []byte("? [foo]\n: bar\n<<: {}\n"), 0o600)
	Ok(t, err)
	err = i18n.ValidateCustomCatalog(mergeKeyPath)
	if err == nil {
		t.Fatalf("expected invalid custom catalog merge-key error")
	}
}
