package cmd

import (
	"fmt"
	"strings"

	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/cmd/repository"
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ListFlags struct {
	dropin     bool
	local      bool
	remote     bool
	includeCmd bool
}

var (
	listFlags = ListFlags{}
)

func AddListCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	packageListCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		Long:  "List installed packages with details",
		PreRun: func(cmd *cobra.Command, args []string) {
			if !listFlags.dropin && !listFlags.local && !listFlags.remote {
				listFlags.dropin = true
				listFlags.local = true
				listFlags.remote = true
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if listFlags.local {
				printPackages(rootCtxt.localRepo, "local repository", listFlags.includeCmd)
			}

			if listFlags.dropin {
				printPackages(rootCtxt.dropinRepo, "dropin repository", listFlags.includeCmd)
			}

			if listFlags.remote {
				remote := remote.CreateRemoteRepository(viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY))
				if packages, err := remote.All(); err == nil {
					printPackageInfos(packages, "remote repository")
				} else {
					console.Warn("Cannot load the remote repository: %v", err)
				}
			}
		},
	}
	packageListCmd.Flags().BoolVar(&listFlags.dropin, "dropin", false, "List only the dropin packages")
	packageListCmd.Flags().BoolVar(&listFlags.local, "local", false, "List only the local packages")
	packageListCmd.Flags().BoolVar(&listFlags.remote, "remote", false, "List only the remote packages")
	packageListCmd.Flags().BoolVar(&listFlags.includeCmd, "include-cmd", false, "List the packages with all commands")
	packageListCmd.Flags().BoolP("all", "a", true, "List all packages")
	packageListCmd.MarkFlagsMutuallyExclusive("all", "dropin", "local", "remote")

	rootCmd.AddCommand(packageListCmd)
}

func printPackages(repo repository.PackageRepository, name string, includeCmd bool) {
	console.Highlight("=== %s ===\n", strings.Title(name))
	for _, pkg := range repo.InstalledPackages() {
		fmt.Printf("%2s %-20s %s\n", "-", pkg.Name(), pkg.Version())
		if includeCmd {
			printCommands(pkg.Commands())
		}
	}
	fmt.Println()
}

func printPackageInfos(packages []remote.PackageInfo, name string) {
	console.Highlight("=== %s ===\n", strings.Title(name))
	for _, pkg := range packages {
		fmt.Printf("%2s %-20s %s\n", "-", pkg.Name, pkg.Version)
	}
	fmt.Println()
}

func printCommands(commands []command.Command) {
	cmdMap := make(map[string][]command.Command)
	cmdMap["__no_group__"] = make([]command.Command, 0)

	for _, cmd := range commands {
		if cmd.Type() == "group" {
			cmdMap[cmd.Name()] = make([]command.Command, 0)
		} else if cmd.Type() == "executable" {
			if cmd.Group() != "" {
				cmdMap[cmd.Group()] = append(cmdMap[cmd.Group()], cmd)
			} else {
				cmdMap["__no_group__"] = append(cmdMap[cmd.Group()], cmd)
			}
		}
	}

	for g, cs := range cmdMap {
		if len(cmdMap[g]) > 0 {
			fmt.Printf("%4s %-20s %s\n", "*", g, "(group)")
			for _, c := range cs {
				fmt.Printf("%6s %-20s %s\n", "-", c.Name(), "(cmd)")
			}
		}
	}
}
