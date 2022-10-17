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

type PackageFlags struct {
	gitUrl     string
	fileUrl    string
	dropin     bool
	local      bool
	remote     bool
	includeCmd bool
}

var (
	packageFlags = PackageFlags{}
)

func AddPackageCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	packageCmd := &cobra.Command{
		Use:   "package",
		Short: "Manage the Packages",
		Long:  "Provide CRUD operations on the Packages",
	}

	packageListCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		Long:  "List installed packages with details",
		PreRun: func(cmd *cobra.Command, args []string) {
			if !packageFlags.dropin && !packageFlags.local && !packageFlags.remote {
				packageFlags.dropin = true
				packageFlags.local = true
				packageFlags.remote = true
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if packageFlags.local {
				printPackages(rootCtxt.localRepo, "local repository", packageFlags.includeCmd)
			}

			if packageFlags.dropin {
				printPackages(rootCtxt.dropinRepo, "dropin repository", packageFlags.includeCmd)
			}

			if packageFlags.remote {
				remote := remote.CreateRemoteRepository(viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY))
				if packages, err := remote.All(); err == nil {
					printPackageInfos(packages, "remote repository")
				} else {
					console.Warn("Cannot load the remote repository: %v", err)
				}
			}
		},
	}
	packageListCmd.Flags().BoolVar(&packageFlags.dropin, "dropin", false, "List only the dropin packages")
	packageListCmd.Flags().BoolVar(&packageFlags.local, "local", false, "List only the local packages")
	packageListCmd.Flags().BoolVar(&packageFlags.remote, "remote", false, "List only the remote packages")
	packageListCmd.Flags().BoolVar(&packageFlags.includeCmd, "include-cmd", false, "List the packages with all commands")
	packageListCmd.Flags().BoolP("all", "a", true, "List all packages")
	packageListCmd.MarkFlagsMutuallyExclusive("all", "dropin", "local", "remote")

	packageInstallCmd := &cobra.Command{
		Use:   "install [package_name]",
		Short: "Install a package",
		Long:  "Install a package package from a git repo or from a zip file or from its name",
		Args:  cobra.MaximumNArgs(1),
		Example: fmt.Sprintf(`
  %s package install --git https://example.com/my-repo.git my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {

			return nil
		},
	}
	packageInstallCmd.Flags().StringVar(&packageFlags.fileUrl, "file", "", "URL or path of a package file")
	packageInstallCmd.Flags().StringVar(&packageFlags.gitUrl, "git", "", "URL of a Git repo of package")
	packageInstallCmd.MarkFlagsMutuallyExclusive("git", "file")

	packageDeleteCmd := &cobra.Command{
		Use:   "delete [package_name]",
		Short: "Remove a package",
		Long:  "Remove a package from its name",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(`
  %s package delete my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	packageUpdateCmd := &cobra.Command{
		Use:   "update [package name]",
		Short: "Update a package",
		Long:  "Update a package from its name, only when the packge is a Git repo",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(`
  %s package update my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	packageCmd.AddCommand(packageListCmd)
	packageCmd.AddCommand(packageInstallCmd)
	packageCmd.AddCommand(packageDeleteCmd)
	packageCmd.AddCommand(packageUpdateCmd)

	rootCmd.AddCommand(packageCmd)
}

func printPackages(repo repository.PackageRepository, name string, includeCmd bool) {
	console.Highlight("=== %s ===\n", strings.Title(name))
	for _, pkg := range repo.InstalledPackages() {
		fmt.Printf("  - %s - %s\n", pkg.Name(), pkg.Version())
		if includeCmd {
			printCommands(pkg.Commands())
		}
	}
	fmt.Println()
}

func printPackageInfos(packages []remote.PackageInfo, name string) {
	console.Highlight("=== %s ===\n", strings.Title(name))
	for _, pkg := range packages {
		fmt.Printf("%2s %s - %s\n", "-", pkg.Name, pkg.Version)
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
