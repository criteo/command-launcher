package cmd

import (
	"fmt"
	"os"
	"syscall"

	"github.com/criteo/command-launcher/internal/helper"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

type LoginFlags struct {
	username string
	password string
}

var loginFlags = LoginFlags{}

func defaultUsername() string {
	user, present := os.LookupEnv("USER")
	if !present {
		user, _ = os.LookupEnv("USERNAME")
	}

	return user
}

func init() {
	loginCmd.Flags().StringVarP(&loginFlags.username, "user", "u", defaultUsername(), "Criteo user name")
	loginCmd.Flags().StringVarP(&loginFlags.password, "password", "p", "", "Criteo user password")
	rootCmd.AddCommand(loginCmd)
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to Criteo services",
	Long: `
Login to Criteo services.

You can specify the your Criteo password from:
1. command option --password (-p)
2. environment variable CDT_PASSWORD
3. from command line input

The credential will be stored in your system vault.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		username := loginFlags.username
		if username == "" {
			username = os.Getenv("CDT_USERNAME")
			if username == "" {
				fmt.Printf("Please enter your Criteo user name: ")
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
			passwd = os.Getenv("CDT_PASSWORD")
			if passwd == "" {
				fmt.Printf("Please enter your Criteo IT password: ")
				pass, err := terminal.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return err
				}
				passwd = string(pass)
			}
		}

		fmt.Println()

		helper.SetSecret("cdt-username", username)
		helper.SetSecret("cdt-password", passwd)
		return nil
	},
}
