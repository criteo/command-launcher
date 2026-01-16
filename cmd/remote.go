package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func AddRemoteCmd(rootCmd *cobra.Command, appCtx context.LauncherContext, back backend.Backend) {
	remoteCmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage command launcher remotes",
		Long:  "Manage command launcher remotes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
			}
			return nil
		},
	}

	remoteListCmd := &cobra.Command{
		Use:   "list",
		Short: "List command launcher remotes",
		Long:  "List command launcher remotes",
		RunE: func(cmd *cobra.Command, args []string) error {
			allRemotes := getAllRemotes()
			for _, v := range allRemotes {
				fmt.Printf("%-15s : %s\n", v.Name, v.RemoteBaseUrl)
			}
			return nil
		},
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		},
	}

	remoteDeleteCmd := &cobra.Command{
		Use:   "delete [remote url]",
		Short: "Delete command launcher remote",
		Long:  "Delete command launcher remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Help()
				return nil
			}
			if args[0] == "default" {
				return fmt.Errorf("can't delete the default remote repository")
			}
			config.RemoveRemote(args[0])
			if err := viper.WriteConfig(); err != nil {
				log.Error("cannot write the default configuration: ", err)
				return err
			}
			return nil
		},
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 1 {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
			remotes := getAllRemotes()
			remoteNames := []string{}
			for _, remote := range remotes {
				remoteNames = append(remoteNames,
					fmt.Sprintf("%s\t%s", remote.Name, remote.RemoteBaseUrl),
				)
			}
			return remoteNames, cobra.ShellCompDirectiveNoFileComp
		},
	}

	var addSyncPolicy string
	remoteAddCmd := &cobra.Command{
		Use:   "add [remote name] [remote base url]",
		Short: "Add command launcher remote",
		Long:  "Add command launcher remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				cmd.Help()
				return nil
			}
			if args[0] == "default" {
				return fmt.Errorf("can't add remote named 'default', it is a reserved remote name")
			}
			if addSyncPolicy != "" && !config.IsValidSyncPolicy(addSyncPolicy) {
				return fmt.Errorf("invalid sync policy: %s, must be one of: %s", addSyncPolicy, strings.Join(config.ValidSyncPolicies(), ", "))
			}
			syncPolicy := addSyncPolicy
			if syncPolicy == "" {
				syncPolicy = "always"
			}
			repoDir := filepath.Join(config.AppDir(), args[0])
			if err := config.AddRemote(args[0], repoDir, args[1], syncPolicy); err != nil {
				return err
			}
			if err := viper.WriteConfig(); err != nil {
				log.Error("cannot write the default configuration: ", err)
				return err
			}
			return nil
		},
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		},
	}
	remoteAddCmd.Flags().StringVar(&addSyncPolicy, "sync-policy", "", "sync policy for the remote (never, always, hourly, daily, weekly, monthly)")
	remoteAddCmd.RegisterFlagCompletionFunc("sync-policy", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidSyncPolicies(), cobra.ShellCompDirectiveNoFileComp
	})

	var setSyncPolicy string
	remoteSetCmd := &cobra.Command{
		Use:   "set [remote name]",
		Short: "Update settings for an existing remote",
		Long:  "Update settings for an existing remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Help()
				return nil
			}
			if args[0] == "default" {
				return fmt.Errorf("can't modify the default remote")
			}
			if setSyncPolicy == "" {
				return fmt.Errorf("no settings to update, use --sync-policy to set the sync policy")
			}
			if !config.IsValidSyncPolicy(setSyncPolicy) {
				return fmt.Errorf("invalid sync policy: %s, must be one of: %s", setSyncPolicy, strings.Join(config.ValidSyncPolicies(), ", "))
			}
			if err := config.UpdateRemote(args[0], setSyncPolicy); err != nil {
				return err
			}
			if err := viper.WriteConfig(); err != nil {
				log.Error("cannot write the default configuration: ", err)
				return err
			}
			fmt.Printf("Remote '%s' sync policy updated to '%s'\n", args[0], setSyncPolicy)
			return nil
		},
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 1 {
				return []string{}, cobra.ShellCompDirectiveNoFileComp
			}
			remotes := getAllRemotes()
			remoteNames := []string{}
			for _, remote := range remotes {
				remoteNames = append(remoteNames,
					fmt.Sprintf("%s\t%s", remote.Name, remote.RemoteBaseUrl),
				)
			}
			return remoteNames, cobra.ShellCompDirectiveNoFileComp
		},
	}
	remoteSetCmd.Flags().StringVar(&setSyncPolicy, "sync-policy", "", "sync policy for the remote (never, always, hourly, daily, weekly, monthly)")
	remoteSetCmd.RegisterFlagCompletionFunc("sync-policy", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return config.ValidSyncPolicies(), cobra.ShellCompDirectiveNoFileComp
	})

	remoteCmd.AddCommand(remoteAddCmd)
	remoteCmd.AddCommand(remoteListCmd)
	remoteCmd.AddCommand(remoteDeleteCmd)
	remoteCmd.AddCommand(remoteSetCmd)
	rootCmd.AddCommand(remoteCmd)
}

func getAllRemotes() []config.ExtraRemote {
	allRemoteNames := []config.ExtraRemote{
		{
			Name:          "default",
			RemoteBaseUrl: viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY),
			RepositoryDir: viper.GetString(config.LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY),
			SyncPolicy:    backend.SYNC_POLICY_ALWAYS,
		},
	}
	remotes, _ := config.Remotes()

	for _, remote := range remotes {
		allRemoteNames = append(allRemoteNames, remote)
	}
	return allRemoteNames
}
