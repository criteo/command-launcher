package cmd

import (
	"fmt"

	"github.com/criteo/command-launcher/cmd/updater"
	"github.com/criteo/command-launcher/cmd/user"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/context"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type UpdateFlags struct {
	Package bool
	Self    bool
}

var (
	updateFlags = UpdateFlags{}
)

func AddUpdateCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	appName := appCtx.AppName()
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: fmt.Sprintf("Update %s, or its commands", appName),
		Long: fmt.Sprintf(`
Check the update of %s and its commands.
`, appName),
		Example: fmt.Sprintf(`
  %s update --package
  %s update --self
`, appName, appName),
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := user.GetUser()
			if err != nil {
				log.Errorln(err)
			}

			if updateFlags.Self {
				console.Highlight("checking available %s version ...", appCtx.AppName())
				selfUpdater := updater.SelfUpdater{
					BinaryName:        appCtx.AppName(),
					LatestVersionUrl:  viper.GetString(config.SELF_UPDATE_LATEST_VERSION_URL_KEY),
					SelfUpdateRootUrl: viper.GetString(config.SELF_UPDATE_BASE_URL_KEY),
					User:              u,
					CurrentVersion:    appCtx.AppVersion(),
					Timeout:           viper.GetDuration(config.SELF_UPDATE_TIMEOUT_KEY),
				}
				selfUpdater.CheckUpdateAsync()
				err := selfUpdater.Update()
				if err != nil {
					console.Error(err.Error())
				} else {
					console.Success("%s is up-to-date", appCtx.AppName())
				}
			}

			if updateFlags.Package {
				console.Highlight("checking available package updates ...")
				enableCI := viper.GetBool(config.CI_ENABLED_KEY)
				packageLockFile := viper.GetString(config.PACKAGE_LOCK_FILE_KEY)
				if enableCI {
					fmt.Printf("CI mode enabled, load package lock file: %s\n", packageLockFile)
				}
				cmdUpdater := updater.CmdUpdater{
					LocalRepo:            rootCtxt.localRepo,
					CmdRepositoryBaseUrl: viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY),
					User:                 u,
					Timeout:              viper.GetDuration(config.SELF_UPDATE_TIMEOUT_KEY),
					EnableCI:             enableCI,
					PackageLockFile:      packageLockFile,
				}
				cmdUpdater.CheckUpdateAsync()
				err := cmdUpdater.Update()
				if err != nil {
					console.Error(err.Error())
				} else {
					console.Success("packages are up-to-date")
				}
			}

			if !updateFlags.Package && !updateFlags.Self {
				cmd.Help()
			}

			return nil
		},
	}

	updateCmd.Flags().BoolVarP(&updateFlags.Package, "package", "p", false, "Update packages and commands")
	updateCmd.Flags().BoolVarP(&updateFlags.Self, "self", "s", false, "Self update")

	rootCmd.AddCommand(updateCmd)
}
