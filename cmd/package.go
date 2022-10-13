package cmd

import (
	"fmt"

	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
)

type PackageFlags struct {
	gitUrl  string
	fileUrl string
	dropin  bool
	local   bool
	remote  bool
}

var (
	packageFlags = PackageFlags{}
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
	packageListCmd.Flags().BoolVar(&packageFlags.dropin, "dropin", false, "List only the dropin packages")
	packageListCmd.Flags().BoolVar(&packageFlags.local, "local", false, "List only the local packages")
	packageListCmd.Flags().BoolVar(&packageFlags.remote, "remote", false, "List only the remote packages")
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
