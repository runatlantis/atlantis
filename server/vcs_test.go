package server_test

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestVCSFeaturesValidate(t *testing.T) {

	cases := []struct {
		description    string
		vcsFeatures    server.VCSFeatures
		configuredVCS  []models.VCSHostType
		userConfig     server.UserConfig
		expectedResult server.VCSSupportSummary
	}{
		{
			description: "No features",
			vcsFeatures: server.VCSFeatures{},
		},
		{
			description: "Only supported on github, which is not configured but no flag set",
			vcsFeatures: server.VCSFeatures{
				{
					Name: "foo",
					SupportedVCSs: []models.VCSHostType{
						models.Github,
					},
					UserConfigField: "enable-foo",
				},
			},
		},
		{
			description: "Only supported on github, which is not configured and flag set",
			vcsFeatures: server.VCSFeatures{
				{
					Name: "foo",
					SupportedVCSs: []models.VCSHostType{
						models.Github,
					},
					UserConfigField: "data-dir",
				},
			},
			userConfig: server.UserConfig{
				DataDir: "foobar",
			},
			expectedResult: server.VCSSupportSummary{
				Err: errors.New(`no configured VCS supports feature foo, cannot specify field "data-dir"`),
			},
		},
		{
			description: "Only supported on github, configured gitlab and github, and flag set",
			vcsFeatures: server.VCSFeatures{
				{
					Name: "foo",
					SupportedVCSs: []models.VCSHostType{
						models.Github,
					},
					UserConfigField: "data-dir",
				},
			},
			configuredVCS: []models.VCSHostType{
				models.Github,
				models.Gitlab,
			},
			userConfig: server.UserConfig{
				DataDir: "foobar",
			},
			expectedResult: server.VCSSupportSummary{
				Warnings: []string{
					`Specified field "data-dir" for feature foo, which is not supported on Gitlab`,
				},
			},
		},
		{
			description: "Supported on both github and gitlab and flag set",
			vcsFeatures: server.VCSFeatures{
				{
					Name: "foo",
					SupportedVCSs: []models.VCSHostType{
						models.Github,
						models.Gitlab,
					},
					UserConfigField: "data-dir",
				},
			},
			configuredVCS: []models.VCSHostType{
				models.Github,
				models.Gitlab,
			},
			userConfig: server.UserConfig{
				DataDir: "data-dir",
			},
		},
		{
			description: "Supported on both github and gitlab and no flag set",
			vcsFeatures: server.VCSFeatures{
				{
					Name: "foo",
					SupportedVCSs: []models.VCSHostType{
						models.Github,
						models.Gitlab,
					},
					UserConfigField: "data-dir",
				},
			},
			configuredVCS: []models.VCSHostType{
				models.Github,
				models.Gitlab,
			},
		},
	}
	for _, tt := range cases {
		t.Run(tt.description, func(t *testing.T) {
			actualResult := tt.vcsFeatures.Validate(tt.configuredVCS, tt.userConfig)
			Equals(t, tt.expectedResult, actualResult)
		})
	}
}
