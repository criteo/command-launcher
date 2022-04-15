package cmd

import (
	"fmt"
	"strings"

	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	rootCmd.AddCommand(configCmd)
}
