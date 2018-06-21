package events_test

//. "github.com/runatlantis/atlantis/testing"

//func TestBuildAutoplanCommands(t *testing.T) {
//	tmpDir, cleanup := TempDir(t)
//	defer cleanup()
//
//	workspace := mocks.NewMockAtlantisWorkspace()
//	vcsClient := vcsmocks.NewMockClientProxy()
//
//	builder := &events.DefaultProjectCommandBuilder{
//		AtlantisWorkspaceLocker: events.NewDefaultAtlantisWorkspaceLocker(),
//		Workspace:               workspace,
//		ParserValidator:         &yaml.ParserValidator{},
//		VCSClient:               vcsClient,
//		ProjectFinder:           &events.DefaultProjectFinder{},
//	}
//
//	// If autoplan is false, should return empty steps.
//	ctxs, err := builder.BuildAutoplanCommands(&events.CommandContext{
//		BaseRepo: models.Repo{},
//		HeadRepo: models.Repo{},
//		Pull:     models.PullRequest{},
//		User:     models.User{},
//		Log:      nil,
//	})
//	Ok(t, err)
//}
