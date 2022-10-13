package cmd

import (
	"fmt"

	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
)

type DropinFlags struct {
	pkgName string
}

var (
	dropinFlags = DropinFlags{}
)

func AddDropinCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	packageCmd := &cobra.Command{
		Use:   "package",
		Short: "Manage the Packages",
		Long:  "Provide CRUD operations on the Packages",
	}

	packageListCmd := &cobra.Command{
		Use:   "list",
		Short: "List installed packages",
		Long:  "List installed packages with details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	packageInstallCmd := &cobra.Command{
		Use:   "install",
		Short: "Install a package",
		Long:  "Install a package package from a git repo or from a zip or from a local repository",
		Example: fmt.Sprintf(`
  %s package install https://example.com/my-repo.git --name my-command`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	packageInstallCmd.Flags().StringVarP(&dropinFlags.pkgName, "name", "n", "", "Package name")

	packageDeleteCmd := &cobra.Command{
		Use:   "delete [package name]",
		Short: "Remove a package",
		Long:  "Remove a package from its name",
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
