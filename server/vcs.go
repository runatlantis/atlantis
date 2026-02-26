package server

import (
	"fmt"
	"slices"

	"github.com/runatlantis/atlantis/server/events/models"
)

type VCSFeature struct {
	Name          string
	Description   string
	SupportedVCSs []models.VCSHostType
	// Which field, if any, in UserConfig enables or configures this feature
	UserConfigField string
	// Per VCS notes, for documentation. Supports markdown
	Notes map[models.VCSHostType]string
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
			Notes: map[models.VCSHostType]string{
				models.Github:      "[Supported Emojis](https://docs.github.com/en/rest/reactions/reactions?apiVersion=2022-11-28#about-reactions)",
				models.Gitlab:      "[Supported Emojis](https://gitlab.com/gitlab-org/gitlab/-/blob/master/fixtures/emojis/digests.json)",
				models.AzureDevops: "[Supported Emojis](https://learn.microsoft.com/en-us/azure/devops/project/wiki/markdown-guidance?view=azure-devops#emoji)",
			},
		},
		{
			Name:        "DiscardApprovalOnPlan",
			Description: "Discard approval if a new plan has been executed",
			SupportedVCSs: []models.VCSHostType{
				models.Github,
				models.Gitlab,
			},
			UserConfigField: "discard-approval-on-plan",
			Notes: map[models.VCSHostType]string{
				models.Gitlab: "A group or Project token is required for this feature, see [reset-approvals-of-a-merge-request](https://docs.gitlab.com/api/merge_request_approvals/#reset-approvals-of-a-merge-request)",
			},
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
		{
			Name:            "HidePreviousPlanComments",
			Description:     "Hide previous plan comments to declutter PRs",
			UserConfigField: "hide-prev-plan-comments",
			SupportedVCSs: []models.VCSHostType{
				models.Github,
				models.BitbucketCloud,
				models.Gitlab,
				models.Gitea,
			},
			Notes: map[models.VCSHostType]string{
				models.BitbucketCloud: "comments are deleted rather than hidden as Bitbucket does not support hiding comments.",
				models.Github:         "Ensure the `--gh-user` is set appropriately or comments will not be hidden. When using the GitHub App, you need to set `--gh-app-slug` to enable this feature.",
			},
		},
	}
}

// IsSupportedBy does the specified vcs support this feature?
func (v VCSFeature) IsSupportedBy(vcsType models.VCSHostType) bool {
	return slices.Contains(v.SupportedVCSs, vcsType)
}

// Validate check to confirm that the configured VCSs support the VCS Features
func (v VCSFeatures) Validate(configuredVCSs []models.VCSHostType, userConfig UserConfig) VCSSupportSummary {

	for _, vcsFeature := range v {
		isFieldSpecified := userConfig.isUserConfigFieldSpecified(vcsFeature.UserConfigField)
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
		isFieldSpecified := userConfig.isUserConfigFieldSpecified(vcsFeature.UserConfigField)
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
