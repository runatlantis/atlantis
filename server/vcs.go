package server

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/runatlantis/atlantis/server/events/models"
)

type VCSFeature struct {
	Name            string
	Description     string
	SupportedVCSs   []models.VCSHostType
	UserConfigField string
}

type VCSFeatures []VCSFeature

type VCSSupportSummary struct {
	Err      error
	Warnings []string
}

// GetVCSFeatures canonical list of per-VCS features, and which VCS support them
func GetVCSFeatures() VCSFeatures {
	return []VCSFeature{
		{
			Name:        "CommentEmojiReaction",
			Description: "Adds an emoji onto a comment when Atlantis is processing it",
			SupportedVCSs: []models.VCSHostType{
				models.Github,
				models.Gitlab,
				models.AzureDevops,
			},
			UserConfigField: "emoji-reaction",
		},
		{
			Name:        "DiscardApprovalOnPlan",
			Description: "Discard approval if a new plan has been executed",
			SupportedVCSs: []models.VCSHostType{
				models.Github,
				models.Gitlab,
			},
			UserConfigField: "discard-approval-on-plan",
		},
		{
			Name:        "SingleFileDownload",
			Description: "Whether we can download a single file from the VCS",
			SupportedVCSs: []models.VCSHostType{
				models.Github,
				models.Gitlab,
				models.Gitea,
			},
		},
		{
			Name:        "DetailedPullIsMergeable",
			Description: "Whether PullIsMergeable returns a detailed reason as to why it's unmergeable",
			SupportedVCSs: []models.VCSHostType{
				models.Github,
				models.Gitlab,
			},
		},
	}
}

// isUserConfigFieldSpecified helper to determine whether a given on userConfig was specified
func isUserConfigFieldSpecified(userConfig UserConfig, field string) bool {
	if field == "" {
		return false
	}
	v := reflect.ValueOf(userConfig)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return false
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if sf.Tag.Get("mapstructure") == field {
			fv := v.Field(i)
			if !fv.IsValid() {
				return false
			}
			if fv.IsZero() {
				// zero value → treat as "not specified"
				return false
			}
			return true
		}
	}

	// no such tag at all
	return false
}

// IsSupportedBy does the specified vcs support this feature?
func (v VCSFeature) IsSupportedBy(vcsType models.VCSHostType) bool {
	return slices.Contains(v.SupportedVCSs, vcsType)
}

// Validate check to confirm that the configured VCSs support the VCS Features
func (v VCSFeatures) Validate(configuredVCSs []models.VCSHostType, userConfig UserConfig) VCSSupportSummary {

	for _, vcsFeature := range v {
		isFieldSpecified := isUserConfigFieldSpecified(userConfig, vcsFeature.UserConfigField)
		if isFieldSpecified {
			// Since we've specified a field, it means we want to use this feature.
			// If none of the configured VCSs support this feature, that's an error
			if !slices.ContainsFunc(configuredVCSs, vcsFeature.IsSupportedBy) {
				return VCSSupportSummary{
					Err: fmt.Errorf("no configured VCS supports feature %s, cannot specify field %q", vcsFeature.Name, vcsFeature.UserConfigField),
				}
			}
		}
	}
	// By this point at least one VCS supports every field that's specified
	// Additionally, we should warn if a user specifies multiple VCSs, but one of them won't implement the feature
	var warnings []string
	for _, vcsFeature := range v {
		isFieldSpecified := isUserConfigFieldSpecified(userConfig, vcsFeature.UserConfigField)
		for _, configuredVCS := range configuredVCSs {
			if vcsFeature.IsSupportedBy(configuredVCS) {
				continue
			}
			if isFieldSpecified {
				warnings = append(warnings, fmt.Sprintf("Specified field %q for feature %s, which is not supported on %s", vcsFeature.UserConfigField, vcsFeature.Name, configuredVCS))
				continue
			}
			// At this point, the VCS is not supported, but there is no flag attempting its use for this feature, so nothing to do.
		}
	}
	return VCSSupportSummary{
		Warnings: warnings,
	}
}
