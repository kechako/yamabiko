package main

import (
	"github.com/kechako/yamabiko/config"
	"github.com/kechako/yamabiko/server"
	"github.com/spf13/cobra"
)

func serverCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run the Yamabiko server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfgPath, err := cmd.Flags().GetString("config")
			if err != nil || cfgPath == "" {
				path, err := config.FindConfigFile()
				if err != nil {
					return exit(err, 1)
				}
				cfgPath = path
			}

			cfg, err := config.LoadFile(cfgPath)
			if err != nil {
				return exit(err, 1)
			}

			s, err := server.New(cfg)
			if err != nil {
				return exit(err, 1)
			}
			defer s.Close()

			err = s.Serve(ctx)
			if err != nil {
				return exit(err, 1)
			}

			return nil
		},
	}

	cmd.Flags().StringP("config", "c", "", "Path to the configuration file")

	return cmd
}
