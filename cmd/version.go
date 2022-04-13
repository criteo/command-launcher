package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const semanticVersion = "1.0.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: fmt.Sprintf("Print the version number of %s command", strings.ToTitle(rootCmd.Use)),
	Long:  fmt.Sprintf(`All software has versions. This is %s's`, strings.ToTitle(rootCmd.Use)),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("%s version %s\n", rootCmd.Use, getVersion())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func getVersion() string {
	if BuildVersion == "" {
		return fmt.Sprintf("%s, build dev-%s", semanticVersion, os.Getenv("USER"))
	}

	return fmt.Sprintf("%s, build %s", semanticVersion, BuildVersion)
}
