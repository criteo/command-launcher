package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/helper"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

type LoginFlags struct {
	username string
	password string
}

var (
	loginFlags = LoginFlags{}
)

func defaultUsername() string {
	user, present := os.LookupEnv("USER")
	if !present {
		user, _ = os.LookupEnv("USERNAME")
	}

	return user
}

func AddLoginCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	loginCmd := &cobra.Command{
		Use:   "login",
		Short: "Login to use services",
		Long: fmt.Sprintf(`
Login to use services.

You can specify the your password from:
1. command option --password (-p)
2. environment variable %s
3. from command line input

The credential will be stored in your system vault.`, appCtx.PasswordEnvVar()),
		RunE: func(cmd *cobra.Command, args []string) error {
			appCtx, _ := context.AppContext()
			username := loginFlags.username
			if username == "" {
				username = os.Getenv(appCtx.UsernameEnvVar())
				if username == "" {
					fmt.Printf("Please enter your user name: ")
					nb, err := fmt.Scan(&username)
					if err != nil {
						return err
					}

					if nb != 1 {
						return fmt.Errorf("invalid entries (expected only one argument)")
					}
				}
			}

			passwd := loginFlags.password
			if passwd == "" {
				passwd = os.Getenv(appCtx.PasswordEnvVar())
				if passwd == "" {
					fmt.Printf("Please enter your password: ")
					pass, err := terminal.ReadPassword(int(syscall.Stdin))
					if err != nil {
						return err
					}
					passwd = string(pass)
				}
			}

			fmt.Println()

			helper.SetUsername(username)
			helper.SetPassword(passwd)
			return nil
		},
	}
	loginCmd.Flags().StringVarP(&loginFlags.username, "user", "u", defaultUsername(), "User name")
	loginCmd.Flags().StringVarP(&loginFlags.password, "password", "p", "", "User password")

	rootCmd.AddCommand(loginCmd)
}
