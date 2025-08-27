package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/criteo/command-launcher/cmd/metrics"
	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/config"
	ctx "github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/frontend"
	"github.com/criteo/command-launcher/internal/repository"
	"github.com/criteo/command-launcher/internal/updater"
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

	selfUpdater updater.SelfUpdater
	cmdUpdaters []*updater.CmdUpdater
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

	rootCtxt.cmdUpdaters = make([]*updater.CmdUpdater, 0)

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
		for _, updater := range rootCtxt.cmdUpdaters {
			updater.CheckUpdateAsync()
		}
	}

	graphite := metrics.NewGraphiteMetricsCollector(viper.GetString(config.METRIC_GRAPHITE_HOST_KEY))
	extensible := metrics.NewExtensibleMetricsCollector(
		rootCtxt.backend.SystemCommand(repository.SYSTEM_METRICS_COMMAND),
	)
	rootCtxt.metrics = metrics.NewCompositeMetricsCollector(graphite, extensible)
	repo, pkg, group, name := cmdAndSubCmd(cmd)
	rootCtxt.metrics.Collect(rootCtxt.user.Partition, repo, pkg, group, name)
}

func postRun(cmd *cobra.Command, args []string) {
	if cmdUpdateEnabled(cmd, args) {
		for _, updater := range rootCtxt.cmdUpdaters {
			updater.Update()
		}
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
		Policy:            config.SelfUpdatePolicy(viper.GetString(config.SELF_UPDATE_POLICY_KEY)),
	}
}

func initCmdUpdater() {
	for _, source := range rootCtxt.backend.AllPackageSources() {
		updater := source.InitUpdater(
			&rootCtxt.user,
			viper.GetDuration(config.SELF_UPDATE_TIMEOUT_KEY),
			viper.GetBool(config.CI_ENABLED_KEY),
			viper.GetString(config.PACKAGE_LOCK_FILE_KEY),
			viper.GetBool(config.VERIFY_PACKAGE_CHECKSUM_KEY),
			viper.GetBool(config.VERIFY_PACKAGE_SIGNATURE_KEY),
		)
		if updater != nil {
			rootCtxt.cmdUpdaters = append(rootCtxt.cmdUpdaters, updater)
		}
	}
}

func initBackend() {
	remotes, _ := config.Remotes()
	extraSources := []*backend.PackageSource{}
	for _, remote := range remotes {
		extraSources = append(extraSources, backend.NewManagedSource(
			remote.Name,
			remote.RepositoryDir,
			remote.RemoteBaseUrl,
			remote.SyncPolicy,
		))
	}

	rootCtxt.backend, _ = backend.NewDefaultBackend(
		config.AppDir(),
		backend.NewDropinSource(viper.GetString(config.DROPIN_FOLDER_KEY)),
		backend.NewManagedSource(
			"default",
			viper.GetString(config.LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY),
			viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY),
			backend.SYNC_POLICY_ALWAYS,
		),
		extraSources...,
	)

	toBeInitiated := []*backend.PackageSource{}
	for _, s := range rootCtxt.backend.AllPackageSources() {
		if s.SyncPolicy != backend.SYNC_POLICY_NEVER && !s.IsInstalled() {
			toBeInitiated = append(toBeInitiated, s)
		}
	}
	if len(toBeInitiated) > 0 {
		log.Info("Initialization...")
		for _, s := range toBeInitiated {
			s.InitialInstallCommands(&rootCtxt.user,
				viper.GetBool(config.CI_ENABLED_KEY),
				viper.GetString(config.PACKAGE_LOCK_FILE_KEY),
				viper.GetBool(config.VERIFY_PACKAGE_CHECKSUM_KEY),
				viper.GetBool(config.VERIFY_PACKAGE_SIGNATURE_KEY),
			)
		}
		rootCtxt.backend.Reload()
	}
}

func initFrontend() {
	frontend := frontend.NewDefaultFrontend(rootCtxt.appCtx, rootCmd, rootCtxt.backend)
	rootCtxt.frontend = frontend

	frontend.AddUserCommands()
}

// return the repo, the package, the group, and the name of the command
func cmdAndSubCmd(cmd *cobra.Command) (string, string, string, string) {
	chain := []string{}

	parent := cmd
	for parent != nil {
		//prepend
		chain = append([]string{parent.Name()}, chain...)
		parent = parent.Parent()
	}

	group := ""
	name := ""

	if len(chain) >= 3 {
		group, name = chain[1], chain[2]
	} else if len(chain) == 2 {
		group, name = "", chain[1]
	}

	// get the internal command from the group and name
	iCmd, err := rootCtxt.backend.FindCommand(group, name)
	if err != nil {
		return "default", "default", "default", "default"
	}

	if group == "" {
		return iCmd.RepositoryID(), iCmd.PackageName(), "default", iCmd.Name()
	}

	return iCmd.RepositoryID(), iCmd.PackageName(), iCmd.Group(), iCmd.Name()
}

func addBuiltinCommands() {
	AddVersionCmd(rootCmd, rootCtxt.appCtx)
	AddConfigCmd(rootCmd, rootCtxt.appCtx)
	AddLoginCmd(rootCmd, rootCtxt.appCtx, rootCtxt.backend.SystemCommand(repository.SYSTEM_LOGIN_COMMAND))
	AddUpdateCmd(rootCmd, rootCtxt.appCtx, rootCtxt.backend.DefaultRepository(), rootCtxt.backend.ExtraPackageSources()...)
	AddCompletionCmd(rootCmd, rootCtxt.appCtx)
	AddPackageCmd(rootCmd, rootCtxt.appCtx)
	AddRenameCmd(rootCmd, rootCtxt.appCtx, rootCtxt.backend)
	AddRemoteCmd(rootCmd, rootCtxt.appCtx, rootCtxt.backend)
}
