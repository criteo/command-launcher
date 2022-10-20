package cmd

import (
	"fmt"
	"os"

	"github.com/criteo/command-launcher/internal/command"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
)

func AddUninstallCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	packageUninstallCmd := &cobra.Command{
		Use:   "delete [package_name]",
		Short: "Remove a package",
		Long:  "Remove a package from its name",
		Args:  cobra.ExactArgs(1),
		Example: fmt.Sprintf(`
  %s package delete my-pkg`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			folder, err := findPackageFolder(args[0])
			if err != nil {
				return err
			}

			return os.RemoveAll(folder)
		},
	}

	rootCmd.AddCommand(packageUninstallCmd)
}

func findPackageFolder(pkgName string) (string, error) {
	if pkgName == "" {
		return "", fmt.Errorf("invalid package name")
	}

	var pkgMf command.PackageManifest
	for _, pkg := range rootCtxt.dropinRepo.InstalledPackages() {
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
