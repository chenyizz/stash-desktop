package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLibraryPaths_LibrariesKey(t *testing.T) {
	cfg := InitializeEmpty()
	cfg.SetInterface(Libraries, LibraryConfigs{{Path: "/new"}})

	got := cfg.GetLibraryPaths()
	require.Len(t, got, 1)
	assert.Equal(t, "/new", got[0].Path)
}

func TestGetLibraryPaths_LegacyStringList(t *testing.T) {
	cfg := InitializeEmpty()
	cfg.SetInterface(Libraries, []string{"/a", "/b"})

	got := cfg.GetLibraryPaths()
	require.Len(t, got, 2)
	assert.Equal(t, "/a", got[0].Path)
	assert.Equal(t, "/b", got[1].Path)
}

func TestLoadFromEnv_LibrariesBindsToLibraries(t *testing.T) {
	t.Setenv("CASE_LIBRARIES", "/env-new")

	cfg := InitializeEmpty()
	cfg.loadFromEnv()

	assert.True(t, cfg.overrides.Exists(Libraries))
}
