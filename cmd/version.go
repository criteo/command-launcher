package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
)

func printVersion() {
	ctx, _ := context.AppContext()
	fmt.Printf("%s version %s\n", ctx.AppName(), getVersion(ctx.AppVersion(), ctx.AppBuildNum()))
}

func getVersion(version string, buildNum string) string {
	if version == "" {
		return fmt.Sprintf("dev, build %s", os.Getenv("USER"))
	}

	return fmt.Sprintf("%s, build %s", version, buildNum)
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
