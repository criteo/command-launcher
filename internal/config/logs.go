package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/criteo/command-launcher/internal/console"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Initialize the log file with a the prefix passed as arguments.
//
// The log level is defined in the configuration settings
func InitLog(prefix string) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	if viper.GetBool(LOG_ENABLED_KEY) {
		lvl, err := log.ParseLevel(viper.GetString(LOG_LEVEL_KEY))
		if err != nil {
			log.SetLevel(log.FatalLevel) // Log on console the fatal errors
			console.Warn("Cannot parse the log level")
		} else {
			log.SetLevel(lvl)
		}

		err = createLogsDir()
		if err != nil {
			console.Error("cannot create the LOGS dir, err=%v", err)
			return
		}

		file, err := os.OpenFile(logFilename(prefix), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			log.SetOutput(file)
		} else {
			console.Warn("Cannot create log file")
		}
	}
	log.Debugf("Config file is loaded from %s, reason: %s", configMetadata.File, configMetadata.Reason)
}

func logFilename(prefix string) string {
	filename := fmt.Sprintf("%s-%s.log", prefix, time.Now().Format("2006-01-02"))
	return filepath.Join(LogsDir(), filename)
}
