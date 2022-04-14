package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
)

const semanticVersion = "1.0.0"

func printVersion() {
	ctx, _ := context.AppContext()
	fmt.Printf("%s version %s\n", ctx.AppName(), getVersion(ctx.AppVersion()))
}

func getVersion(version string) string {
	if version == "" {
		return fmt.Sprintf("%s, build dev-%s", semanticVersion, os.Getenv("USER"))
	}

	return fmt.Sprintf("%s, build %s", semanticVersion, version)
}

func AddversionCmd(rootCmd *cobra.Command, appCtx context.LauncherContext) {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: fmt.Sprintf("Print the version number of %s command", strings.ToTitle(appCtx.AppName())),
		Long:  fmt.Sprintf(`All software has versions. This is %s's`, strings.ToTitle(appCtx.AppName())),
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
	rootCmd.AddCommand(versionCmd)
}
