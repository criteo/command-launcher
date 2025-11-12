package updateConfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

type UpdateConfig struct {
	Date time.Time `json:"updateAfter"`
}

const DEFAULT_UPDATE_LOCK_DURATION = 24 * time.Hour
const PACKAGE_UPDATE_LOCK_FILE = ".update"

func ReadFromDir(dir string) (*UpdateConfig, error) {
	path := filepath.Join(dir, PACKAGE_UPDATE_LOCK_FILE)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config UpdateConfig
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func IsUpdateConfigExists(dir string) (bool, error) {
	path := filepath.Join(dir, PACKAGE_UPDATE_LOCK_FILE)
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
	}
}

func (config *UpdateConfig) WriteToDir(dir string) error {
	path := filepath.Join(dir, PACKAGE_UPDATE_LOCK_FILE)
	jsonData, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, jsonData, 0644)
}

func (config *UpdateConfig) IsExpired() bool {
	return time.Now().After(config.Date)
}

func (config *UpdateConfig) UpdateAfterDate(duration time.Duration) {
	config.Date = time.Now().Add(duration)
}
