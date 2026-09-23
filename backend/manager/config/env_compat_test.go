package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnv_NewPrefixPriority(t *testing.T) {
	t.Setenv("STASH_GENERATED", "/old")
	t.Setenv("CASE_GENERATED", "/new")

	cfg := InitializeEmpty()
	cfg.loadFromEnv()

	assert.Equal(t, "/new", cfg.overrides.String(Generated))
}

func TestLoadFromEnv_LegacyPrefixFallback(t *testing.T) {
	t.Setenv("STASH_GENERATED", "/old")

	cfg := InitializeEmpty()
	cfg.loadFromEnv()

	assert.Equal(t, "/old", cfg.overrides.String(Generated))
}

func TestConfigFileEnv_NewPriority(t *testing.T) {
	flags.configFilePath = ""

	dir := t.TempDir()
	newPath := filepath.Join(dir, "new.yml")
	oldPath := filepath.Join(dir, "old.yml")

	t.Setenv("CASE_CONFIG_FILE", newPath)
	t.Setenv("STASH_CONFIG_FILE", oldPath)

	cfg := InitializeEmpty()
	require.NoError(t, cfg.initConfig())

	assert.True(t, cfg.IsNewSystem())
	assert.Equal(t, newPath, cfg.GetConfigFile())
}

func TestConfigFileEnv_LegacyFallback(t *testing.T) {
	flags.configFilePath = ""

	oldPath := filepath.Join(t.TempDir(), "old.yml")

	t.Setenv("CASE_CONFIG_FILE", "")
	t.Setenv("STASH_CONFIG_FILE", oldPath)

	cfg := InitializeEmpty()
	require.NoError(t, cfg.initConfig())

	assert.True(t, cfg.IsNewSystem())
	assert.Equal(t, oldPath, cfg.GetConfigFile())
}

func TestFileEnvSet_NewAndLegacy(t *testing.T) {
	t.Setenv("CASE_CONFIG_FILE", "")
	t.Setenv("STASH_CONFIG_FILE", "")
	assert.False(t, FileEnvSet())

	t.Setenv("STASH_CONFIG_FILE", "/legacy.yml")
	assert.True(t, FileEnvSet())

	t.Setenv("CASE_CONFIG_FILE", "/new.yml")
	assert.True(t, FileEnvSet())
}
