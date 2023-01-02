package cmd

import (
	"fmt"
	"strings"

	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/console"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
)

type RenameFlags struct {
	delete bool
}

var (
	renameFlags = RenameFlags{}
)

func AddRenameCmd(rootCmd *cobra.Command, appCtx context.LauncherContext, back backend.Backend) {
	renameCmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename installed command",
		Long: fmt.Sprintf(`
Rename installed command to a different name.

Each command has an unique internal name in form of:
[name]@[group]@[package]@[repository]

For group command, the internal name is:
[group]@@[package]@[repository]

Without any conflict, the command name registered to Command Launcher is '%s [group] [name]'

To change the group name:
%s rename [group]@@[package]@[repository] [new group]

To change the command name:
%s rename [name]@[group]@[package]@[repository] [new name]`,
			strings.ToTitle(appCtx.AppName()),
			strings.ToTitle(appCtx.AppName()),
			strings.ToTitle(appCtx.AppName())),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 && !renameFlags.delete {
				cmd.Help()
				return nil
			}

			if renameFlags.delete && len(args) < 1 {
				cmd.Help()
				return nil
			}
			icmd, err := back.FindCommandByFullName(args[0])
			if err != nil {
				console.Error("No command with full name %s", args[0])
				return err
			}

			new_name := ""
			if !renameFlags.delete {
				new_name = args[1]
			}
			err = back.RenameCommand(icmd, new_name)
			return err
		},
		ValidArgsFunction: func(c *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			options := []string{}
			if len(args) == 0 {
				for _, group := range back.GroupCommands() {
					options = append(options, group.FullName())
				}
				for _, exec := range back.ExecutableCommands() {
					options = append(options, exec.FullName())
				}
				return options, cobra.ShellCompDirectiveNoFileComp
			}
			return []string{}, cobra.ShellCompDirectiveNoFileComp
		},
	}

	renameCmd.Flags().BoolVarP(&renameFlags.delete, "delete", "d", false, "delete renaming")
	rootCmd.AddCommand(renameCmd)
}
