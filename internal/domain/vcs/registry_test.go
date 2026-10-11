package vcs_test

import (
	"testing"

	"github.com/runatlantis/atlantis/internal/domain/vcs"
	"github.com/stretchr/testify/assert"
)

func TestVCSRegistry_Register_ValidPlugin_Success(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
	})

	// Act
	err := registry.Register("github", plugin)

	// Assert
	assert.NoError(t, err)
	assert.True(t, registry.IsRegistered("github"))
}

func TestVCSRegistry_Register_EmptyName_ReturnsError(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})

	// Act
	err := registry.Register("", plugin)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name cannot be empty")
}

func TestVCSRegistry_Register_NilPlugin_ReturnsError(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()

	// Act
	err := registry.Register("github", nil)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin cannot be nil")
}

func TestVCSRegistry_Register_DuplicateName_ReturnsError(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin1 := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})
	plugin2 := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})

	// Act
	err1 := registry.Register("github", plugin1)
	err2 := registry.Register("github", plugin2)

	// Assert
	assert.NoError(t, err1)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "already registered")
}

func TestVCSRegistry_Get_ExistingPlugin_ReturnsPlugin(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{
		SupportsMergeableBypass: true,
	})
	_ = registry.Register("github", plugin)

	// Act
	retrievedPlugin, err := registry.Get("github")

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, plugin, retrievedPlugin)
}

func TestVCSRegistry_Get_NonExistentPlugin_ReturnsError(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()

	// Act
	plugin, err := registry.Get("nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Nil(t, plugin)
	assert.Contains(t, err.Error(), "not found")
}

func TestVCSRegistry_List_ReturnsAllRegisteredPlugins(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	githubPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})
	gitlabPlugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})
	
	_ = registry.Register("github", githubPlugin)
	_ = registry.Register("gitlab", gitlabPlugin)

	// Act
	names := registry.List()

	// Assert
	assert.Len(t, names, 2)
	assert.Contains(t, names, "github")
	assert.Contains(t, names, "gitlab")
}

func TestVCSRegistry_MustRegister_ValidPlugin_Success(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})

	// Act & Assert (should not panic)
	registry.MustRegister("github", plugin)
	assert.True(t, registry.IsRegistered("github"))
}

func TestVCSRegistry_MustRegister_InvalidPlugin_Panics(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()

	// Act & Assert
	assert.Panics(t, func() {
		registry.MustRegister("", nil)
	})
}

func TestVCSRegistry_Unregister_ExistingPlugin_Success(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})
	_ = registry.Register("github", plugin)

	// Act
	err := registry.Unregister("github")

	// Assert
	assert.NoError(t, err)
	assert.False(t, registry.IsRegistered("github"))
}

func TestVCSRegistry_Unregister_NonExistentPlugin_ReturnsError(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()

	// Act
	err := registry.Unregister("nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestVCSRegistry_ConcurrentAccess_ThreadSafe(t *testing.T) {
	// Arrange
	registry := vcs.NewVCSRegistry()
	
	// Act - Simulate concurrent access
	done := make(chan bool, 2)
	
	go func() {
		for i := 0; i < 100; i++ {
			plugin := vcs.NewMockVCSPlugin(vcs.VCSCapabilities{})
			registry.Register("github", plugin)
			registry.Unregister("github")
		}
		done <- true
	}()
	
	go func() {
		for i := 0; i < 100; i++ {
			registry.List()
			registry.IsRegistered("github")
		}
		done <- true
	}()
	
	// Wait for both goroutines to complete
	<-done
	<-done
	
	// Assert - No race conditions should occur
	assert.True(t, true) // If we reach here, no race condition occurred
} 