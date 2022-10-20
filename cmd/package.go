package cmd

import (
	"github.com/criteo/command-launcher/cmd/remote"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func findPackageNames(includeRemote bool) map[string][]string {
	packages := map[string][]string{}
	packages["local"] = make([]string, 0)
	packages["dropin"] = make([]string, 0)
	packages["remote"] = make([]string, 0)

	for _, pkg := range rootCtxt.localRepo.InstalledPackages() {
		packages["local"] = append(packages["local"], pkg.Name())
	}

	for _, pkg := range rootCtxt.dropinRepo.InstalledPackages() {
		packages["dropin"] = append(packages["dropin"], pkg.Name())
	}

	if includeRemote {
		remote := remote.CreateRemoteRepository(viper.GetString(config.COMMAND_REPOSITORY_BASE_URL_KEY))
		if remotes, err := remote.All(); err == nil {
			for _, pkg := range remotes {
				packages["remote"] = append(packages["remote"], pkg.Name)
			}
		}
	}

	return packages
}

func packageNameValidatonFunc(includeRemote bool) func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
	return func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		packages := findPackageNames(includeRemote)

		availablePkgs := []string{}
		for _, v := range packages {
			availablePkgs = append(availablePkgs, v...)
		}

		return availablePkgs, cobra.ShellCompDirectiveNoFileComp
	}
}

// func AddPackageCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {

// 	packageUpdateCmd := &cobra.Command{
// 		Use:   "update [package name]",
// 		Short: "Update a package",
// 		Long:  "Update a package from its name, only when the packge is a Git repo",
// 		Args:  cobra.ExactArgs(1),
// 		Example: fmt.Sprintf(`
//   %s package update my-pkg`, appCtx.AppName()),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			folder, err := findPackageFolder(args[0])
// 			if err != nil {
// 				return err
// 			}

// 			gitFolder := filepath.Join(folder, ".git")
// 			stat, err := os.Stat(gitFolder)
// 			if os.IsNotExist(err) || !stat.IsDir() {
// 				return fmt.Errorf("the package %s is not installed from a git repo", args[0])
// 			}

// 			ctx := exec.Command("git", "pull")
// 			ctx.Dir = folder
// 			ctx.Stdout = os.Stdout
// 			ctx.Stderr = os.Stderr
// 			ctx.Stdin = os.Stdin

// 			if err = ctx.Run(); err != nil {
// 				return fmt.Errorf("git pull has failed: %v", err)
// 			}

// 			return nil
// 		},
// 	}
// }
