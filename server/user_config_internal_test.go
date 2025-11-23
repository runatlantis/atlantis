package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserConfig_IsUserConfigFieldSpecified(t *testing.T) {
	tests := []struct {
		name       string
		userConfig UserConfig
		field      string
		want       bool
	}{
		{
			name:       "empty user config does not specify blank field",
			userConfig: UserConfig{},
			field:      "",
			want:       false,
		},
		{
			name:       "empty user config does not specify invalid blank field",
			userConfig: UserConfig{},
			field:      "notafield",
			want:       false,
		},
		{
			name:       "empty user config does not specify actual field",
			userConfig: UserConfig{},
			field:      "data-dir",
			want:       false,
		},
		{
			name: "user config with value set to zero value is not specified",
			userConfig: UserConfig{
				DataDir: "",
			},
			field: "data-dir",
			want:  false,
		},
		{
			name: "user config with value set to real value is specified",
			userConfig: UserConfig{
				DataDir: "foobar",
			},
			field: "data-dir",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.userConfig.isUserConfigFieldSpecified(tt.field)
			assert.Equalf(t, tt.want, got, "isUserConfigFieldSpecified()")
		})
	}
}
