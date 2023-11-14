package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/criteo/command-launcher/internal/backend"
	"github.com/criteo/command-launcher/internal/config"
	"github.com/criteo/command-launcher/internal/context"
	"github.com/criteo/command-launcher/internal/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type ServeFlags struct {
	port      int
	noBrowser bool
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

			go func() {
				for {
					time.Sleep(1 * time.Second)
					resp, err := http.Get(fmt.Sprintf("http://localhost:%d/health", port))
					if err != nil {
						continue
					}
					if resp.StatusCode != http.StatusOK {
						continue
					}
					break
				}
				fmt.Printf("Server started on port %d\n", port)
				if !serveFlags.noBrowser {
					openbrowser(fmt.Sprintf("http://localhost:%d", port))
				}
			}()

			fmt.Printf("Starting server on port %d\n", port)
			if err := server.Serve(&back, port); err != nil {
				fmt.Printf("Failed to start server: %s\n", err)
				os.Exit(1)
			}
		},
	}

	serveCmd.Flags().IntVarP(&serveFlags.port, "port", "p", 0, "Port to listen to")
	serveCmd.Flags().BoolVarP(&serveFlags.noBrowser, "no-browser", "n", false, "Do not open the browser")
	rootCmd.AddCommand(serveCmd)
}

func openbrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}
