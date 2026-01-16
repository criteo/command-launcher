package updateConfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// UpdateConfig stores pause information for all packages in a repository
type UpdateConfig struct {
	// Packages maps package name to the time after which updates are allowed
	Packages map[string]time.Time `json:"packages"`
}

const DEFAULT_UPDATE_PAUSE_DURATION = 24 * time.Hour
const PACKAGE_UPDATE_FILE = ".update"

// ReadFromDir reads the update config from the repository directory
func ReadFromDir(dir string) (*UpdateConfig, error) {
	path := filepath.Join(dir, PACKAGE_UPDATE_FILE)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config UpdateConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// Initialize map if nil (for backwards compatibility or empty file)
	if config.Packages == nil {
		config.Packages = make(map[string]time.Time)
	}

	return &config, nil
}

// IsUpdateConfigExists checks if the update config file exists in the directory
func IsUpdateConfigExists(dir string) (bool, error) {
	path := filepath.Join(dir, PACKAGE_UPDATE_FILE)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

// WriteToDir writes the update config to the repository directory
func (config *UpdateConfig) WriteToDir(dir string) error {
	path := filepath.Join(dir, PACKAGE_UPDATE_FILE)
	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}

// IsPackagePaused checks if a specific package is paused (not expired)
func (config *UpdateConfig) IsPackagePaused(packageName string) bool {
	if config.Packages == nil {
		return false
	}
	updateAfter, exists := config.Packages[packageName]
	if !exists {
		return false
	}
	// Paused if current time is BEFORE the updateAfter time
	return time.Now().Before(updateAfter)
}

// PausePackage sets the pause duration for a specific package
func (config *UpdateConfig) PausePackage(packageName string, duration time.Duration) {
	if config.Packages == nil {
		config.Packages = make(map[string]time.Time)
	}
	config.Packages[packageName] = time.Now().Add(duration)
}

// RemoveExpiredPauses removes all expired pause entries from the config
func (config *UpdateConfig) RemoveExpiredPauses() {
	if config.Packages == nil {
		return
	}
	now := time.Now()
	for pkg, updateAfter := range config.Packages {
		if now.After(updateAfter) {
			delete(config.Packages, pkg)
		}
	}
}

// NewUpdateConfig creates a new empty UpdateConfig
func NewUpdateConfig() *UpdateConfig {
	return &UpdateConfig{
		Packages: make(map[string]time.Time),
	}
}
