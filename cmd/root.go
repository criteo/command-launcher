package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/criteo/command-launcher/cmd/dropin"
	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/cmd/repository"
	"github.com/criteo/command-launcher/cmd/updater"
	"github.com/criteo/command-launcher/cmd/user"
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/criteo/command-launcher/internal/metrics"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	EXECUTABLE_NOT_DEFINED = "Executable not defined"
)

type rootContext struct {
	localRepo   repository.PackageRepository
	dropinRepo  dropin.DropinRepository
	selfUpdater updater.SelfUpdater
	cmdUpdater  updater.CmdUpdater
	user        user.User
	metrics     metrics.Metrics
}

const (
	BINARY_NAME = "cdt"
)

var (
	BuildNum string
	rootCtxt = rootContext{}

	rootCmd = &cobra.Command{
		Use:   BINARY_NAME,
		Short: "Criteo Dev Toolkit - A command launcher 🚀 made with <3",
		Long: `
Criteo Dev Toolkit - A command launcher 🚀 made with <3

Happy Coding!

Example:
  cdt hotfix
  cdt --help
`,
		PersistentPreRun: preRun,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			}
		},
		PersistentPostRun: postRun,
	}
)

func init() {
	log.SetLevel(log.FatalLevel)
	config.LoadConfig()
	initUser()
	initApp()
	addLocalCommands()
	addDropinCommands()
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

	rootCtxt.metrics = metrics.NewMetricsCollector(viper.GetString(config.METRIC_GRAPHITE_HOST_KEY))
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
		err := rootCtxt.metrics.Send(cmd.Context().Err())
		if err != nil {
			log.Errorln("Metrics usage ♾️ sending has failed")
		}
	}
}

func selfUpdateEnabled(cmd *cobra.Command, args []string) bool {
	if !viper.GetBool(config.SELF_UPDATE_ENABLED_KEY) {
		return false
	}

	cmdPath := cmd.CommandPath()
	cmdPath = strings.TrimSpace(strings.TrimPrefix(cmdPath, BINARY_NAME))
	// exclude commands for update check
	// for example version command, you don't want to check new update when requesting current version
	for _, w := range []string{"version", "config", "completion", "help", "__complete"} {
		if strings.HasPrefix(cmdPath, w) {
			return false
		}
	}
	return true
}

func cmdUpdateEnabled(cmd *cobra.Command, args []string) bool {
	cmdPath := cmd.CommandPath()
	cmdPath = strings.TrimSpace(strings.TrimPrefix(cmdPath, BINARY_NAME))
	for _, w := range []string{"version", "config", "completion", "help", "__complete"} {
		if strings.HasPrefix(cmdPath, w) {
			return false
		}
	}
	return true
}

func metricsEnabled(cmd *cobra.Command, args []string) bool {
	if !viper.GetBool(config.USAGE_METRICS_ENABLED_KEY) {
		return false
	}
	cmdPath := cmd.CommandPath()
	cmdPath = strings.TrimSpace(strings.TrimPrefix(cmdPath, BINARY_NAME))
	for _, w := range []string{"version", "config", "completion", "help", "__complete"} {
		if strings.HasPrefix(cmdPath, w) {
			return false
		}
	}
	return true
}

func initUser() {
	var err error = nil
	rootCtxt.user, err = user.GetUser()
	if err != nil {
		log.Errorln(err)
	}
	log.Infof("User ID: %s User Partition: %d\n", rootCtxt.user.UID, rootCtxt.user.Partition)
}

func initSelfUpdater() {
	rootCtxt.selfUpdater = updater.SelfUpdater{
		BinaryName:        BINARY_NAME,
		LatestVersionUrl:  viper.GetString(config.SELF_UPDATE_LATEST_VERSION_URL_KEY),
		SelfUpdateRootUrl: viper.GetString(config.SELF_UPDATE_BASE_URL_KEY),
		User:              rootCtxt.user,
		CurrentVersion:    BuildNum,
		Timeout:           viper.GetDuration(config.SELF_UPDATE_TIMEOUT_KEY),
	}
}

func initCmdUpdater() {
	rootCtxt.cmdUpdater = updater.CmdUpdater{
		LocalRepo:            rootCtxt.localRepo,
		CmdRepositoryBaseUrl: viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY),
		User:                 rootCtxt.user,
		Timeout:              viper.GetDuration(config.SELF_UPDATE_TIMEOUT_KEY),
	}
}

func initApp() repository.PackageRepository {
	config.InitLog("cdt")

	repo, err := repository.CreateLocalRepository(viper.GetString(config.LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY))
	if err != nil {
		log.Fatal(err)
	}

	installed := repo.InstalledPackages()
	if len(installed) == 0 {
		log.Info("Initialization...")
		installCommands(repo)
	}

	rootCtxt.localRepo = repo

	if dropinRepo, err := dropin.Load(viper.GetString(config.DROPIN_FOLDER_KEY)); err == nil {
		rootCtxt.dropinRepo = *dropinRepo
	}

	return repo
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
	if pkgs, err := remote.PackageNames(); err == nil {
		for _, pkgName := range pkgs {
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
			pkg, err := remote.Package(latest.Name, latest.Version)
			if err != nil {
				log.Error(err)
				errors = append(errors, fmt.Sprintf("cannot get the package %s: %v", latest.Name, err))
				continue
			}
			err = repo.Install(pkg)
			if err != nil {
				errors = append(errors, fmt.Sprintf("cannot install the package %s: %v", latest.Name, err))
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

func addLocalCommands() {
	addCommands(rootCtxt.localRepo.InstalledGroupCommands(), rootCtxt.localRepo.InstalledExecutableCommands())
}

func addDropinCommands() {
	addCommands(rootCtxt.dropinRepo.GroupCommands(), rootCtxt.dropinRepo.ExecutableCommands())
}

func addCommands(groups []command.Command, executables []command.Command) {
	// first add group commands
	groupCmds := map[string]*cobra.Command{}
	for _, v := range groups {
		group := v.Group()
		name := v.Name()
		requiredFlags := v.RequiredFlags()
		cmd := &cobra.Command{
			DisableFlagParsing: true,
			Use:                v.Name(),
			Short:              v.ShortDescription(),
			Long:               v.LongDescription(),
			RunE: func(cmd *cobra.Command, args []string) error {
				err := executeCommand(group, name, args)
				if err != nil && err.Error() == EXECUTABLE_NOT_DEFINED {
					cmd.Help()
					return nil
				}
				return err
			},
		}
		for _, flag := range requiredFlags {
			flagName, flagShort, flagDesc := parseFlagDefinition(flag)
			cmd.Flags().StringP(flagName, flagShort, "", flagDesc)
		}
		groupCmds[v.Name()] = cmd
		rootCmd.AddCommand(cmd)
	}

	// add executable commands
	for _, v := range executables {
		group := v.Group()
		name := v.Name()
		requiredFlags := v.RequiredFlags()
		validArgs := v.ValidArgs()
		validArgsCmd := v.ValidArgsCmd()
		// flagValuesCmd := v.FlagValuesCmd()
		cmd := &cobra.Command{
			DisableFlagParsing: true,
			Use:                v.Name(),
			Short:              v.ShortDescription(),
			Long:               v.LongDescription(),
			RunE: func(c *cobra.Command, args []string) error {
				// TODO: in order to support flag value auto completion, we need to set DisableFlagParsing: false
				// when setting disable flagParsing to false, the parent command will parse the flags
				// so the args pass to the subcommand will not include the flags
				// we need to restore the flags into args here
				// considering the complexity here, we will cover it later
				return executeCommand(group, name, args)
			},
		}

		// TODO: uncomment to enable the flag value auto-completion
		/*
			cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
				err := executeCommand(group, name, args)
				if err != nil {
					c.Help()
				}
			})
		*/

		for _, flag := range requiredFlags {
			flagName, flagShort, flagDesc := parseFlagDefinition(flag)
			cmd.Flags().StringP(flagName, flagShort, "", flagDesc)
			// TODO: enable flag parsing in cdt command to enable the flag value auto-completion.
			// for now comment out this code as it will impact the flag parsing for the subcommand
			// need to work it later
			/*
				cmd.RegisterFlagCompletionFunc(flagName, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
					// call external command for flag value completon
					if len(flagValuesCmd) > 0 {
						flagValuesCmdArgs := append([]string{flagName}, args...)
						output, err := executeFlagValuesOfCommand(group, name, flagValuesCmdArgs)
						if err != nil {
							return []string{}, cobra.ShellCompDirectiveNoFileComp
						}
						parts := strings.Split(output, "\n")
						if len(parts) > 0 {
							if strings.HasPrefix(parts[0], "#") { // skip the first control line, for further controls
								return parts[1:], cobra.ShellCompDirectiveNoFileComp
							}
							return parts, cobra.ShellCompDirectiveNoFileComp
						}
					}
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				})
			*/
		}

		cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(validArgsCmd) > 0 {
				output, err := executeValidArgsOfCommand(group, name, args)
				if err != nil {
					return []string{}, cobra.ShellCompDirectiveNoFileComp
				}
				parts := strings.Split(output, "\n")
				if len(parts) > 0 {
					if strings.HasPrefix(parts[0], "#") { // skip the first control line, for further controls
						return parts[1:], cobra.ShellCompDirectiveNoFileComp
					}
					return parts, cobra.ShellCompDirectiveNoFileComp
				}
			}
			if len(validArgs) > 0 {
				return validArgs, cobra.ShellCompDirectiveNoFileComp
			}
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		}

		if v.Group() == "" {
			rootCmd.AddCommand(cmd)
		} else {
			if group, exists := groupCmds[v.Group()]; exists {
				group.AddCommand(cmd)
			} else {
				log.Errorf("cannot install cmd %s in group %s: group not found", v.Name(), v.Group())
			}
		}

	}
}

func getExecutableCommand(group, name string) (command.Command, error) {
	/* first check dropin repository, if not found, check the local repo
	this will allow the dropin command override remote version for testing
	*/
	iCmd, err := rootCtxt.dropinRepo.Command(group, name)
	if err != nil {
		return rootCtxt.localRepo.Command(group, name)
	}
	return iCmd, err
}

// execute a cdt command
func executeCommand(group, name string, args []string) error {
	iCmd, err := getExecutableCommand(group, name)
	if err != nil {
		return err
	}
	if iCmd.Executable() == "" {
		return errors.New(EXECUTABLE_NOT_DEFINED)
	}
	secrets := defaultSecrets()
	_, err = iCmd.Execute(secrets, args...)
	if err != nil {
		return err
	}
	return nil
}

// execute the valid args command of the cdt command
func executeValidArgsOfCommand(group, name string, args []string) (string, error) {
	iCmd, err := getExecutableCommand(group, name)
	if err != nil {
		return "", err
	}
	secrets := defaultSecrets()
	_, output, err := iCmd.ExecuteValidArgsCmd(secrets, args...)
	if err != nil {
		return "", err
	}
	return output, nil
}

// execute the flag values command of the cdt command
func executeFlagValuesOfCommand(group, name string, args []string) (string, error) {
	iCmd, err := getExecutableCommand(group, name)
	if err != nil {
		return "", err
	}
	secrets := defaultSecrets()
	_, output, err := iCmd.ExecuteFlagValuesCmd(secrets, args...)
	if err != nil {
		return "", err
	}
	return output, nil
}

func parseFlagDefinition(line string) (string, string, string) {
	flagParts := strings.Split(line, "\t")
	name := strings.TrimSpace(flagParts[0])
	short := ""
	description := ""
	if len(flagParts) == 2 {
		description = strings.TrimSpace(flagParts[1])
	}
	if len(flagParts) >= 3 {
		short = strings.TrimSpace(flagParts[1])
		description = strings.TrimSpace(flagParts[2])
	}
	return name, short, description
}

func defaultSecrets() []string {
	cdtVars := []string{}
	// user credential
	username, err := helper.GetSecret("cdt-username")
	if err != nil {
		username = ""
	}
	password, err := helper.GetSecret("cdt-password")
	if err != nil {
		password = ""
	}
	if username != "" {
		cdtVars = append(cdtVars, fmt.Sprintf("CDT_USERNAME=%s", username))
	}
	if password != "" {
		cdtVars = append(cdtVars, fmt.Sprintf("CDT_PASSWORD=%s", password))
	}
	// append debug flags from configuration
	debugFlags := os.Getenv("CDT_DEBUG_FLAGS")
	cdtVars = append(cdtVars, fmt.Sprintf("CDT_DEBUG_FLAGS=%s,%s",
		debugFlags,
		viper.GetString(config.DEBUG_FLAGS_KEY),
	))

	return cdtVars
}

// We have to add the ctrl+C
func Execute() {
	rootCmd.SilenceUsage = true
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
