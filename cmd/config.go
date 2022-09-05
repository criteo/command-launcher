package cmd

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ConfigFlags struct {
	Json bool
}

var (
	configFlags = ConfigFlags{}
)

func AddConfigCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configurations",
		Long: fmt.Sprintf(`Manage the command launcher configurations

  Example:
    get configuration
    %s config [key]

    set configuration
    %s config [key] [value]
	`, appCtx.AppName(), appCtx.AppName()),
		Run: func(cmd *cobra.Command, args []string) {
			// list all configs
			if len(args) == 0 {
				settings := viper.AllSettings()
				printSettings(settings, configFlags.Json)
			}

			// get configuration with key
			if len(args) == 1 {
				if viper.Get(args[0]) == nil {
					return
				}
				if configFlags.Json {
					printSettings(map[string]interface{}{
						args[0]: viper.Get(args[0]),
					}, true)
				} else {
					fmt.Println(viper.Get(args[0]))
				}

			}

			// set configuration with key
			if len(args) == 2 {
				if err := config.SetSettingValue(args[0], args[1]); err != nil {
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

			lowerKeys := []string{}
			for _, k := range config.SettingKeys {
				lowerKeys = append(lowerKeys, strings.ToLower(k))
			}

			return lowerKeys, cobra.ShellCompDirectiveNoFileComp
		},
	}

	configCmd.Flags().BoolVarP(&configFlags.Json, "json", "", false, "output in JSON format")
	rootCmd.AddCommand(configCmd)
}

// get printable settings in alphabet order
func printableSettingsInOrder(settings map[string]interface{}) []string {
	sorted := []string{}
	keys := []string{}
	for k := range settings {
		keys = append(keys, k)
	}

	sort.Strings(keys)
	for _, k := range keys {
		sorted = append(sorted, fmt.Sprintf("%-40v: %v", k, settings[k]))
	}

	return sorted
}

func printSettings(settings map[string]interface{}, jsonFormat bool) {
	if jsonFormat {
		printInJson(settings)
	} else {
		printInPlainText(settings)
	}
}

func printInJson(settings map[string]interface{}) {
	val, err := json.MarshalIndent(settings, "", "    ")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(val))
	}
}

func printInPlainText(settings map[string]interface{}) {
	printableSettings := printableSettingsInOrder(settings)
	for _, line := range printableSettings {
		fmt.Println(line)
	}
}
