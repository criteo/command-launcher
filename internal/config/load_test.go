package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
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
