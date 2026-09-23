package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromEnv_CasePrefix(t *testing.T) {
	t.Setenv("CASE_GENERATED", "/new")

	cfg := InitializeEmpty()
	cfg.loadFromEnv()

	assert.Equal(t, "/new", cfg.overrides.String(Generated))
}

func TestConfigFileEnv_Used(t *testing.T) {
	flags.configFilePath = ""

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yml")
	require.NoError(t, os.WriteFile(configFile, []byte("generated: /tmp/gen\n"), 0o644))

	t.Setenv("CASE_CONFIG_FILE", configFile)

	cfg := InitializeEmpty()
	require.NoError(t, cfg.initConfig())

	assert.False(t, cfg.IsNewSystem())
	assert.Equal(t, configFile, cfg.GetConfigFile())
	assert.Equal(t, "/tmp/gen", cfg.getString(Generated))
}

func TestFileEnvSet(t *testing.T) {
	t.Setenv("CASE_CONFIG_FILE", "")
	assert.False(t, FileEnvSet())

	t.Setenv("CASE_CONFIG_FILE", "/new.yml")
	assert.True(t, FileEnvSet())
}
