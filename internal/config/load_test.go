package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_shouldHaveConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFileName := "config.json"

	file, err := os.Create(filepath.Join(tmpDir, configFileName))
	assert.Nil(t, err)
	defer file.Close()

	found := hasConfigFile(tmpDir, configFileName)
	assert.True(t, found)
}

func Test_shouldNotHaveConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFileName := "config.json"
	found := hasConfigFile(tmpDir, configFileName)
	assert.False(t, found)
}

func Test_shouldNotFindLocalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configFileName := "config.json"
	_, found := findLocalConfig(tmpDir, configFileName)
	assert.False(t, found)
}

func Test_shouldFindLocalConfig(t *testing.T) {
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "a", "b", "c", "d")
	err := os.MkdirAll(testDir, 0750)
	assert.Nil(t, err)

	// create config file in the path
	configFileName := "config.json"
	file, err := os.Create(filepath.Join(tmpDir, "a", configFileName))
	assert.Nil(t, err)
	defer file.Close()

	_, found := findLocalConfig(testDir, configFileName)
	assert.True(t, found)
}

func setupLoadConfigTest(t *testing.T) (context.LauncherContext, string) {
	t.Helper()
	appCtx := context.InitContext("cltest", "test", "1")
	home := t.TempDir()
	t.Setenv(appCtx.AppHomeEnvVar(), home)
	t.Setenv(appCtx.ConfigurationFileEnvVar(), "")
	t.Setenv(appCtx.RemoteConfigurationUrlEnvVar(), "")
	// no local cltest.json in the working dir or its parents
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(t.TempDir()))
	t.Cleanup(func() { os.Chdir(wd) })
	viper.Reset()
	t.Cleanup(viper.Reset)
	return appCtx, home
}

func TestLoadConfig_CreatesDefaultFile(t *testing.T) {
	appCtx, home := setupLoadConfigTest(t)

	LoadConfig(appCtx)

	cfg := filepath.Join(home, "config.json")
	content := readJSON(t, cfg)
	assert.Equal(t, "fatal", content["log_level"])
	assert.FileExists(t, filepath.Join(home, configLockFileName))
	assert.Empty(t, tempFiles(t, home))
}

func TestLoadConfig_PersistsRemoteConfig(t *testing.T) {
	appCtx, home := setupLoadConfigTest(t)
	remote := filepath.Join(t.TempDir(), "remote.json")
	require.NoError(t, os.WriteFile(remote, []byte(`{"log_level":"debug"}`), 0644))
	t.Setenv(appCtx.RemoteConfigurationUrlEnvVar(), remote)

	LoadConfig(appCtx)

	content := readJSON(t, filepath.Join(home, "config.json"))
	assert.Equal(t, "debug", content["log_level"])
	assert.Contains(t, content, "remote_config_check_time")
}

func TestWriteConfig_UsesLoadedConfigFile(t *testing.T) {
	appCtx, home := setupLoadConfigTest(t)
	LoadConfig(appCtx)

	viper.Set(LOG_LEVEL_KEY, "info")
	require.NoError(t, WriteConfig())

	assert.Equal(t, "info", readJSON(t, filepath.Join(home, "config.json"))["log_level"])
}
