package config

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func CdtDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("User home dir: %s", home)

	return filepath.Join(home, ".cdt")
}

func LogsDir() string {
	return filepath.Join(CdtDir(), "logs")
}

func createLogsDir() error {
	err := maybeCreateDir(LogsDir())
	if err != nil {
		return fmt.Errorf("cannot create the LOGS folder %s, err=%v", LogsDir(), err)
	}
	log.Tracef("LOGS folder: %s", LogsDir())

	return nil
}

func createCdtDir() {
	err := maybeCreateDir(CdtDir())
	if err != nil {
		log.Fatalf("cannot create the CDT folder %s, err=%v", CdtDir(), err)
	}
	log.Infof("Create CDT folder: %s", CdtDir())
}

func maybeCreateDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}

	return nil
}
