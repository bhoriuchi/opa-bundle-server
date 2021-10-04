package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bhoriuchi/opa-bundle-server/core/server"
	"github.com/bhoriuchi/opa-bundle-server/core/service"
	"github.com/spf13/cobra"
)

func initServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "server commands",
	}

	cmd.AddCommand(initServerStartCmd())
	return cmd
}

func initServerStartCmd() *cobra.Command {
	var (
		watch      bool
		configFile string
		logLevel   string
		logFormat  string
	)

	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if configFile == "" {
				return fmt.Errorf("no config specified")
			}

			p, err := filepath.Abs(configFile)
			if err != nil {
				return fmt.Errorf("failed to get absolute file path: %s", err)
			}

			cfg := &service.Config{
				Watch:     watch,
				File:      p,
				LogLevel:  logLevel,
				LogFormat: logFormat,
			}

			srv, err := server.NewServer(cfg)
			if err != nil {
				return fmt.Errorf("failed to create new server: %s", err)
			}

			return srv.Start(context.Background())
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&configFile, "config", "c", os.Getenv("OPA_BUNDLE_SERVER_CONFIG"), "Location of the config file")
	flags.StringVarP(&logLevel, "log-level", "l", os.Getenv("OPA_BUNDLE_SERVER_LOG_LEVEL"), "Log level (error, info, warn, debug)")
	flags.StringVar(&logFormat, "log-format", os.Getenv("OPA_BUNDLE_SERVER_LOG_FORMAT"), "Log format (text, json-pretty, or json)")

	return cmd
}
