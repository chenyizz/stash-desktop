package config

import (
	"os"
	"path/filepath"
	"testing"

	"case/backend/pkg/fsutil"
	"case/backend/pkg/models/paths"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigLocations_UseCaseDir(t *testing.T) {
	want := filepath.Join(fsutil.GetHomeDirectory(), ".case", "config.yml")
	assert.Contains(t, defaultConfigLocations, want)

	for _, location := range defaultConfigLocations {
		assert.NotContains(t, location, ".stash")
	}
}

func TestGetCaseHomeDirectory_DefaultsToCase(t *testing.T) {
	assert.Equal(t, filepath.Join(fsutil.GetHomeDirectory(), ".case"), paths.GetCaseHomeDirectory())
}

func TestGetDefaultDatabaseFilePath_UsesCaseDB(t *testing.T) {
	dir := t.TempDir()
	cfg := InitializeEmpty()
	cfg.SetConfigFile(filepath.Join(dir, "config.yml"))

	assert.Equal(t, filepath.Join(dir, "case.db"), cfg.GetDefaultDatabaseFilePath())
}

func TestInitConfig_ExplicitConfigFileWins(t *testing.T) {
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
