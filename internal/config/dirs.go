package config

import (
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func LauncherDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("User home dir: %s", home)

	return filepath.Join(home, ".cdt")
}

func LogsDir() string {
	return filepath.Join(LauncherDir(), "logs")
}

func createLogsDir() error {
	err := maybeCreateDir(LogsDir())
	if err != nil {
		return fmt.Errorf("cannot create the LOGS folder %s, err=%v", LogsDir(), err)
	}
	log.Tracef("LOGS folder: %s", LogsDir())

	return nil
}

func createLauncherDir() {
	err := maybeCreateDir(LauncherDir())
	if err != nil {
		log.Fatalf("cannot create the launcher folder %s, err=%v", LauncherDir(), err)
	}
	log.Infof("Create Launcher folder: %s", LauncherDir())
}

func maybeCreateDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}

	return nil
}
