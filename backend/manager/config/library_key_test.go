package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStashPaths_LibrariesKey(t *testing.T) {
	cfg := InitializeEmpty()
	cfg.SetInterface(Libraries, StashConfigs{{Path: "/new"}})

	got := cfg.GetStashPaths()
	require.Len(t, got, 1)
	assert.Equal(t, "/new", got[0].Path)
}

func TestGetStashPaths_LegacyStringList(t *testing.T) {
	cfg := InitializeEmpty()
	cfg.SetInterface(Libraries, []string{"/a", "/b"})

	got := cfg.GetStashPaths()
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
