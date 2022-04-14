package config

import (
	"os"
	"path/filepath"
	"time"

	"github.com/criteo/command-launcher/cmd/user"
	"github.com/criteo/command-launcher/internal/context"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

func LoadConfig(appCtx context.LauncherContext) {
	// NOTE: we don't put default value for the DEBUG_FLAGS configuration, it will not show in a newly created config file
	// Please keep it as a hidden config, better not to let developer directly see this option
	setDefaultConfig()

	cfgFile := os.Getenv(appCtx.ConfigurationFileEnvVar())
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(AppDir())
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

func setDefaultConfig() {
	viper.SetDefault(LOG_ENABLED_KEY, false)
	viper.SetDefault(LOG_LEVEL_KEY, "fatal") // trace, debug, info, warn, error, fatal, panic

	viper.SetDefault(SELF_UPDATE_ENABLED_KEY, false)
	viper.SetDefault(SELF_UPDATE_TIMEOUT_KEY, 2*time.Second) // In seconds
	viper.SetDefault(SELF_UPDATE_LATEST_VERSION_URL_KEY, "https://dummy/version")
	viper.SetDefault(SELF_UPDATE_BASE_URL_KEY, "https://dummy/")

	viper.SetDefault(COMMAND_UPDATE_ENABLED_KEY, false)
	viper.SetDefault(COMMAND_REPOSITORY_BASE_URL_KEY, "https://dummy/repos")

	viper.SetDefault(DROPIN_FOLDER_KEY, filepath.Join(AppDir(), "dropins"))
	viper.SetDefault(LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY, filepath.Join(AppDir(), "current"))

	viper.SetDefault(USAGE_METRICS_ENABLED_KEY, false)
	viper.SetDefault(METRIC_GRAPHITE_HOST_KEY, "dummy")

	viper.SetDefault(user.INTERNAL_COMMAND_ENABLED_KEY, false)
	viper.SetDefault(user.EXPERIMENTAL_COMMAND_ENABLED_KEY, false)
}

func initDefaultConfig() {
	log.Info("Create default config file")
	createAppDir()
	if err := viper.SafeWriteConfig(); err != nil {
		log.Error("cannot write the default configuration: ", err)
	}
}
