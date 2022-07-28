package command

import (
	"fmt"
	"path"
	"strings"
)

// NewComment constructs a Command, setting all missing fields to defaults.
func NewComment(repoRelDir string, flags []string, name Name, forceApply bool, workspace string, project string, logLevel string) *Comment {
	// If repoRelDir was empty we want to keep it that way to indicate that it
	// wasn't specified in the comment.
	if repoRelDir != "" {
		repoRelDir = path.Clean(repoRelDir)
		if repoRelDir == "/" {
			repoRelDir = "."
		}
	}
	return &Comment{
		RepoRelDir:  repoRelDir,
		Flags:       flags,
		Name:        name,
		Workspace:   workspace,
		ProjectName: project,
		ForceApply:  forceApply,
		LogLevel:    logLevel,
	}
}

// Comment is a command that was triggered by a pull request comment.
type Comment struct {
	// RepoRelDir is the path relative to the repo root to run the command in.
	// Will never end in "/". If empty then the comment specified no directory.
	RepoRelDir string
	// Flags are the extra arguments appended to the comment,
	// ex. atlantis plan -- -target=resource
	Flags []string
	// Name is the name of the command the comment specified.
	Name Name
	//ForceApply is true of the command should ignore apply_requirments.
	ForceApply bool
	// Workspace is the name of the Terraform workspace to run the command in.
	// If empty then the comment specified no workspace.
	Workspace string
	// ProjectName is the name of a project to run the command on. It refers to a
	// project specified in an atlantis.yaml file.
	// If empty then the comment specified no project.
	ProjectName string
	// LogLevel is the name log level verbosity requested on the underlying Terraform operation.
	LogLevel string
}

// IsForSpecificProject returns true if the command is for a specific dir, workspace
// or project name. Otherwise it's a command like "atlantis plan" or "atlantis
// apply".
func (c Comment) IsForSpecificProject() bool {
	return c.RepoRelDir != "" || c.Workspace != "" || c.ProjectName != ""
}

// CommandName returns the name of this command.
func (c Comment) CommandName() Name {
	return c.Name
}

// IsAutoplan will be false for comment commands.
func (c Comment) IsAutoplan() bool {
	return false
}

// String returns a string representation of the command.
func (c Comment) String() string {
	return fmt.Sprintf("command=%q dir=%q workspace=%q project=%q loglevel=%q flags=%q", c.Name.String(), c.RepoRelDir, c.Workspace, c.ProjectName, c.LogLevel, strings.Join(c.Flags, ","))
}
