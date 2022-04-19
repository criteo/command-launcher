package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/criteo/command-launcher/internal/context"
	log "github.com/sirupsen/logrus"
)

func AppDir() string {
	ctx, err := context.AppContext()
	if err != nil {
		log.Fatal(err)
	}

	appDir := os.Getenv(ctx.AppHomeEnvVar())
	if appDir != "" {
		log.Tracef("Use app home dir from environment variable, %s: %s", ctx.AppHomeEnvVar(), appDir)
		return appDir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	log.Tracef("User home dir: %s", home)

	return filepath.Join(home, ctx.AppDirname())
}

func LogsDir() string {
	return filepath.Join(AppDir(), "logs")
}

func createLogsDir() error {
	err := maybeCreateDir(LogsDir())
	if err != nil {
		return fmt.Errorf("cannot create the LOGS folder %s, err=%v", LogsDir(), err)
	}
	log.Tracef("LOGS folder: %s", LogsDir())

	return nil
}

func createAppDir() {
	err := maybeCreateDir(AppDir())
	if err != nil {
		log.Fatalf("cannot create the App folder %s, err=%v", AppDir(), err)
	}
	log.Infof("App folder: %s", AppDir())
}

func maybeCreateDir(path string) error {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}

	return nil
}
