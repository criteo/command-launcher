package cmd

import (
	"fmt"
	"os"

	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ServeFlags struct {
	port int
}

var (
	serveFlags = ServeFlags{}
)

func AddServeCmd(rootCmd *cobra.Command, appCtx context.LauncherContext, back backend.Backend) {
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the server",
		Long:  `Start the server`,
		Run: func(cmd *cobra.Command, args []string) {
			port := viper.GetInt(config.DAEMON_PORT_KEY)
			if serveFlags.port != 0 {
				port = serveFlags.port
			}
			fmt.Printf("Starting server on port %d", port)
			if err := back.Serve(port); err != nil {
				fmt.Printf("Failed to start server: %s\n", err)
				os.Exit(1)
			}
		},
	}

	serveCmd.Flags().IntVarP(&serveFlags.port, "port", "p", 0, "Port to listen to")
	rootCmd.AddCommand(serveCmd)
}
