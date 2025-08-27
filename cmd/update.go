package cmd

import (
	"fmt"
	"time"

	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/repository"
	"github.com/criteo/command-launcher/internal/updater"
	"github.com/criteo/command-launcher/internal/user"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type UpdateFlags struct {
	Package bool
	Self    bool
	Timeout time.Duration
}

var (
	updateFlags = UpdateFlags{}
)

func AddUpdateCmd(rootCmd *cobra.Command, appCtx context.LauncherContext, localRepo repository.PackageRepository, extraPackageSources ...*backend.PackageSource) {
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
					Timeout:           updateFlags.Timeout,
					Policy:            config.SelfUpdatePolicy(viper.GetString(config.SELF_UPDATE_POLICY_KEY)),
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
					LocalRepo:            localRepo,
					CmdRepositoryBaseUrl: viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY),
					User:                 u,
					Timeout:              updateFlags.Timeout,
					EnableCI:             enableCI,
					PackageLockFile:      packageLockFile,
					SyncPolicy:           "always", // TODO: use constant instead of string
				}
				cmdUpdater.CheckUpdateAsync()
				err := cmdUpdater.Update()
				if err != nil {
					console.Error(err.Error())
				} else {
					console.Success("packages in 'default' repository are up-to-date")
				}

				// now update the packages in extra remote
				for _, source := range extraPackageSources {
					updater := source.InitUpdater(&u, updateFlags.Timeout, enableCI, packageLockFile, false, false)
					// force sync policy to always, as the intention of running this command is to update the packages
					updater.SyncPolicy = "always"
					updater.CheckUpdateAsync()
					err := updater.Update()
					if err != nil {
						console.Error(err.Error())
					} else {
						console.Success("packages in '%s' repository are up-to-date", source.Name)
					}
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
	updateCmd.Flags().DurationVarP(&updateFlags.Timeout, "timeout", "t", 10*time.Second, "Timeout for update operations")

	rootCmd.AddCommand(updateCmd)
}
