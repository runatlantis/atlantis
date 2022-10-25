package events

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/fs
var repos embed.FS

func Test_findModuleDependants(t *testing.T) {

	type args struct {
		files                    fs.FS
		autoplanModuleDependants string
	}
	a, err := fs.Sub(repos, "testdata/fs/repoA")
	assert.NoError(t, err)
	b, err := fs.Sub(repos, "testdata/fs/repoB")
	assert.NoError(t, err)

	tests := []struct {
		name    string
		args    args
		want    map[string][]string
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "repoA",
			args: args{
				files:                    a,
				autoplanModuleDependants: "**/init.tf",
			},
			want: map[string][]string{
				"modules/bar": {"baz", "qux/quxx"},
				"modules/foo": {"qux/quxx"},
			},
			wantErr: assert.NoError,
		},
		{
			name: "repoB",
			args: args{
				files:                    b,
				autoplanModuleDependants: "**/init.tf",
			},
			want: map[string][]string{
				"modules/bar": {"dev/quxx", "prod/quxx"},
				"modules/foo": {"dev/quxx", "prod/quxx"},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findModuleDependants(tt.args.files, tt.args.autoplanModuleDependants)
			if !tt.wantErr(t, err, fmt.Sprintf("findModuleDependants(%v, %v)", tt.args.files, tt.args.autoplanModuleDependants)) {
				return
			}
			for k, v := range tt.want {
				projects := got.DependentProjects(k)
				sort.Strings(projects)
				assert.Equalf(t, v, projects, "%v.DownstreamProjects(%v)", got, k)
			}
		})
	}
}
