// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package i18n

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v4"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	EnglishLanguage = "en"
	SpanishLanguage = "es"
	DefaultLanguage = EnglishLanguage
)

var supportedLanguages = []string{
	EnglishLanguage,
	SpanishLanguage,
}

//go:embed locales/*.yaml
var localeFS embed.FS

type catalog struct {
	CommandTitles     map[string]string `yaml:"command_titles"`
	PullRequestLabel  string            `yaml:"pull_request_label"`
	MergeRequestLabel string            `yaml:"merge_request_label"`
}

func (base catalog) merged(overrides catalog) catalog {
	merged := base
	if merged.CommandTitles == nil {
		merged.CommandTitles = make(map[string]string)
	}
	for commandName, title := range overrides.CommandTitles {
		if strings.TrimSpace(title) == "" {
			continue
		}
		merged.CommandTitles[normalizeCommandName(commandName)] = strings.TrimSpace(title)
	}
	if strings.TrimSpace(overrides.PullRequestLabel) != "" {
		merged.PullRequestLabel = strings.TrimSpace(overrides.PullRequestLabel)
	}
	if strings.TrimSpace(overrides.MergeRequestLabel) != "" {
		merged.MergeRequestLabel = strings.TrimSpace(overrides.MergeRequestLabel)
	}
	return merged
}

// SupportedLanguages returns a copy of all supported language codes.
func SupportedLanguages() []string {
	return slices.Clone(supportedLanguages)
}

// NormalizeLanguageCode canonicalizes the user language input.
func NormalizeLanguageCode(code string) string {
	normalized := strings.TrimSpace(strings.ToLower(code))
	if normalized == "" {
		return DefaultLanguage
	}
	if strings.Contains(normalized, "-") {
		parts := strings.SplitN(normalized, "-", 2)
		normalized = parts[0]
	}
	return normalized
}

// IsSupportedLanguage returns whether the language code is supported.
func IsSupportedLanguage(code string) bool {
	return slices.Contains(supportedLanguages, NormalizeLanguageCode(code))
}

// SupportedLanguagesDescription returns a stable human-readable language list.
func SupportedLanguagesDescription() string {
	return strings.Join(supportedLanguages, ", ")
}

// ValidateLanguage returns an error if the language is not supported.
func ValidateLanguage(code string) error {
	normalized := NormalizeLanguageCode(code)
	if IsSupportedLanguage(normalized) {
		return nil
	}
	return fmt.Errorf("unsupported language %q: supported languages are %s", normalized, SupportedLanguagesDescription())
}

// ValidateCustomCatalog validates a custom YAML catalog file.
func ValidateCustomCatalog(path string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if _, err := loadCatalogFromPath(path); err != nil {
		return fmt.Errorf("invalid language-config-file %q: %w", path, err)
	}
	return nil
}

// Translator contains localized strings for comment rendering.
type Translator struct {
	languageCode string
	catalog      catalog
}

// TranslatorConfig configures translation loading.
type TranslatorConfig struct {
	LanguageCode string
	CatalogPath  string
}

// NewTranslator creates a translator from a built-in language plus optional custom YAML overrides.
func NewTranslator(config TranslatorConfig) (*Translator, error) {
	normalized := NormalizeLanguageCode(config.LanguageCode)
	customPath := config.CatalogPath

	baseLanguage := normalized
	if err := ValidateLanguage(baseLanguage); err != nil {
		if strings.TrimSpace(customPath) == "" {
			return nil, err
		}
		baseLanguage = DefaultLanguage
	}

	baseCatalog, err := loadCatalogFromBuiltIn(baseLanguage)
	if err != nil {
		return nil, err
	}
	loadedCatalog := baseCatalog

	if strings.TrimSpace(customPath) != "" {
		overrides, loadErr := loadCatalogFromPath(customPath)
		if loadErr != nil {
			return nil, fmt.Errorf("loading language config file %q: %w", customPath, loadErr)
		}
		loadedCatalog = loadedCatalog.merged(overrides)
	}

	return &Translator{
		languageCode: normalized,
		catalog:      loadedCatalog,
	}, nil
}

// NewTranslatorOrDefault creates a translator or falls back to English.
func NewTranslatorOrDefault(config TranslatorConfig) *Translator {
	translator, err := NewTranslator(config)
	if err == nil {
		return translator
	}
	translator, _ = NewTranslator(TranslatorConfig{LanguageCode: DefaultLanguage})
	if translator != nil {
		return translator
	}
	return &Translator{
		languageCode: DefaultLanguage,
		catalog: catalog{
			CommandTitles: make(map[string]string),
		},
	}
}

// LanguageCode returns the normalized language code.
func (t *Translator) LanguageCode() string {
	return t.languageCode
}

// CommandTitle returns the display title for a command name.
func (t *Translator) CommandTitle(commandName string) string {
	normalized := normalizeCommandName(commandName)
	if title, ok := t.catalog.CommandTitles[normalized]; ok && strings.TrimSpace(title) != "" {
		return title
	}
	return fallbackCommandTitle(normalized)
}

// PullRequestLabel returns a localized pull request label.
func (t *Translator) PullRequestLabel() string {
	if strings.TrimSpace(t.catalog.PullRequestLabel) != "" {
		return t.catalog.PullRequestLabel
	}
	return "Pull Request"
}

// MergeRequestLabel returns a localized merge request label.
func (t *Translator) MergeRequestLabel() string {
	if strings.TrimSpace(t.catalog.MergeRequestLabel) != "" {
		return t.catalog.MergeRequestLabel
	}
	return "Merge Request"
}

func loadCatalogFromBuiltIn(languageCode string) (catalog, error) {
	path := filepath.ToSlash(filepath.Join("locales", languageCode+".yaml"))
	data, err := localeFS.ReadFile(path)
	if err != nil {
		return catalog{}, fmt.Errorf("reading built-in language catalog %q: %w", languageCode, err)
	}
	return parseCatalog(data)
}

func loadCatalogFromPath(path string) (catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return catalog{}, err
	}
	return parseCatalog(data)
}

func parseCatalog(data []byte) (loadedCatalog catalog, retErr error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			retErr = fmt.Errorf("parsing language catalog: %v", recovered)
		}
	}()

	var c catalog
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	var node yaml.Node
	if err := decoder.Decode(&node); err != nil {
		if errors.Is(err, io.EOF) {
			c.CommandTitles = make(map[string]string)
			return c, nil
		}
		return catalog{}, err
	}
	if err := node.Load(&c, yaml.WithKnownFields()); err != nil {
		return catalog{}, err
	}
	if c.CommandTitles == nil {
		c.CommandTitles = make(map[string]string)
	}
	return c, nil
}

func normalizeCommandName(commandName string) string {
	return strings.TrimSpace(strings.ToLower(commandName))
}

func fallbackCommandTitle(commandName string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(commandName, "_", " "))
}
