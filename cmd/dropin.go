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
	dropinCmd := &cobra.Command{
		Use:   "dropin",
		Short: "Manage the dropin commands",
		Long:  "Provide CRUD operations on the dropin commands",
	}

	dropinListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all installed dropin packages",
		Long:  "List all installed dropin packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	dropinInstallCmd := &cobra.Command{
		Use:   "install",
		Short: "Install a dropin package from a git repo",
		Long:  "Install a dropin package from a git repo",
		Example: fmt.Sprintf(`
  %s dropin install https://example.com/my-repo.git --name my-command`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	dropinInstallCmd.Flags().StringVarP(&dropinFlags.pkgName, "name", "n", "", "Package name")

	dropinDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove a dropin package",
		Long:  "Remove a dropin package",
		Example: fmt.Sprintf(`
  %s dropin delete [package name]`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	dropinUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update a dropin package",
		Long:  "Update a dropin package",
		Example: fmt.Sprintf(`
  %s dropin update [package name]`, appCtx.AppName()),
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	dropinCmd.AddCommand(dropinListCmd)
	dropinCmd.AddCommand(dropinInstallCmd)
	dropinCmd.AddCommand(dropinDeleteCmd)
	dropinCmd.AddCommand(dropinUpdateCmd)

	rootCmd.AddCommand(dropinCmd)
}
