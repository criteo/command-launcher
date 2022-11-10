package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/criteo/command-launcher/cmd/consent"
	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/cmd/repository"
	"github.com/criteo/command-launcher/cmd/updater"
	"github.com/criteo/command-launcher/cmd/user"
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	ctx "github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/criteo/command-launcher/internal/metrics"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	EXECUTABLE_NOT_DEFINED = "Executable not defined"
)

type rootContext struct {
	appCtx      ctx.LauncherContext
	localRepo   repository.PackageRepository
	dropinRepo  repository.PackageRepository
	selfUpdater updater.SelfUpdater
	cmdUpdater  updater.CmdUpdater
	user        user.User
	metrics     metrics.Metrics
}

var (
	rootCmd      *cobra.Command
	rootCtxt     = rootContext{}
	rootExitCode = 0
)

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
	}
}

func initApp() repository.PackageRepository {
	repo, err := repository.CreateLocalRepository(viper.GetString(config.LOCAL_COMMAND_REPOSITORY_DIRNAME_KEY), nil)
	if err != nil {
		log.Fatal(err)
	}

	installed := repo.InstalledPackages()
	if len(installed) == 0 {
		log.Info("Initialization...")
		installCommands(repo)
	}

	rootCtxt.localRepo = repo

	if dropinRepo, err := repository.CreateLocalRepository(viper.GetString(config.DROPIN_FOLDER_KEY), nil); err == nil {
		rootCtxt.dropinRepo = dropinRepo
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
	AddLoginCmd(rootCmd, rootCtxt.appCtx)
	AddUpdateCmd(rootCmd, rootCtxt.appCtx, rootCtxt.localRepo)
	AddCompletionCmd(rootCmd, rootCtxt.appCtx)
}

func addLocalCommands() {
	addCommands(rootCtxt.localRepo.InstalledGroupCommands(), rootCtxt.localRepo.InstalledExecutableCommands())
}

func addDropinCommands() {
	addCommands(rootCtxt.dropinRepo.InstalledGroupCommands(), rootCtxt.dropinRepo.InstalledExecutableCommands())
}

func addCommands(groups []command.Command, executables []command.Command) {
	// first add group commands
	groupCmds := map[string]*cobra.Command{}
	for _, v := range groups {
		group := v.Group()
		name := v.Name()
		usage := strings.TrimSpace(fmt.Sprintf("%s %s",
			v.Name(),
			strings.TrimSpace(strings.Trim(v.ArgsUsage(), v.Name())),
		))
		requiredFlags := v.RequiredFlags()
		requestedResources := v.RequestedResources()
		cmd := &cobra.Command{
			DisableFlagParsing: true,
			Use:                usage,
			Example:            formatExamples(v.Examples()),
			Short:              v.ShortDescription(),
			Long:               v.LongDescription(),
			Run: func(cmd *cobra.Command, args []string) {
				consents, err := consent.GetConsents(group, name, requestedResources, viper.GetBool(config.ENABLE_USER_CONSENT_KEY))
				if err != nil {
					log.Warnf("failed to get user consent: %v", err)
				}
				exitCode, err := executeCommand(group, name, args, []string{}, consents)
				if err != nil && err.Error() == EXECUTABLE_NOT_DEFINED {
					cmd.Help()
				}
				rootExitCode = exitCode
			},
		}
		for _, flag := range requiredFlags {
			addFlagToCmd(cmd, flag)
		}
		groupCmds[v.Name()] = cmd
		rootCmd.AddCommand(cmd)
	}

	// add executable commands
	for _, v := range executables {
		group := v.Group()
		name := v.Name()
		usage := strings.TrimSpace(fmt.Sprintf("%s %s",
			v.Name(),
			strings.TrimSpace(strings.Trim(v.ArgsUsage(), v.Name())),
		))
		requiredFlags := v.RequiredFlags()
		validArgs := v.ValidArgs()
		validArgsCmd := v.ValidArgsCmd()
		checkFlags := v.CheckFlags()
		requestedResources := v.RequestedResources()
		// flagValuesCmd := v.FlagValuesCmd()
		cmd := &cobra.Command{
			DisableFlagParsing: true,
			Use:                usage,
			Example:            formatExamples(v.Examples()),
			Short:              v.ShortDescription(),
			Long:               v.LongDescription(),
			Run: func(c *cobra.Command, args []string) {
				consents, err := consent.GetConsents(group, name, requestedResources, viper.GetBool(config.ENABLE_USER_CONSENT_KEY))
				if err != nil {
					log.Warnf("failed to get user consent: %v", err)
				}

				envVars, code, shouldQuit := parseArgsToEnvVars(c, args, checkFlags)
				if shouldQuit {
					rootExitCode = code
					return
				}

				// TODO: in order to support flag value auto completion, we need to set DisableFlagParsing: false
				// when setting disable flagParsing to false, the parent command will parse the flags
				// so the args pass to the subcommand will not include the flags
				// we need to restore the flags into args here
				// considering the complexity here, we will cover it later
				if exitCode, err := executeCommand(group, name, args, envVars, consents); err != nil {
					rootExitCode = exitCode
				}
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
			addFlagToCmd(cmd, flag)
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
						// the first line starting with # is the control line, it controls the completion behavior when the return body is empty
						shellDirective := cobra.ShellCompDirectiveNoFileComp
						switch strings.TrimSpace(strings.TrimLeft(parts[0], "#")) {
						case "dir-completion-only":
							shellDirective = cobra.ShellCompDirectiveFilterDirs
						case "default":
							shellDirective = cobra.ShellCompDirectiveDefault
						case "no-file-completion":
							shellDirective = cobra.ShellCompDirectiveNoFileComp
						}
						return parts[1:], shellDirective
					}
					return parts, cobra.ShellCompDirectiveNoFileComp
				}
			}
			if len(validArgs) > 0 {
				return validArgs, cobra.ShellCompDirectiveNoFileComp
			}
			return []string{}, cobra.ShellCompDirectiveDefault
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

// parse args and inject environment vars
// return the environment vars, and if we should exit
func parseArgsToEnvVars(c *cobra.Command, args []string, checkFlags bool) ([]string, int, bool) {
	var envVars []string = []string{}
	var envTable map[string]string = map[string]string{}

	log.Debugf("checkFlags=%t", checkFlags)
	if checkFlags {
		var err error = nil
		envVarPrefix := strings.ToUpper(rootCtxt.appCtx.AppName())
		envVars, envTable, err = parseCmdArgsToEnv(c, args, envVarPrefix)
		if err != nil {
			console.Error("Failed to parse arguments: %v", err)
			// set exit code to 1, and should quit
			return envVars, 1, true
		}
		if h, exist := envTable[fmt.Sprintf("%s_FLAG_HELP", envVarPrefix)]; exist && h == "true" {
			c.Help()
			// show help and should quit
			return envVars, 0, true
		}
	}
	log.Debugf("flag & args environments: %v", envVars)

	return envVars, 0, false
}

func formatExamples(examples []command.ExampleEntry) string {
	if examples == nil || len(examples) == 0 {
		return ""
	}

	output := []string{}

	for _, v := range examples {
		output = append(output, fmt.Sprintf(`  # %s
  %s
`, v.Scenario, v.Command))
	}

	return strings.Join(output, "\n")
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
func executeCommand(group, name string, args []string, initialEnvCtx []string, consent []string) (int, error) {
	iCmd, err := getExecutableCommand(group, name)
	if err != nil {
		return 1, err
	}
	if iCmd.Executable() == "" {
		return 1, errors.New(EXECUTABLE_NOT_DEFINED)
	}

	envCtx := getCmdEnvContext(initialEnvCtx, consent)
	exitCode, err := iCmd.Execute(envCtx, args...)
	if err != nil {
		return exitCode, err
	}

	return 0, nil
}

// execute the valid args command of the cdt command
func executeValidArgsOfCommand(group, name string, args []string) (string, error) {
	iCmd, err := getExecutableCommand(group, name)
	if err != nil {
		return "", err
	}

	envCtx := getCmdEnvContext([]string{}, []string{})

	_, output, err := iCmd.ExecuteValidArgsCmd(envCtx, args...)
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

	envCtx := getCmdEnvContext([]string{}, []string{})

	_, output, err := iCmd.ExecuteFlagValuesCmd(envCtx, args...)
	if err != nil {
		return "", err
	}

	return output, nil
}

func addFlagToCmd(cmd *cobra.Command, flag string) {
	flagName, flagShort, flagDesc, flagType, defaultValue := parseFlagDefinition(flag)
	switch flagType {
	case "bool":
		// always use false as the default for the bool type
		cmd.Flags().BoolP(flagName, flagShort, false, flagDesc)
	default:
		cmd.Flags().StringP(flagName, flagShort, defaultValue, flagDesc)
	}
}

func parseFlagDefinition(line string) (string, string, string, string, string) {
	flagParts := strings.Split(line, "\t")
	name := strings.TrimSpace(flagParts[0])
	short := ""
	description := ""
	flagType := "string"
	defaultValue := ""
	if len(flagParts) == 2 {
		description = strings.TrimSpace(flagParts[1])
	}
	if len(flagParts) > 2 {
		short = strings.TrimSpace(flagParts[1])
		description = strings.TrimSpace(flagParts[2])
	}
	if len(flagParts) > 3 {
		flagType = strings.TrimSpace(flagParts[3])
	}
	if len(flagParts) > 4 {
		defaultValue = strings.TrimSpace(flagParts[4])
	}

	return name, short, description, flagType, defaultValue
}

func getCmdEnvContext(envVars []string, consents []string) []string {
	vars := append([]string{}, envVars...)

	for _, item := range consents {
		switch item {
		case consent.USERNAME:
			username, err := helper.GetUsername()
			if err != nil {
				username = ""
			}
			if username != "" {
				vars = append(vars, fmt.Sprintf("%s=%s", rootCtxt.appCtx.UsernameEnvVar(), username))
			}
		case consent.PASSWORD:
			password, err := helper.GetPassword()
			if err != nil {
				password = ""
			}
			if password != "" {
				vars = append(vars, fmt.Sprintf("%s=%s", rootCtxt.appCtx.PasswordEnvVar(), password))
			}
		case consent.LOGIN_TOKEN:
			// TODO: add login token
		case consent.LOG_LEVEL:
			// append log level from configuration
			logLevel := viper.GetString(config.LOG_LEVEL_KEY)
			vars = append(vars, fmt.Sprintf("%s=%s",
				rootCtxt.appCtx.LogLevelEnvVar(),
				logLevel,
			))
		case consent.DEBUG_FLAGS:
			// append debug flags from configuration
			debugFlags := os.Getenv(rootCtxt.appCtx.DebugFlagsEnvVar())
			vars = append(vars, fmt.Sprintf("%s=%s,%s",
				rootCtxt.appCtx.DebugFlagsEnvVar(),
				debugFlags,
				viper.GetString(config.DEBUG_FLAGS_KEY),
			))
		}
	}

	// Enable variable with prefix [binary_name] and COLA
	// TODO: remove it when in version 1.8 all variables are migrated to COLA prefix
	outputVars := []string{}
	for _, v := range vars {
		prefix := fmt.Sprintf("%s_", strings.ToUpper(rootCtxt.appCtx.AppName()))
		if strings.HasPrefix(v, prefix) && prefix != "COLA_" {
			outputVars = append(outputVars, strings.Replace(v, prefix, "COLA_", 1))
		}
		outputVars = append(outputVars, v)
	}

	return outputVars
}

func parseCmdArgsToEnv(c *cobra.Command, args []string, envVarPrefix string) ([]string, map[string]string, error) {
	envVars := []string{}
	envTable := map[string]string{}
	if err := c.LocalFlags().Parse(args); err != nil {
		return envVars, envTable, err
	}
	c.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		n := strings.ReplaceAll(strings.ToUpper(flag.Name), "-", "_")
		v := flag.Value.String()
		k := fmt.Sprintf("%s_FLAG_%s", envVarPrefix, n)
		envVars = append(envVars,
			fmt.Sprintf(
				"%s=%s",
				k, v,
			),
		)
		envTable[k] = v
	})
	for idx, arg := range c.LocalFlags().Args() {
		k := fmt.Sprintf("%s_ARG_%s", envVarPrefix, strconv.Itoa(idx+1))
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, arg))
		envTable[k] = arg
	}
	return envVars, envTable, nil
}

func initContext(appName string, appVersion string, buildNum string) {
	log.SetLevel(log.FatalLevel)
	rootCtxt.appCtx = ctx.InitContext(appName, appVersion, buildNum)
	config.LoadConfig(rootCtxt.appCtx)
	config.InitLog(rootCtxt.appCtx.AppName())

	initUser()
	initApp()
	addBuiltinCommands()
	addLocalCommands()
	addDropinCommands()
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

func InitCommands(appName string, appLongName string, version string, buildNum string) {
	rootCmd = createRootCmd(appName, appLongName)
	initContext(appName, version, buildNum)
}

// We have to add the ctrl+C
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(rootExitCode)
}
