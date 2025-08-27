package command

import (
	"reflect"
	"regexp"
	"strings"
)

// ProjectScopeTags defines the tags that are added to metrics.
// Note: Branch names are intentionally excluded to prevent high cardinality issues.
// Only stable, low-cardinality identifiers should be included here.
type ProjectScopeTags struct {
	BaseRepo              string
	PrNumber              string
	Project               string
	ProjectPath           string
	TerraformDistribution string
	TerraformVersion      string
	Workspace             string
}

// validateTagValue ensures that tag values don't contain branch names or other high-cardinality data
func validateTagValue(key, value string) string {
	// Normalize the value to prevent high cardinality
	normalized := strings.TrimSpace(value)

	// If the value is empty, return a default
	if normalized == "" {
		return "unknown"
	}

	// For project names, ensure they don't contain branch-like patterns
	if key == "project" {
		// Remove branch-like suffixes that match common branch naming patterns
		// This regex looks for patterns like -feature-branch, -main, -master, etc.
		// but only at the end of the string to avoid false positives
		branchPattern := regexp.MustCompile(`-(?:feature|hotfix|release|main|master|develop|staging|prod)(?:-[a-zA-Z0-9_-]*)?$`)
		normalized = branchPattern.ReplaceAllString(normalized, "")

		// Handle cases where the entire project name is just a branch name
		// Check if the remaining string is just a branch name
		branchOnlyPattern := regexp.MustCompile(`^(?:feature|hotfix|release|main|master|develop|staging|prod)(?:-[a-zA-Z0-9_-]*)?$`)
		if branchOnlyPattern.MatchString(normalized) {
			normalized = "default-project"
		}

		// If the result is empty after removing branch suffix, use a default
		if strings.TrimSpace(normalized) == "" {
			normalized = "default-project"
		}
	}

	return normalized
}

func (s ProjectScopeTags) Loadtags() map[string]string {
	tags := make(map[string]string)

	v := reflect.ValueOf(s)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldName := t.Field(i).Name
		fieldValue := v.Field(i).String()
		tagKey := toSnakeCase(fieldName)
		tagValue := validateTagValue(tagKey, fieldValue)
		tags[tagKey] = tagValue
	}

	return tags
}

func toSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}
