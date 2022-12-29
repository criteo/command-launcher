package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/criteo/command-launcher/cmd/metrics"
	"github.com/criteo/command-launcher/cmd/updater"
	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	ctx "github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/frontend"
	"github.com/criteo/command-launcher/internal/remote"
	"github.com/criteo/command-launcher/internal/repository"
	"github.com/criteo/command-launcher/internal/user"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	EXECUTABLE_NOT_DEFINED = "Executable not defined"
)

type rootContext struct {
	appCtx   ctx.LauncherContext
	frontend frontend.Frontend
	backend  backend.Backend

	localRepo  repository.PackageRepository
	dropinRepo repository.PackageRepository

	selfUpdater updater.SelfUpdater
	cmdUpdater  updater.CmdUpdater
	user        user.User
	metrics     metrics.Metrics
}

var (
	rootCmd  *cobra.Command
	rootCtxt = rootContext{}
)

func InitCommands(appName string, appLongName string, version string, buildNum string) {
	rootCmd = createRootCmd(appName, appLongName)
	initApp(appName, version, buildNum)
}

func createRootCmd(appName string, appLongName string) *cobra.Command {
	return &cobra.Command{
		Use:   appName,
		Short: fmt.Sprintf("%s - A command launcher 🚀 made with <3", appLongName),
		Long: fmt.Sprintf(`
%s - A command launcher 🚀 made with <3

Happy Coding!

Example:
  %s --help
`, appLongName, appName),
		PersistentPreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
		PersistentPostRun: postRun,
		SilenceUsage:      true,
	}
}

func initApp(appName string, appVersion string, buildNum string) {
	log.SetLevel(log.FatalLevel)
	rootCtxt.appCtx = ctx.InitContext(appName, appVersion, buildNum)
	config.LoadConfig(rootCtxt.appCtx)
	config.InitLog(rootCtxt.appCtx.AppName())

	initUser()
	initBackend()
	addBuiltinCommands()
	initFrontend()
}

// We have to add the ctrl+C
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(frontend.RootExitCode)
}

func preRun(cmd *cobra.Command, args []string) {
	if selfUpdateEnabled(cmd, args) {
		initSelfUpdater()
		rootCtxt.selfUpdater.CheckUpdateAsync()
	}

	if cmdUpdateEnabled(cmd, args) {
		initCmdUpdater()
		rootCtxt.cmdUpdater.CheckUpdateAsync()
	}

	graphite := metrics.NewGraphiteMetricsCollector(viper.GetString(config.METRIC_GRAPHITE_HOST_KEY))
	extensible := metrics.NewExtensibleMetricsCollector(
		getSystemCommand(repository.SYSTEM_METRICS_COMMAND),
	)
	rootCtxt.metrics = metrics.NewCompositeMetricsCollector(graphite, extensible)
	subcmd, subsubcmd := cmdAndSubCmd(cmd)
	rootCtxt.metrics.Collect(rootCtxt.user.Partition, subcmd, subsubcmd)
}

func postRun(cmd *cobra.Command, args []string) {
	if cmdUpdateEnabled(cmd, args) {
		rootCtxt.cmdUpdater.Update()
	}

	if selfUpdateEnabled(cmd, args) {
		rootCtxt.selfUpdater.Update()
	}

	if metricsEnabled(cmd, args) {
		err := rootCtxt.metrics.Send(frontend.RootExitCode, cmd.Context().Err())
		if err != nil {
			log.Errorln("Metrics usage ♾️ sending has failed")
		}
		log.Debug("Successfully send metrics")
	}
}

func isUpdatePossible(cmd *cobra.Command) bool {
	cmdPath := cmd.CommandPath()
	cmdPath = strings.TrimSpace(strings.TrimPrefix(cmdPath, rootCtxt.appCtx.AppName()))
	// exclude commands for update check
	// for example version command, you don't want to check new update when requesting current version
	for _, w := range []string{"version", "config", "completion", "help", "update", "__complete"} {
		if strings.HasPrefix(cmdPath, w) {
			return false
		}
	}

	return true
}

func selfUpdateEnabled(cmd *cobra.Command, args []string) bool {
	return viper.GetBool(config.SELF_UPDATE_ENABLED_KEY) && isUpdatePossible(cmd)
}

func cmdUpdateEnabled(cmd *cobra.Command, args []string) bool {
	return viper.GetBool(config.COMMAND_UPDATE_ENABLED_KEY) && isUpdatePossible(cmd)
}

func metricsEnabled(cmd *cobra.Command, args []string) bool {
	return viper.GetBool(config.USAGE_METRICS_ENABLED_KEY) && isUpdatePossible(cmd)
}

func initUser() {
	var err error = nil
	rootCtxt.user, err = user.GetUser()
	if err != nil {
		log.Errorln(err)
	}
	log.Infof("User ID: %s User Partition: %d", rootCtxt.user.UID, rootCtxt.user.Partition)
}

func initSelfUpdater() {
	rootCtxt.selfUpdater = updater.SelfUpdater{
		BinaryName:        rootCtxt.appCtx.AppName(),
		LatestVersionUrl:  viper.GetString(config.SELF_UPDATE_LATEST_VERSION_URL_KEY),
		SelfUpdateRootUrl: viper.GetString(config.SELF_UPDATE_BASE_URL_KEY),
		User:              rootCtxt.user,
		CurrentVersion:    rootCtxt.appCtx.AppVersion(),
		Timeout:           viper.GetDuration(config.SELF_UPDATE_TIMEOUT_KEY),
	}
}

func initCmdUpdater() {
	rootCtxt.cmdUpdater = updater.CmdUpdater{
		LocalRepo:            rootCtxt.localRepo,
		CmdRepositoryBaseUrl: viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY),
		User:                 rootCtxt.user,
		Timeout:              viper.GetDuration(config.SELF_UPDATE_TIMEOUT_KEY),
		EnableCI:             viper.GetBool(config.CI_ENABLED_KEY),
		PackageLockFile:      viper.GetString(config.PACKAGE_LOCK_FILE_KEY),
		VerifyChecksum:       viper.GetBool(config.VERIFY_PACKAGE_CHECKSUM_KEY),
		VerifySignature:      viper.GetBool(config.VERIFY_PACKAGE_SIGNATURE_KEY),
	}
}

func initBackend() repository.PackageRepository {
	rootCtxt.backend, _ = backend.NewDefaultBackend(
		config.AppDir(),
		viper.GetString(config.DROPIN_FOLDER_KEY),
		viper.GetString(config.LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY),
	)
	rootCtxt.localRepo = rootCtxt.backend.DefaultRepository()
	rootCtxt.dropinRepo = rootCtxt.backend.DropinRepository()

	installed := rootCtxt.localRepo.InstalledPackages()
	if len(installed) == 0 {
		log.Info("Initialization...")
		installCommands(rootCtxt.localRepo)
		// reload again
		rootCtxt.backend.Reload()
	}

	return rootCtxt.localRepo
}

func initFrontend() {
	frontend := frontend.NewDefaultFrontend(rootCtxt.appCtx, rootCmd, rootCtxt.backend)
	rootCtxt.frontend = frontend

	frontend.AddUserCommands()
}

func cmdAndSubCmd(cmd *cobra.Command) (string, string) {
	chain := []string{}

	parent := cmd
	for parent != nil {
		//prepend
		chain = append([]string{parent.Name()}, chain...)
		parent = parent.Parent()
	}

	if len(chain) >= 3 {
		return chain[1], chain[2]
	} else if len(chain) == 2 {
		return chain[1], "default"
	}
	return "default", "default"
}

func installCommands(repo repository.PackageRepository) error {
	remote := remote.CreateRemoteRepository(viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY))
	errors := make([]string, 0)

	// check locked packages if ci is enabled
	lockedPackages := map[string]string{}
	if viper.GetBool(config.CI_ENABLED_KEY) {
		pkgs, err := rootCtxt.cmdUpdater.LoadLockedPackages(viper.GetString(config.PACKAGE_LOCK_FILE_KEY))
		if err == nil {
			lockedPackages = pkgs
		}
	}

	if pkgs, err := remote.PackageNames(); err == nil {
		for _, pkgName := range pkgs {
			pkgVersion := "unspecified"
			if lockedVersion, ok := lockedPackages[pkgName]; ok {
				pkgVersion = lockedVersion
			} else {
				latest, err := remote.LatestPackageInfo(pkgName)
				if err != nil {
					log.Error(err)
					errors = append(errors, fmt.Sprintf("cannot get the latest version of the package %s: %v", latest.Name, err))
					continue
				}
				if !rootCtxt.user.InPartition(latest.StartPartition, latest.EndPartition) {
					log.Infof("Skip installing package %s, user not in partition (%d %d)\n", latest.Name, latest.StartPartition, latest.EndPartition)
					continue
				}
				pkgVersion = latest.Version
			}

			pkg, err := remote.Package(pkgName, pkgVersion)
			if err != nil {
				log.Error(err)
				errors = append(errors, fmt.Sprintf("cannot get the package %s: %v", pkgName, err))
				continue
			}
			if ok, err := remote.Verify(pkg,
				viper.GetBool(config.VERIFY_PACKAGE_CHECKSUM_KEY),
				viper.GetBool(config.VERIFY_PACKAGE_SIGNATURE_KEY),
			); !ok || err != nil {
				log.Error(err)
				errors = append(errors, fmt.Sprintf("failed to verify package %s, skip it: %v", pkgName, err))
				continue
			}
			err = repo.Install(pkg)
			if err != nil {
				errors = append(errors, fmt.Sprintf("cannot install the package %s: %v", pkgName, err))
				continue
			}
		}
	} else {
		errors = append(errors, fmt.Sprintf("cannot get remote packages: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("install failed for the following reasons: [%s]", strings.Join(errors, ", "))
	}

	return nil
}

func addBuiltinCommands() {
	AddVersionCmd(rootCmd, rootCtxt.appCtx)
	AddConfigCmd(rootCmd, rootCtxt.appCtx)
	AddLoginCmd(rootCmd, rootCtxt.appCtx, getSystemCommand(repository.SYSTEM_LOGIN_COMMAND))
	AddUpdateCmd(rootCmd, rootCtxt.appCtx, rootCtxt.localRepo)
	AddCompletionCmd(rootCmd, rootCtxt.appCtx)
	AddPackageCmd(rootCmd, rootCtxt.appCtx)
}

func getSystemCommand(name string) command.Command {
	sysCmds := rootCtxt.localRepo.InstalledSystemCommands()
	switch name {
	case repository.SYSTEM_LOGIN_COMMAND:
		return sysCmds.Login
	case repository.SYSTEM_METRICS_COMMAND:
		return sysCmds.Metrics
	}
	return nil
}
