package config

import (
	"bytes"
	"os"
	"path/filepath"
	"time"

	"github.com/criteo/command-launcher/cmd/user"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/helper"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/viper"
)

const REMOTE_CONFIG_HASH_KEY = "REMOTE_CONFIG_HASH"

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
			initDefaultConfigFile()
		} else {
			log.Fatal("Cannot read configuration file: ", err)
		}
	}

	if loadRemoteConfig(appCtx) {
		log.Info("Remote Configuration loaded...")
		viper.WriteConfig()
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
	viper.SetDefault(REMOTE_CONFIG_HASH_KEY, "")
}

func initDefaultConfigFile() {
	log.Info("Create default config file")
	createAppDir()
	if err := viper.SafeWriteConfig(); err != nil {
		log.Error("cannot write the default configuration: ", err)
	}
}

func loadRemoteConfig(appCtx context.LauncherContext) bool {
	if urlCfg := os.Getenv(appCtx.RemoteConfigurationUrlEnvVar()); urlCfg != "" {
		_, hash, err := helper.HttpEtag(urlCfg)
		if err != nil {
			log.Error("Cannot find the remote Configuration %s: %v", urlCfg, err)
		}

		prev := viper.GetString(REMOTE_CONFIG_HASH_KEY)
		if prev != hash {
			viper.Set(REMOTE_CONFIG_HASH_KEY, hash)
			data, err := helper.LoadFile(urlCfg)
			if err == nil {
				err = viper.ReadConfig(bytes.NewReader(data))
				return err == nil
			}
		}
	}

	return false
}
