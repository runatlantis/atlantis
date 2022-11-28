package events

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/moby/moby/pkg/fileutils"
)

type module struct {
	// path to the module
	path string
	// dependencies of this module
	dependencies map[string]bool
	// projects that depend on this module
	projects map[string]bool
}

func (m *module) String() string {
	if m == nil {
		return "nil"
	}
	return fmt.Sprintf("%+v", *m)
}

type ModuleProjects interface {
	// DependentProjects returns all projects that depend on the module at moduleDir
	DependentProjects(moduleDir string) []string
}

type moduleInfo map[string]*module

var _ ModuleProjects = moduleInfo{}

func (m moduleInfo) String() string {
	return fmt.Sprintf("%+v", map[string]*module(m))
}

func (m moduleInfo) DependentProjects(moduleDir string) (projectPaths []string) {
	if m == nil || m[moduleDir] == nil {
		return nil
	}
	for project := range m[moduleDir].projects {
		projectPaths = append(projectPaths, project)
	}
	return projectPaths
}

type tfFs struct {
	fs.FS
}

func (t tfFs) Open(name string) (tfconfig.File, error) {
	return t.FS.Open(name)
}

func (t tfFs) ReadFile(name string) ([]byte, error) {
	return fs.ReadFile(t.FS, name)
}

func (t tfFs) ReadDir(dirname string) ([]os.FileInfo, error) {
	ls, err := fs.ReadDir(t.FS, dirname)
	if err != nil {
		return nil, err
	}
	var infos []os.FileInfo
	for _, l := range ls {
		info, err := l.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get info for %s: %w", l.Name(), err)
		}
		infos = append(infos, info)
	}
	return infos, err
}

var _ tfconfig.FS = tfFs{}

func (m moduleInfo) load(files fs.FS, dir string, projects ...string) (_ *module, diags tfconfig.Diagnostics) {
	if _, set := m[dir]; !set {
		tfFiles := tfFs{files}
		var mod *tfconfig.Module
		mod, diags = tfconfig.LoadModuleFromFilesystem(tfFiles, dir)

		deps := make(map[string]bool)
		if mod != nil {
			for _, c := range mod.ModuleCalls {
				mPath := path.Join(dir, c.Source)
				if !tfconfig.IsModuleDirOnFilesystem(tfFiles, mPath) {
					continue
				}
				deps[mPath] = true
			}
		}

		m[dir] = &module{
			path:         dir,
			dependencies: deps,
			projects:     make(map[string]bool),
		}
	}
	// set projects on my dependencies
	for dep := range m[dir].dependencies {
		_, err := m.load(files, dep, projects...)
		if err != nil {
			diags = append(diags, err...)
		}
	}
	// add projects to the list of dependant projects
	for _, p := range projects {
		m[dir].projects[p] = true
	}
	return m[dir], diags
}

// FindModuleProjects returns a mapping of modules to projects that depend on them.
func FindModuleProjects(absRepoDir string, autoplanModuleDependants string) (ModuleProjects, error) {
	return findModuleDependants(os.DirFS(absRepoDir), autoplanModuleDependants)
}

func findModuleDependants(files fs.FS, autoplanModuleDependants string) (ModuleProjects, error) {
	if autoplanModuleDependants == "" {
		return moduleInfo{}, nil
	}
	// find all the projects matching autoplanModuleDependants
	filter, _ := fileutils.NewPatternMatcher(strings.Split(autoplanModuleDependants, ","))
	var projects []string
	err := fs.WalkDir(files, ".", func(rel string, info fs.DirEntry, err error) error {
		if match, _ := filter.Matches(rel); match {
			if projectDir := getProjectDirFromFs(files, rel); projectDir != "" {
				projects = append(projects, projectDir)
			}
		}
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("find projects for module dependants: %w", err)
	}

	result := make(moduleInfo)
	var diags tfconfig.Diagnostics
	// for each project, find the modules it depends on, their deps, etc.
	for _, projectDir := range projects {
		if _, err := result.load(files, projectDir, projectDir); err != nil {
			diags = append(diags, err...)
		}
	}
	// if there are any errors, prefer one with a source location
	if diags.HasErrors() {
		for _, d := range diags {
			if d.Pos != nil {
				return nil, fmt.Errorf("%s:%d - %s: %s", d.Pos.Filename, d.Pos.Line, d.Summary, d.Detail)
			}
		}
	}
	return result, diags.Err()
}
