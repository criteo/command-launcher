package cmd

import (
	"fmt"
	"os"

	"github.com/criteo/command-launcher/cmd/completion"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
)

func AddCompletionCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	var completionCmd = &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate completion script",
		Long: fmt.Sprintf(`To load completions:

Bash:

  $ source <(%[1]s completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ %[1]s completion bash > /etc/bash_completion.d/%[1]s
  # macOS:
  $ %[1]s completion bash > $(brew --prefix)/etc/bash_completion.d/%[1]s

Zsh:

  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:

  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ %[1]s completion zsh > "${fpath[1]}/_%[1]s"

  # You will need to start a new shell for this setup to take effect.

fish:

  $ %[1]s completion fish | source

  # To load completions for each session, execute once:
  $ %[1]s completion fish > ~/.config/fish/completions/%[1]s.fish

PowerShell:

  PS> %[1]s completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> %[1]s completion powershell > %[1]s.ps1
  # and source this file from your PowerShell profile.
`, appCtx.AppName()),
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		//Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) > 0 {
				switch args[0] {
				case "bash":
					// use our own fork of bash completion, it contains a slightly change
					// to ensure bash performance on windows
					// NOTE: bash on windows has low performance to process the completion descriptions
					// we disable the completion description for bash
					completion.GenBashCompletionV2(os.Stdout, appCtx.AppName(), false)
				case "zsh":
					cmd.Root().GenZshCompletion(os.Stdout)
				case "fish":
					cmd.Root().GenFishCompletion(os.Stdout, true)
				case "powershell":
					cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
				}
			} else {
				cmd.Help()
			}
		},
	}

	rootCmd.AddCommand(completionCmd)
}
