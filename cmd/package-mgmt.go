package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/criteo/command-launcher/internal/pkg"
	"github.com/criteo/command-launcher/internal/remote"
	"github.com/criteo/command-launcher/internal/repository"
	"github.com/criteo/command-launcher/internal/updateConfig"
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
		Short: "Manage command launcher packages",
		Long:  "Manage command launcher packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				cmd.Help()
			}
			return nil
		},
	}
	packageListCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packages and commands",
		Long:  "List installed packages and commands with details",
		PreRun: func(cmd *cobra.Command, args []string) {
			if !packageFlags.dropin && !packageFlags.local && !packageFlags.remote {
				packageFlags.dropin = true
				packageFlags.local = true
				packageFlags.remote = false
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if packageFlags.local {
				for _, s := range rootCtxt.backend.AllPackageSources() {
					if s.IsManaged && s.Repo != nil {
						printPackages(s.Repo, fmt.Sprintf("managed repository: %s", s.Repo.Name()), packageFlags.includeCmd)
					}
				}
			}

			if packageFlags.dropin {
				printPackages(rootCtxt.backend.DropinRepository(), "dropin repository", packageFlags.includeCmd)
			}

			if packageFlags.remote {
				for _, s := range rootCtxt.backend.AllPackageSources() {
					if s.IsManaged {
						remote := remote.CreateRemoteRepository(s.RemoteBaseURL)
						if packages, err := remote.All(); err == nil {
							printPackageInfos(packages, fmt.Sprintf("remote registry: %s", s.Repo.Name()))
						} else {
							console.Warn("Cannot load the remote registry: %v", err)
						}
					}
				}
			}
		},
		ValidArgsFunction: noArgCompletion,
	}
	packageListCmd.Flags().BoolVar(&packageFlags.dropin, "dropin", false, "List only the dropin packages")
	packageListCmd.Flags().BoolVar(&packageFlags.local, "local", false, "List only the local packages")
	packageListCmd.Flags().BoolVar(&packageFlags.remote, "remote", false, "List only the remote packages")
	packageListCmd.Flags().BoolVar(&packageFlags.includeCmd, "include-cmd", false, "List the packages with all commands")
	packageListCmd.Flags().BoolP("all", "a", true, "List all packages")
	packageListCmd.MarkFlagsMutuallyExclusive("all", "dropin", "local", "remote")

	packageInstallCmd := &cobra.Command{
		Use:   "install [package_name]",
		Short: "Install a dropin package",
		Long:  "Install a dropin package package from a git repo or from a zip file or from its name",
		Args:  cobra.MaximumNArgs(1),
		Example: fmt.Sprintf(`
  %s install --git https://example.com/my-repo.git my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			if packageFlags.fileUrl != "" {
				return installZipFile(packageFlags.fileUrl)
			}

			if packageFlags.gitUrl != "" {
				return installGitRepo(packageFlags.gitUrl)
			}

			return nil
		},
		ValidArgsFunction: noArgCompletion,
	}
	packageInstallCmd.Flags().StringVar(&packageFlags.fileUrl, "file", "", "URL or path of a package file")
	packageInstallCmd.Flags().StringVar(&packageFlags.gitUrl, "git", "", "URL of a Git repo of package")
	packageInstallCmd.MarkFlagsMutuallyExclusive("git", "file")

	packageDeleteCmd := &cobra.Command{
		Use:   "delete [package_name]",
		Short: "Remove a dropin package",
		Long:  "Remove a dropin package from its name",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(`
  %s delete my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			folder, err := findPackageFolder(args[0])
			if err != nil {
				return err
			}

			return os.RemoveAll(folder)
		},
		ValidArgsFunction: packageNameValidatonFunc(false, true, false),
	}

	packageSetupCmd := &cobra.Command{
		Use:   "setup [package_name]",
		Short: "Setup a package",
		Long: `
Manually setup a package.

This command  will trigger the system command __setup__ defined in the package manifest.
To enable the automatic setup during package installation, enable the configuration:
"enable_package_setup_hook".
`,
		Args: cobra.ExactArgs(1),
		Example: fmt.Sprintf(`
  %s setup my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, s := range rootCtxt.backend.AllPackageSources() {
				for _, installedPkg := range s.Repo.InstalledPackages() {
					if installedPkg.Name() == args[0] {
						return pkg.ExecSetupHookFromPackage(installedPkg, "")
					}
				}
			}
			return fmt.Errorf("no package named %s found", args[0])
		},
		ValidArgsFunction: packageNameValidatonFunc(true, true, false),
	}

	packageInspectCmd := &cobra.Command{
		Use:   "inspect [package_name]",
		Short: "Show details of an installed package",
		Long:  "Show detailed information about an installed package including its source, version, local path, pause status, and commands",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(`
  %s package inspect my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return inspectPackage(args[0])
		},
		ValidArgsFunction: packageNameValidatonFunc(true, true, false),
	}

	packagePauseCmd := &cobra.Command{
		Use:   "pause [package_name]",
		Short: "Pause update for a package",
		Long:  "Pause update for a package",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(`
  %s pause my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := rootCtxt.backend.DefaultRepository().PausePackageUpdate(args[0])
			if err != nil {
				return err
			}
			console.Success("Package %s updates are paused", args[0])
			return nil
		},
		ValidArgsFunction: packageNameValidatonFunc(true, true, false),
	}

	packageCmd.AddCommand(packageListCmd)
	packageCmd.AddCommand(packageInstallCmd)
	packageCmd.AddCommand(packageDeleteCmd)
	packageCmd.AddCommand(packageSetupCmd)
	packageCmd.AddCommand(packagePauseCmd)
	packageCmd.AddCommand(packageInspectCmd)
	rootCmd.AddCommand(packageCmd)
}

func packageNameValidatonFunc(includeLocal bool, includeDropin bool, includeRemote bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		pkgTable := map[string]string{}

		if includeLocal {
			for _, s := range rootCtxt.backend.AllPackageSources() {
				if s.IsManaged && s.Repo != nil {
					for _, pkg := range s.Repo.InstalledPackages() {
						pkgTable[pkg.Name()] = pkg.Version()
					}
				}
			}
		}
		if includeDropin {
			for _, pkg := range rootCtxt.backend.DropinRepository().InstalledPackages() {
				pkgTable[pkg.Name()] = pkg.Version()
			}
		}

		if includeRemote {
			remote := remote.CreateRemoteRepository(viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY))
			if packages, err := remote.All(); err == nil {
				for _, pkg := range packages {
					pkgTable[pkg.Name] = pkg.Version
				}
			}
		}

		availablePkgs := []string{}
		for k, _ := range pkgTable {
			availablePkgs = append(availablePkgs, k)
		}

		return availablePkgs, cobra.ShellCompDirectiveNoFileComp
	}
}

func noArgCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func printPackages(repo repository.PackageRepository, name string, includeCmd bool) {
	console.Highlight("=== %s ===\n", strings.Title(name))
	for _, pkg := range repo.InstalledPackages() {
		if includeCmd {
			fmt.Printf("  Package: %s (v%s)\n", pkg.Name(), pkg.Version())
			printCommands(pkg.Commands())
			fmt.Println()
		} else {
			fmt.Printf("  - %-50s %s\n", pkg.Name(), pkg.Version())
		}
	}
	if !includeCmd {
		fmt.Println()
	}
}

func printPackageInfos(packages []remote.PackageInfo, name string) {
	console.Highlight("=== %s ===\n", strings.Title(name))
	for _, pkg := range packages {
		fmt.Printf("%2s %-50s %s\n", "-", pkg.Name, pkg.Version)
	}
	fmt.Println()
}

func printCommands(commands []command.Command) {
	cmdMap := make(map[string][]command.Command)
	cmdMap["__no_group__"] = make([]command.Command, 0)
	groupDescs := make(map[string]string)

	for _, cmd := range commands {
		if cmd.Type() == "group" {
			cmdMap[cmd.Name()] = make([]command.Command, 0)
			groupDescs[cmd.Name()] = cmd.ShortDescription()
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
			desc := groupDescs[g]
			if desc != "" {
				fmt.Printf("%4s %-30s %s\n", "*", g, desc)
			} else {
				fmt.Printf("%4s %s\n", "*", g)
			}
			for _, c := range cs {
				cmdDesc := c.ShortDescription()
				if cmdDesc != "" {
					fmt.Printf("%6s %-28s %s\n", "-", c.Name(), cmdDesc)
				} else {
					fmt.Printf("%6s %s\n", "-", c.Name())
				}
			}
		}
	}
}

func installGitRepo(gitUrl string) error {
	_, err := url.Parse(gitUrl)
	if err != nil {
		return fmt.Errorf("invalid url or pathname: %v", err)
	}

	path := viper.GetString(config.DROPIN_FOLDER_KEY)
	gitPkg, err := pkg.CreateGitRepoPackage(gitUrl)
	if err != nil {
		os.RemoveAll(path)
		return fmt.Errorf("failed to install git package %s: %v", gitUrl, err)
	}

	mf, err := gitPkg.InstallTo(viper.GetString(config.DROPIN_FOLDER_KEY))
	if err != nil {
		os.RemoveAll(path)
		return fmt.Errorf("failed to install git package %s: %v", gitUrl, err)
	}

	console.Success("Package %s installed in the dropin repository", mf.Name())
	return nil
}

func installZipFile(fileUrl string) error {
	url, err := url.Parse(fileUrl)
	if err != nil {
		return fmt.Errorf("invalid url or pathname: %v", err)
	}

	var pathname string
	if url.IsAbs() {
		if url.Scheme == "file" {
			pathname = url.Path
		} else {
			tmpDir, err := os.MkdirTemp("", "package-download-*")
			if err != nil {
				return fmt.Errorf("cannot create temporary dir (%v)", err)
			}
			defer os.RemoveAll(tmpDir)

			pathname = filepath.Join(tmpDir, filepath.Base(url.Path))
			if err := helper.DownloadFile(fileUrl, pathname, true); err != nil {
				return fmt.Errorf("error downloading %s: %v", url, err)
			}
		}
	} else {
		pathname = url.Path
	}

	zipPkg, err := pkg.CreateZipPackage(pathname)
	if err != nil {
		return fmt.Errorf("cannot create the package from the zip file: %v", err)
	}

	targetDir := filepath.Join(viper.GetString(config.DROPIN_FOLDER_KEY), zipPkg.Name())
	mf, err := zipPkg.InstallTo(targetDir)
	if err != nil {
		return fmt.Errorf("failed to install zip package %s: %v", fileUrl, err)
	}

	console.Success("Package '%s' version %s installed in the dropin repository", mf.Name(), mf.Version())
	return nil
}

type packageMatch struct {
	pkg    command.PackageManifest
	source *backend.PackageSource
}

func inspectPackage(pkgName string) error {
	matches := []packageMatch{}

	for _, s := range rootCtxt.backend.AllPackageSources() {
		if s.Repo == nil {
			continue
		}
		for _, p := range s.Repo.InstalledPackages() {
			if p.Name() == pkgName {
				matches = append(matches, packageMatch{pkg: p, source: s})
				break
			}
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("no package named %s found", pkgName)
	}

	for i, m := range matches {
		if i > 0 {
			fmt.Println()
		}
		printPackageDetails(m.pkg, m.source)
	}

	return nil
}

func printPackageDetails(pkg command.PackageManifest, source *backend.PackageSource) {
	console.Highlight("Package: %s (source: %s)\n", pkg.Name(), source.Name)
	fmt.Printf("  Full Name:  %s@%s\n", pkg.Name(), source.Name)
	fmt.Printf("  Version:    %s\n", pkg.Version())
	fmt.Printf("  Source:     %s\n", source.Name)
	fmt.Printf("  Managed:    %v\n", source.IsManaged)

	if source.IsManaged {
		fmt.Printf("  Remote URL: %s\n", source.RemoteBaseURL)
		fmt.Printf("  Registry:   %s\n", source.RemoteRegistryURL)
		fmt.Printf("  Sync:       %s\n", source.SyncPolicy)
	}

	if repoFolder, err := source.Repo.RepositoryFolder(); err == nil {
		fmt.Printf("  Local Path: %s\n", repoFolder)
	}

	if source.IsManaged {
		paused := false
		pausedUntil := time.Time{}
		if exists, _ := updateConfig.IsUpdateConfigExists(source.RepoDir); exists {
			if cfg, err := updateConfig.ReadFromDir(source.RepoDir); err == nil {
				if cfg.IsPackagePaused(pkg.Name()) {
					paused = true
					pausedUntil = cfg.PausedUntil[pkg.Name()]
				}
			}
		}
		fmt.Printf("  Paused:     %v\n", paused)
		if paused {
			fmt.Printf("  Paused Until: %s\n", pausedUntil.Format(time.RFC3339))
		}
	}

	fmt.Println()
	fmt.Println("  Commands:")
	printCommands(pkg.Commands())
}

func findPackageFolder(pkgName string) (string, error) {
	if pkgName == "" {
		return "", fmt.Errorf("invalid package name")
	}

	var pkgMf command.PackageManifest
	for _, pkg := range rootCtxt.backend.DropinRepository().InstalledPackages() {
		if pkg.Name() == pkgName {
			pkgMf = pkg
			break
		}
	}

	if pkgMf == nil {
		return "", fmt.Errorf("cannot find the package in the dropin repository")
	}

	if len(pkgMf.Commands()) == 0 {
		return "", fmt.Errorf("cannot find the package folder in the dropin repository")
	}

	return pkgMf.Commands()[0].PackageDir(), nil
}
