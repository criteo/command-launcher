package frontend

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/criteo/command-launcher/cmd/consent"
	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/helper"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	EXECUTABLE_NOT_DEFINED = "Executable not defined"
)

var (
	RootExitCode = 0
)

type defaultFrontend struct {
	rootCmd *cobra.Command

	appCtx  context.LauncherContext
	backend backend.Backend

	groupCmds      map[string]*cobra.Command
	executableCmds map[string]*cobra.Command
}

func NewDefaultFrontend(appCtx context.LauncherContext, rootCmd *cobra.Command, backend backend.Backend) Frontend {
	frontend := &defaultFrontend{
		appCtx:  appCtx,
		rootCmd: rootCmd,
		backend: backend,

		groupCmds:      make(map[string]*cobra.Command),
		executableCmds: make(map[string]*cobra.Command),
	}
	return frontend
}

func (self *defaultFrontend) AddUserCommands() {
	self.addGroupCommands()
	self.addExecutableCommands()
}

func (self *defaultFrontend) addGroupCommands() {
	groups := self.backend.GroupCommands()
	for _, v := range groups {
		group := v.RuntimeGroup()
		name := v.RuntimeName()
		usage := strings.TrimSpace(fmt.Sprintf("%s %s",
			v.RuntimeName(),
			strings.TrimSpace(strings.Trim(v.ArgsUsage(), v.RuntimeName())),
		))
		requiredFlags := v.RequiredFlags()
		requestedResources := v.RequestedResources()
		flags := v.Flags()
		exclusiveFlags := v.ExclusiveFlags()
		groupFlags := v.GroupFlags()
		cmd := &cobra.Command{
			DisableFlagParsing: true, // not enable the checkFlags feature for group command for now
			Use:                usage,
			Example:            formatExamples(v.Examples()),
			Short:              v.ShortDescription(),
			Long:               v.LongDescription(),
			Run: func(cmd *cobra.Command, args []string) {
				consents, err := consent.GetConsents(group, name, requestedResources, viper.GetBool(config.ENABLE_USER_CONSENT_KEY))
				if err != nil {
					log.Warnf("failed to get user consent: %v", err)
				}
				exitCode, err := self.executeCommand(group, name, args, []string{}, consents)
				if err != nil && err.Error() == EXECUTABLE_NOT_DEFINED {
					cmd.Help()
				}
				RootExitCode = exitCode
			},
		}
		// legacy flag definition ("requiredFlags")
		// deprecated
		for _, flag := range requiredFlags {
			addFlagToCmd(cmd, flag)
		}
		// new ways to handle flags
		self.processFlags(group, name, cmd, flags, exclusiveFlags, groupFlags)

		self.groupCmds[v.RuntimeName()] = cmd
		self.rootCmd.AddCommand(cmd)
	}
}

func (self *defaultFrontend) addExecutableCommands() {
	executables := self.backend.ExecutableCommands()
	for _, v := range executables {
		group := v.RuntimeGroup()
		name := v.RuntimeName()
		usage := strings.TrimSpace(fmt.Sprintf("%s %s",
			v.RuntimeName(),
			strings.TrimSpace(strings.Trim(v.ArgsUsage(), v.RuntimeName())),
		))
		requiredFlags := v.RequiredFlags()
		validArgs := v.ValidArgs()
		validArgsCmd := v.ValidArgsCmd()
		checkFlags := v.CheckFlags()
		requestedResources := v.RequestedResources()
		flags := v.Flags()
		exclusiveFlags := v.ExclusiveFlags()
		groupFlags := v.GroupFlags()
		cmd := &cobra.Command{
			DisableFlagParsing: !checkFlags,
			Use:                usage,
			Example:            formatExamples(v.Examples()),
			Short:              v.ShortDescription(),
			Long:               v.LongDescription(),
			Run: func(c *cobra.Command, args []string) {
				consents, err := consent.GetConsents(group, name, requestedResources, viper.GetBool(config.ENABLE_USER_CONSENT_KEY))
				if err != nil {
					log.Warnf("failed to get user consent: %v", err)
				}

				envVars, originalArgs, code, shouldQuit := self.parseArgsToEnvVars(c, args, checkFlags)
				if shouldQuit {
					RootExitCode = code
					return
				}

				if exitCode, err := self.executeCommand(group, name, originalArgs, envVars, consents); err != nil {
					RootExitCode = exitCode
				}
			},
		}

		// legacy flag definition ("requiredFlags")
		// deprecated
		for _, flag := range requiredFlags {
			addFlagToCmd(cmd, flag)
		}
		// new ways to handle flags
		self.processFlags(group, name, cmd, flags, exclusiveFlags, groupFlags)

		cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(validArgsCmd) > 0 {
				output, err := self.executeValidArgsOfCommand(group, name, args)
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

		if v.RuntimeGroup() == "" {
			self.rootCmd.AddCommand(cmd)
		} else {
			if group, exists := self.groupCmds[v.RuntimeGroup()]; exists {
				group.AddCommand(cmd)
			} else {
				log.Errorf("cannot install cmd %s in group %s: group not found", v.Name(), v.Group())
			}
		}

	}
}

// parse args and inject environment vars
// if checkFlags is disabled, it simply returns the empty variables, and the args input
// otherwise, return the environment vars, original args, exit code, and if we should exit
func (self *defaultFrontend) parseArgsToEnvVars(c *cobra.Command, args []string, checkFlags bool) ([]string, []string, int, bool) {
	var envVars []string = []string{}
	var envTable map[string]string = map[string]string{}
	var originalArgs = args

	log.Debugf("checkFlags=%t", checkFlags)
	if checkFlags {
		var err error = nil
		envVarPrefix := strings.ToUpper(self.appCtx.AppName())
		envVars, envTable, originalArgs, err = parseCmdArgsToEnv(c, args, envVarPrefix)
		if err != nil {
			console.Error("Failed to parse arguments: %v", err)
			// set exit code to 1, and should quit
			return envVars, originalArgs, 1, true
		}
		if h, exist := envTable[fmt.Sprintf("%s_FLAG_HELP", envVarPrefix)]; exist && h == "true" {
			c.Help()
			// show help and should quit
			return envVars, originalArgs, 0, true
		}
	}
	log.Debugf("flag & args environments: %v", envVars)

	return envVars, originalArgs, 0, false
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

func (self *defaultFrontend) getExecutableCommand(group, name string) (command.Command, error) {
	iCmd, err := self.backend.FindCommand(group, name)
	return iCmd, err
}

// execute a cdt command
func (self *defaultFrontend) executeCommand(group, name string, args []string, initialEnvCtx []string, consent []string) (int, error) {
	iCmd, err := self.getExecutableCommand(group, name)
	if err != nil {
		return 1, err
	}
	if iCmd.Executable() == "" {
		return 1, errors.New(EXECUTABLE_NOT_DEFINED)
	}

	envCtx := self.getCmdEnvContext(initialEnvCtx, consent)
	envCtx = append(envCtx, fmt.Sprintf("%s=%s", self.appCtx.CmdPackageDirEnvVar(), iCmd.PackageDir()))

	exitCode, err := iCmd.Execute(envCtx, args...)
	if err != nil {
		return exitCode, err
	}

	return 0, nil
}

// execute the valid args command of the cdt command
func (self *defaultFrontend) executeValidArgsOfCommand(group, name string, args []string) (string, error) {
	iCmd, err := self.getExecutableCommand(group, name)
	if err != nil {
		return "", err
	}

	envCtx := self.getCmdEnvContext([]string{}, []string{})

	_, output, err := iCmd.ExecuteValidArgsCmd(envCtx, args...)
	if err != nil {
		return "", err
	}

	return output, nil
}

// execute the flag values command of the cdt command
func (self *defaultFrontend) executeFlagValuesOfCommand(group, name string, flagCmd []string, args []string) (string, error) {
	iCmd, err := self.getExecutableCommand(group, name)
	if err != nil {
		return "", err
	}

	envCtx := self.getCmdEnvContext([]string{}, []string{})

	_, output, err := iCmd.ExecuteFlagValuesCmd(envCtx, flagCmd, args...)
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

func (self *defaultFrontend) processFlags(cmdGroup, cmdName string, cmd *cobra.Command, flags []command.Flag, exclusive [][]string, group [][]string) {
	for _, flag := range flags {
		switch flag.Type() {
		case "bool":
			defaultV, err := strconv.ParseBool(flag.Default())
			if err != nil {
				defaultV = false
			}
			cmd.Flags().BoolP(flag.Name(), flag.ShortName(), defaultV, flag.Description())
		default:
			cmd.Flags().StringP(flag.Name(), flag.ShortName(), flag.Default(), flag.Description())
		}

		if flag.Required() {
			cmd.MarkFlagRequired(flag.Name())
		}

		if len(flag.Values()) > 0 {
			values := flag.Values()
			cmd.RegisterFlagCompletionFunc(flag.Name(), func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				return values, cobra.ShellCompDirectiveNoFileComp
			})
		} else if len(flag.ValuesCmd()) > 0 {
			valuesCmd := flag.ValuesCmd()
			cmd.RegisterFlagCompletionFunc(flag.Name(), func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
				fmt.Println(args)
				originalArgs := args
				c.LocalFlags().VisitAll(func(flag *pflag.Flag) {
					switch flag.Value.Type() {
					case "bool":
						if flag.Value.String() == "true" {
							originalArgs = append(originalArgs, fmt.Sprintf("--%s", flag.Name))
						}
					default:
						if flag.Value.String() != "" {
							originalArgs = append(originalArgs, fmt.Sprintf("--%s", flag.Name), flag.Value.String())
						}
					}
				})
				output, err := self.executeFlagValuesOfCommand(cmdGroup, cmdName, valuesCmd, originalArgs)
				if err != nil {
					return []string{}, cobra.ShellCompDirectiveDefault
				}
				return strings.Split(output, "\n"), cobra.ShellCompDirectiveNoFileComp
			})
		}
	}

	for _, ex := range exclusive {
		exist := true
		for _, f := range ex {
			if cmd.Flags().Lookup(f) == nil {
				exist = false
			}
		}
		if !exist {
			continue
		}
		cmd.MarkFlagsMutuallyExclusive(ex...)
	}
	for _, ex := range group {
		exist := true
		for _, f := range ex {
			if cmd.Flags().Lookup(f) == nil {
				exist = false
			}
		}
		if !exist {
			continue
		}
		cmd.MarkFlagsRequiredTogether(ex...)
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

func (self *defaultFrontend) getCmdEnvContext(envVars []string, consents []string) []string {
	vars := append([]string{}, envVars...)

	for _, item := range consents {
		switch item {
		case consent.USERNAME:
			username, err := helper.GetUsername()
			if err != nil {
				username = ""
			}
			if username != "" {
				vars = append(vars, fmt.Sprintf("%s=%s", self.appCtx.UsernameEnvVar(), username))
			}
		case consent.PASSWORD:
			password, err := helper.GetPassword()
			if err != nil {
				password = ""
			}
			if password != "" {
				vars = append(vars, fmt.Sprintf("%s=%s", self.appCtx.PasswordEnvVar(), password))
			}
		case consent.AUTH_TOKEN:
			token, err := helper.GetAuthToken()
			if err != nil {
				token = ""
			}
			if token != "" {
				vars = append(vars, fmt.Sprintf("%s=%s", self.appCtx.AuthTokenEnvVar(), token))
			}
		case consent.LOG_LEVEL:
			// append log level from configuration
			logLevel := viper.GetString(config.LOG_LEVEL_KEY)
			vars = append(vars, fmt.Sprintf("%s=%s",
				self.appCtx.LogLevelEnvVar(),
				logLevel,
			))
		case consent.DEBUG_FLAGS:
			// append debug flags from configuration
			debugFlags := os.Getenv(self.appCtx.DebugFlagsEnvVar())
			vars = append(vars, fmt.Sprintf("%s=%s,%s",
				self.appCtx.DebugFlagsEnvVar(),
				debugFlags,
				viper.GetString(config.DEBUG_FLAGS_KEY),
			))
		}
	}

	// Enable variable with prefix [binary_name] and COLA
	// TODO: remove it when in version 1.8 all variables are migrated to COLA prefix
	outputVars := []string{}
	for _, v := range vars {
		prefix := fmt.Sprintf("%s_", strings.ToUpper(self.appCtx.AppName()))
		if strings.HasPrefix(v, prefix) && prefix != "COLA_" {
			outputVars = append(outputVars, strings.Replace(v, prefix, "COLA_", 1))
		}
		outputVars = append(outputVars, v)
	}

	return outputVars
}

// return environment variable list, env variable table, original args including flags
func parseCmdArgsToEnv(c *cobra.Command, args []string, envVarPrefix string) ([]string, map[string]string, []string, error) {
	envVars := []string{}
	envTable := map[string]string{}
	originalArgs := []string{}
	if err := c.LocalFlags().Parse(args); err != nil {
		return envVars, envTable, args, err
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

		switch flag.Value.Type() {
		case "bool":
			if flag.Value.String() == "true" {
				originalArgs = append(originalArgs, fmt.Sprintf("--%s", flag.Name))
			}
		default:
			if flag.Value.String() != "" {
				originalArgs = append(originalArgs, fmt.Sprintf("--%s", flag.Name), flag.Value.String())
			}
		}

	})
	for idx, arg := range c.LocalFlags().Args() {
		k := fmt.Sprintf("%s_ARG_%s", envVarPrefix, strconv.Itoa(idx+1))
		envVars = append(envVars, fmt.Sprintf("%s=%s", k, arg))
		envTable[k] = arg
	}
	// new variable for arg number
	nargs_k := fmt.Sprintf("%s_NARGS", envVarPrefix)
	envTable[nargs_k] = strconv.Itoa(len(c.LocalFlags().Args()))
	envVars = append(envVars, fmt.Sprintf("%s=%s", nargs_k, envTable[nargs_k]))

	// reconstruct the original command args including flags
	parsedArgs := c.LocalFlags().Args()
	originalArgs = append(originalArgs, parsedArgs...)

	return envVars, envTable, originalArgs, nil
}
