package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/criteo/command-launcher/internal/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configurations",
	Long: fmt.Sprintf(`Manage the command launcher configurations

Example:
  get configuration
    %s config [key]

  set configuration
    %s config [key] [value]
`, rootCmd.Use, rootCmd.Use),
	Run: func(cmd *cobra.Command, args []string) {
		// list all configs
		if len(args) == 0 {
			settings := viper.AllSettings()
			for k, v := range settings {
				fmt.Printf("%-40v: %v\n", k, v)
			}
		}

		// get configuration with key
		if len(args) == 1 {
			if viper.Get(args[0]) == nil {
				return
			}
			fmt.Println(viper.Get(args[0]))
		}

		// set configuration with key
		if len(args) == 2 {
			if err := setConfig(args[0], args[1]); err != nil {
				fmt.Println(err)
				return
			}
			if err := viper.WriteConfig(); err != nil {
				log.Error("cannot write the default configuration: ", err)
				return
			}
		}
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		configableKeys := []string{
			config.LOG_ENABLED_KEY,
			config.LOG_LEVEL_KEY,
			config.SELF_UPDATE_ENABLED_KEY,
			config.SELF_UPDATE_TIMEOUT_KEY,
			config.SELF_UPDATE_LATEST_VERSION_URL_KEY,
			config.SELF_UPDATE_BASE_URL_KEY,
			config.COMMAND_REPOSITORY_BASE_URL_KEY,
			config.LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY,
			config.USAGE_METRICS_ENABLED_KEY,
			config.METRIC_GRAPHITE_HOST_KEY,
			config.DEBUG_FLAGS_KEY,
			config.DROPIN_FOLDER_KEY,
		}

		lowerKeys := []string{}
		for _, k := range configableKeys {
			lowerKeys = append(lowerKeys, strings.ToLower(k))
		}

		return lowerKeys, cobra.ShellCompDirectiveNoFileComp
	},
}

func setConfig(key string, value string) error {
	upperKey := strings.ToUpper(key)
	switch upperKey {
	case config.LOG_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case config.LOG_LEVEL_KEY:
		return setLogLevelConfig(value)
	case config.SELF_UPDATE_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case config.SELF_UPDATE_TIMEOUT_KEY:
		return setDurationConfig(upperKey, value)
	case config.SELF_UPDATE_BASE_URL_KEY:
		return setStringConfig(upperKey, value)
	case config.SELF_UPDATE_LATEST_VERSION_URL_KEY:
		return setStringConfig(upperKey, value)
	case config.COMMAND_REPOSITORY_BASE_URL_KEY:
		return setStringConfig(upperKey, value)
	case config.LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY:
		return setStringConfig(upperKey, value)
	case config.USAGE_METRICS_ENABLED_KEY:
		return setBooleanConfig(upperKey, value)
	case config.METRIC_GRAPHITE_HOST_KEY:
		return setStringConfig(upperKey, value)
	case config.DEBUG_FLAGS_KEY:
		return setStringConfig(upperKey, value)
	case config.DROPIN_FOLDER_KEY:
		return setStringConfig(upperKey, value)
	}

	return fmt.Errorf("unsupported config %s", key)
}

func setBooleanConfig(key string, value string) error {
	if value == "true" {
		viper.Set(key, true)
		return nil
	} else if value == "false" {
		viper.Set(key, false)
		return nil
	}
	return fmt.Errorf("invalid format for boolean type")
}

func setDurationConfig(key string, value string) error {
	if d, err := time.ParseDuration(value); err != nil {
		viper.Set(key, d)
		return nil
	} else {
		return fmt.Errorf("invalid format for duration type")
	}
}

func setStringConfig(key string, value string) error {
	viper.Set(key, value)
	return nil
}

func setLogLevelConfig(value string) error {
	_, err := log.ParseLevel(strings.ToLower(value))
	if err == nil {
		viper.Set(config.LOG_LEVEL_KEY, strings.ToLower(value))
	}
	return err
}
