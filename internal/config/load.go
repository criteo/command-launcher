package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/criteo/command-launcher/cmd/user"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

const (
	LOG_ENABLED_KEY                      = "LOG_ENABLED"
	LOG_LEVEL_KEY                        = "LOG_LEVEL"
	SELF_UPDATE_ENABLED_KEY              = "SELF_UPDATE_ENABLED"
	SELF_UPDATE_TIMEOUT_KEY              = "SELF_UPDATE_TIMEOUT"
	SELF_UPDATE_LATEST_VERSION_URL_KEY   = "SELF_UPDATE_LATEST_VERSION_URL"
	SELF_UPDATE_BASE_URL_KEY             = "SELF_UPDATE_BASE_URL"
	COMMAND_REPOSITORY_BASE_URL_KEY      = "COMMAND_REPOSITORY_BASE_URL"
	LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY = "LOCAL_COMMAND_REPOSITORY_DIRNAME"
	USAGE_METRICS_ENABLED_KEY            = "USAGE_METRICS_ENABLED"
	METRIC_GRAPHITE_HOST_KEY             = "METRIC_GRAPHITE_HOST"
	DEBUG_FLAGS_KEY                      = "DEBUG_FLAGS"
	DROPIN_FOLDER_KEY                    = "DROPIN_FOLDER"
)

func LoadConfig() {
	// NOTE: we don't put default value for the DEBUG_FLAGS configuration, it will not show in a newly created config file
	// Please keep it as a hidden config, better not to let developer directly see this option
	SetDefaultConfig()

	cfgFile := os.Getenv("CDT_CONFIG_FILE")
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(CdtDir())
		viper.SetConfigType("json")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			initDefaultConfig()
		} else {
			log.Fatal("Cannot read configuration file: ", err)
		}
	}
}

func SetDefaultConfig() {
	viper.SetDefault(LOG_ENABLED_KEY, false)
	viper.SetDefault(LOG_LEVEL_KEY, "fatal") // trace, debug, info, warn, error, fatal, panic
	viper.SetDefault(SELF_UPDATE_ENABLED_KEY, true)
	viper.SetDefault(SELF_UPDATE_TIMEOUT_KEY, 2*time.Second) // In seconds

	viper.SetDefault(SELF_UPDATE_LATEST_VERSION_URL_KEY, "https://dummy/version")
	viper.SetDefault(SELF_UPDATE_BASE_URL_KEY, "https://dummy/")
	viper.SetDefault(COMMAND_REPOSITORY_BASE_URL_KEY, "https://dummy/repos")

	viper.SetDefault(DROPIN_FOLDER_KEY, filepath.Join(CdtDir(), "dropins"))
	viper.SetDefault(LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY, filepath.Join(CdtDir(), "current"))

	viper.SetDefault(USAGE_METRICS_ENABLED_KEY, true)

	viper.SetDefault(user.INTERNAL_COMMAND_ENABLED_KEY, false)
	viper.SetDefault(user.EXPERIMENTAL_COMMAND_ENABLED_KEY, false)
}

func initDefaultConfig() {
	log.Info("Create default config file")
	createCdtDir()
	if err := viper.SafeWriteConfig(); err != nil {
		log.Error("cannot write the default configuration: ", err)
	}
}
