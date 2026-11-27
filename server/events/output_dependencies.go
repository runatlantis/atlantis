// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

// ProjectKey identifies a project within a plan run.
type ProjectKey struct {
	RepoRelDir  string
	Workspace   string
	ProjectName string
}

// NewProjectKey returns the key Atlantis uses to de-duplicate projects in a
// single plan run.
func NewProjectKey(repoRelDir, workspace, projectName string) ProjectKey {
	return ProjectKey{
		RepoRelDir:  filepath.ToSlash(filepath.Clean(repoRelDir)),
		Workspace:   workspace,
		ProjectName: projectName,
	}
}

func (k ProjectKey) String() string {
	return fmt.Sprintf("%s (workspace %q, project %q)", k.RepoRelDir, k.Workspace, k.ProjectName)
}

// OutputDependentTrigger describes a planned project whose outputs changed.
type OutputDependentTrigger struct {
	// RepoRelDir is the producer project's directory relative to the repo root.
	RepoRelDir string
	// ProjectName is the producer's project name, if it has one.
	ProjectName string
	// Workspace is the producer's workspace.
	Workspace string
	// ChangedOutputs is the set of output names the plan changes.
	ChangedOutputs map[string]bool
}

// OutputDependencyIndex maps a producer project directory to the projects in
// the same repository that read that project's state outputs.
type OutputDependencyIndex struct {
	// dependents is keyed by producer repo-relative directory because a
	// remote state backend identifies a directory's state, not one of the
	// projects that may be configured for it at a particular workspace.
	dependents map[string]map[ProjectKey]*outputDependent
}

type outputDependent struct {
	key ProjectKey
	// outputs is the set of output names the dependent reads. Nil means it may
	// read any output and can only be checked when the producer's changed
	// output names are unknown.
	outputs map[string]bool
	// sources records how the edge was discovered, for logging.
	sources []string
}

// BuildOutputDependencyIndex inspects every project in the repository for
// dependencies on another project's Terraform outputs. Two kinds of
// dependency are recognised:
//
//   - a project that names another project in atlantis.yaml depends_on;
//   - a data "terraform_remote_state" block whose backend configuration
//     resolves to another project's backend in this repository.
//
// The index is best effort: configuration that cannot be resolved statically
// is skipped (and logged) rather than guessed at. Projects that use those
// patterns can still declare the dependency through depends_on.
func BuildOutputDependencyIndex(log logging.SimpleLogging, repoDir string, projects []valid.MergedProjectCfg) *OutputDependencyIndex {
	index := &OutputDependencyIndex{dependents: make(map[string]map[ProjectKey]*outputDependent)}

	// Project names are the only handle depends_on has. Atlantis does not
	// require them to be unique across workspaces, so one name can resolve to
	// several directories; treat all of them as producers.
	dirsByName := make(map[string][]string)
	for _, project := range projects {
		if project.Name == "" {
			continue
		}
		dirsByName[project.Name] = append(dirsByName[project.Name], project.RepoRelDir)
	}

	for _, project := range projects {
		dependent := NewProjectKey(project.RepoRelDir, project.Workspace, project.Name)
		for _, dependency := range project.DependsOn {
			dirs, ok := dirsByName[dependency]
			if !ok {
				log.Debug("project %s declares depends_on %q but no project in the repo has that name", dependent, dependency)
				continue
			}
			for _, dir := range dirs {
				index.add(dir, dependent, nil, "depends_on")
			}
		}
	}

	specsByDir := make(map[string][]backendSpec)
	for _, project := range projects {
		dir := filepath.ToSlash(filepath.Clean(project.RepoRelDir))
		if _, ok := specsByDir[dir]; ok {
			continue
		}
		specs, err := backendSpecsForProject(repoDir, dir)
		if err != nil {
			log.Debug("unable to inspect Terraform backends for project %s: %s", dir, err)
		}
		specsByDir[dir] = specs
	}

	for _, project := range projects {
		dir := filepath.ToSlash(filepath.Clean(project.RepoRelDir))
		refs, err := remoteStateRefsForProject(repoDir, dir)
		if err != nil {
			log.Debug("unable to inspect remote state references for project %s: %s", dir, err)
			continue
		}
		if len(refs) == 0 {
			continue
		}
		dependent := NewProjectKey(project.RepoRelDir, project.Workspace, project.Name)
		for _, ref := range refs {
			var matched []string
			for producerDir, specs := range specsByDir {
				if producerKey(producerDir) == producerKey(dir) {
					// A project reading its own state is not a cross-project
					// dependency and must not make it plan twice.
					continue
				}
				for _, spec := range specs {
					if spec.matches(dir, ref.backendType, ref.config) {
						matched = append(matched, producerDir)
						break
					}
				}
			}
			if len(matched) == 0 {
				log.Debug("remote state data source %q in project %s does not resolve to another project's backend in this repository; ignoring", ref.name, dir)
				continue
			}
			sort.Strings(matched)
			if len(matched) > 1 {
				log.Debug("remote state data source %q in project %s matches multiple project backends (%s); treating all of them as producers", ref.name, dir, strings.Join(matched, ", "))
			}
			for _, producerDir := range matched {
				index.add(producerDir, dependent, ref.outputs, "remote_state:"+ref.name)
			}
		}
	}

	return index
}

func producerKey(dir string) string {
	return filepath.ToSlash(filepath.Clean(dir))
}

func (i *OutputDependencyIndex) add(producerDir string, dependent ProjectKey, outputs map[string]bool, source string) {
	if i == nil {
		return
	}
	if producerKey(producerDir) == dependent.RepoRelDir {
		return
	}
	if i.dependents[producerDir] == nil {
		i.dependents[producerDir] = make(map[ProjectKey]*outputDependent)
	}
	existing, ok := i.dependents[producerDir][dependent]
	if !ok {
		existing = &outputDependent{key: dependent, outputs: cloneOutputSet(outputs)}
		i.dependents[producerDir][dependent] = existing
	} else if existing.outputs != nil {
		if outputs == nil {
			// One edge reads the whole output map, so the dependent must be
			// re-planned whenever any output changes.
			existing.outputs = nil
		} else {
			for output := range outputs {
				existing.outputs[output] = true
			}
		}
	}
	for _, s := range existing.sources {
		if s == source {
			return
		}
	}
	existing.sources = append(existing.sources, source)
}

func cloneOutputSet(in map[string]bool) map[string]bool {
	if in == nil {
		return nil
	}
	out := make(map[string]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// Dependents returns the projects that read the outputs of any trigger and
// that are not already in exclude. A dependent is skipped when it reads a
// known set of outputs and none of them changed.
func (i *OutputDependencyIndex) Dependents(triggers []OutputDependentTrigger, exclude map[ProjectKey]bool) []ProjectKey {
	if i == nil {
		return nil
	}
	selected := make(map[ProjectKey]bool)
	for _, trigger := range triggers {
		if len(trigger.ChangedOutputs) == 0 {
			continue
		}
		for key, dependent := range i.dependents[producerKey(trigger.RepoRelDir)] {
			if exclude[key] || key.RepoRelDir == producerKey(trigger.RepoRelDir) {
				continue
			}
			if len(dependent.outputs) > 0 && !anyOutputChanged(dependent.outputs, trigger.ChangedOutputs) {
				continue
			}
			selected[key] = true
		}
	}
	keys := make([]ProjectKey, 0, len(selected))
	for key := range selected {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(a, b int) bool {
		if keys[a].RepoRelDir != keys[b].RepoRelDir {
			return keys[a].RepoRelDir < keys[b].RepoRelDir
		}
		if keys[a].Workspace != keys[b].Workspace {
			return keys[a].Workspace < keys[b].Workspace
		}
		return keys[a].ProjectName < keys[b].ProjectName
	})
	return keys
}

func anyOutputChanged(dependent, changed map[string]bool) bool {
	for output := range dependent {
		if changed[output] {
			return true
		}
	}
	return false
}

// backendSpec is a statically resolvable terraform backend block.
type backendSpec struct {
	typeName string
	config   map[string]string
	// localPath is the repo-relative state file path for the local backend.
	localPath string
}

// remoteStateRef is a data "terraform_remote_state" block plus the outputs the
// project reads from it.
type remoteStateRef struct {
	name        string
	backendType string
	config      map[string]string
	outputs     map[string]bool
}

// backendIdentityKeys are the backend options that identify a state file.
// Credentials, regions and other access options are deliberately excluded: a
// bucket or storage account is globally unique, so a matching key there is
// enough to identify the state without requiring identical credentials.
var backendIdentityKeys = map[string][]string{
	"local":      {"path"},
	"s3":         {"bucket", "key", "workspace_key_prefix"},
	"gcs":        {"bucket", "prefix"},
	"azurerm":    {"storage_account_name", "container_name", "key"},
	"remote":     {"hostname", "organization", "workspaces.name"},
	"pg":         {"conn_str", "schema_name"},
	"consul":     {"path"},
	"http":       {"address"},
	"cos":        {"bucket", "prefix"},
	"oss":        {"bucket", "prefix"},
	"kubernetes": {"namespace", "secret_suffix"},
}

func applyBackendDefaults(typeName string, config map[string]string) {
	switch typeName {
	case "gcs":
		if _, ok := config["prefix"]; !ok {
			config["prefix"] = "terraform/state"
		}
	case "s3":
		if _, ok := config["workspace_key_prefix"]; !ok {
			config["workspace_key_prefix"] = "env:"
		}
	case "remote":
		if _, ok := config["hostname"]; !ok {
			config["hostname"] = "app.terraform.io"
		}
	}
}

func (s backendSpec) matches(consumerDir, consumerType string, consumerConfig map[string]string) bool {
	if s.typeName != consumerType {
		return false
	}
	consumer := cloneStringMap(consumerConfig)
	applyBackendDefaults(consumerType, consumer)
	producer := cloneStringMap(s.config)
	applyBackendDefaults(s.typeName, producer)

	if s.typeName == "local" {
		if s.localPath == "" {
			return false
		}
		consumerPath, ok := resolveStatePath(consumerDir, consumer["path"])
		return ok && consumerPath == s.localPath
	}

	keys, known := backendIdentityKeys[s.typeName]
	if !known || len(keys) == 0 {
		// Unknown backend: fall back to every option the consumer set. That is
		// conservative (it can miss a match when the producer relies on
		// defaults) but cannot invent a dependency between unrelated states.
		keys = sortedKeys(consumer)
	}
	for _, key := range keys {
		producerValue, producerOK := producer[key]
		consumerValue, consumerOK := consumer[key]
		if !producerOK || !consumerOK || producerValue != consumerValue {
			return false
		}
	}
	return true
}

func cloneStringMap(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func resolveStatePath(dir, path string) (string, bool) {
	if path == "" {
		return "", false
	}
	// Terraform resolves a local backend path relative to the working
	// directory of the module that declares it.
	joined := filepath.ToSlash(filepath.Clean(filepath.Join(dir, filepath.FromSlash(path))))
	if joined == ".." || strings.HasPrefix(joined, "../") {
		return "", false
	}
	return joined, true
}

func newBackendSpec(typeName string, config map[string]string, projectDir string) backendSpec {
	spec := backendSpec{typeName: typeName, config: config}
	if typeName == "local" {
		path := config["path"]
		if path == "" {
			path = "terraform.tfstate"
		}
		if resolved, ok := resolveStatePath(projectDir, path); ok {
			spec.localPath = resolved
		}
	}
	return spec
}

// projectFiles returns the Terraform files directly inside the project
// directory. Child modules are not walked: a remote state block in a module
// cannot be attributed to one of the repository's projects without resolving
// module sources, and guessing would create spurious plans.
func projectFiles(repoDir, projectDir string) ([]namedFile, error) {
	rel, err := safeRelPath(repoDir, projectDir)
	if err != nil {
		return nil, err
	}
	abs := filepath.Join(repoDir, filepath.FromSlash(rel))
	entries, err := os.ReadDir(abs)
	if err != nil {
		return nil, err
	}
	var files []namedFile
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".tf") && !strings.HasSuffix(name, ".tofu") {
			continue
		}
		// OpenInRoot refuses a symlink that resolves outside the directory, so
		// a pull request cannot point a Terraform file at host content.
		file, err := os.OpenInRoot(abs, name)
		if err != nil {
			return nil, err
		}
		src, readErr := io.ReadAll(file)
		closeErr := file.Close()
		if readErr != nil {
			return nil, readErr
		}
		if closeErr != nil {
			return nil, closeErr
		}
		files = append(files, namedFile{name: name, src: src})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	return files, nil
}

type namedFile struct {
	name string
	src  []byte
}

func safeRelPath(repoDir, projectDir string) (string, error) {
	absRepo, err := filepath.Abs(repoDir)
	if err != nil {
		return "", err
	}
	absProject, err := filepath.Abs(filepath.Join(repoDir, filepath.FromSlash(projectDir)))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absRepo, absProject)
	if err != nil {
		return "", err
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", fmt.Errorf("project directory %q escapes the repository", projectDir)
	}
	return rel, nil
}

func backendSpecsForProject(repoDir, projectDir string) ([]backendSpec, error) {
	files, err := projectFiles(repoDir, projectDir)
	if err != nil {
		return nil, err
	}
	var specs []backendSpec
	sawBackendBlock := false
	for _, file := range files {
		body, err := parseTerraformBody(file)
		if err != nil {
			continue
		}
		for _, block := range body.Blocks {
			if block.Type != "terraform" {
				continue
			}
			for _, inner := range block.Body.Blocks {
				switch inner.Type {
				case "backend":
					if len(inner.Labels) != 1 {
						continue
					}
					sawBackendBlock = true
					config, ok := evalBlockBody(inner.Body, "")
					if !ok {
						// The backend is configured dynamically. Terraform may
						// be told the real values through -backend-config, so
						// the state cannot be identified from code alone.
						continue
					}
					specs = append(specs, newBackendSpec(inner.Labels[0], config, projectDir))
				case "cloud":
					// A cloud block stores state in Terraform Cloud, which a
					// remote state data source can read through the remote
					// backend.
					sawBackendBlock = true
					config, ok := evalBlockBody(inner.Body, "")
					if !ok {
						continue
					}
					specs = append(specs, newBackendSpec("remote", config, projectDir))
				}
			}
		}
	}
	if !sawBackendBlock {
		specs = append(specs, newBackendSpec("local", map[string]string{}, projectDir))
	}
	return specs, nil
}

var remoteStateReferenceRE = regexp.MustCompile(`data\.terraform_remote_state\.([A-Za-z0-9_-]+)(?:\.outputs(?:\.([A-Za-z0-9_-]+))?)?`)

func remoteStateRefsForProject(repoDir, projectDir string) ([]remoteStateRef, error) {
	files, err := projectFiles(repoDir, projectDir)
	if err != nil {
		return nil, err
	}

	// First collect the declarations and their backend configuration.
	refsByName := make(map[string]*remoteStateRef)
	referenced := make(map[string]map[string]bool)
	for _, file := range files {
		body, err := parseTerraformBody(file)
		if err != nil {
			continue
		}
		for _, block := range body.Blocks {
			if block.Type != "data" || len(block.Labels) != 2 || block.Labels[0] != "terraform_remote_state" {
				continue
			}
			ref := &remoteStateRef{name: block.Labels[1], backendType: "local", config: map[string]string{}}
			if backendAttr, ok := block.Body.Attributes["backend"]; ok {
				value, diags := backendAttr.Expr.Value(nil)
				if diags.HasErrors() || !value.IsKnown() || value.IsNull() || value.Type() != cty.String {
					// The backend is chosen dynamically; no safe match.
					continue
				}
				ref.backendType = value.AsString()
			}
			if configAttr, ok := block.Body.Attributes["config"]; ok {
				value, diags := configAttr.Expr.Value(nil)
				if diags.HasErrors() || !value.IsKnown() || value.IsNull() {
					continue
				}
				config := make(map[string]string)
				if !flattenValue(value, "", config) {
					continue
				}
				ref.config = config
			}
			refsByName[ref.name] = ref
		}

		// Then find which outputs the rest of the module reads. This is a
		// textual scan on purpose: Terraform only exposes these references in
		// expression form, and the output name does not need full evaluation
		// to be useful.
		for _, match := range remoteStateReferenceRE.FindAllSubmatch(file.src, -1) {
			name := string(match[1])
			if referenced[name] == nil {
				referenced[name] = make(map[string]bool)
			}
			if len(match[2]) == 0 {
				// The whole output object is used, so any changed output can
				// affect this project.
				referenced[name] = nil
				continue
			}
			if referenced[name] != nil {
				referenced[name][string(match[2])] = true
			}
		}
	}

	var refs []remoteStateRef
	for name, ref := range refsByName {
		outputs, ok := referenced[name]
		if !ok {
			// Declared but never read; nothing to react to.
			continue
		}
		ref.outputs = outputs
		refs = append(refs, *ref)
	}
	sort.Slice(refs, func(i, j int) bool { return refs[i].name < refs[j].name })
	return refs, nil
}

func parseTerraformBody(file namedFile) (*hclsyntax.Body, error) {
	hclFile, diags := hclsyntax.ParseConfig(file.src, file.name, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("%s: %s", file.name, diags.Error())
	}
	body, ok := hclFile.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("%s: not an HCL body", file.name)
	}
	return body, nil
}

// evalBlockBody flattens a block's literal attributes and nested blocks into
// dotted keys. It returns false as soon as any value depends on a variable,
// function or other runtime value, because a half-evaluated backend
// configuration could match the wrong project.
func evalBlockBody(body *hclsyntax.Body, prefix string) (map[string]string, bool) {
	values := make(map[string]string)
	if !evalBodyInto(body, prefix, values) {
		return nil, false
	}
	return values, true
}

func evalBodyInto(body *hclsyntax.Body, prefix string, values map[string]string) bool {
	for name, attr := range body.Attributes {
		value, diags := attr.Expr.Value(nil)
		if diags.HasErrors() || !value.IsKnown() || value.IsNull() {
			return false
		}
		if !flattenValue(value, joinKey(prefix, name), values) {
			return false
		}
	}
	for _, block := range body.Blocks {
		blockPrefix := block.Type
		for _, label := range block.Labels {
			blockPrefix += "." + label
		}
		if !evalBodyInto(block.Body, joinKey(prefix, blockPrefix), values) {
			return false
		}
	}
	return true
}

func joinKey(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

// flattenValue writes a cty value into out using dotted keys. It returns false
// for collections that cannot be represented as a stable string identity.
func flattenValue(value cty.Value, prefix string, out map[string]string) bool {
	if !value.IsKnown() || value.IsNull() {
		return false
	}
	ty := value.Type()
	switch {
	case ty.IsObjectType() || ty.IsMapType():
		iter := value.ElementIterator()
		for iter.Next() {
			key, element := iter.Element()
			if !flattenValue(element, joinKey(prefix, key.AsString()), out) {
				return false
			}
		}
		return true
	case ty.IsTupleType() || ty.IsListType() || ty.IsSetType():
		return false
	case ty == cty.String:
		out[prefix] = value.AsString()
		return true
	case ty == cty.Number:
		out[prefix] = value.AsBigFloat().Text('f', -1)
		return true
	case ty == cty.Bool:
		out[prefix] = strconv.FormatBool(value.True())
		return true
	default:
		return false
	}
}
