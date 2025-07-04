package main

import (
	"context"

	"github.com/kechako/yamabiko/config"
	"github.com/kechako/yamabiko/server"
	"github.com/urfave/cli/v3"
)

func serverCmd(ctx context.Context, cmd *cli.Command) error {
	cfgPath := cmd.String("config")
	if cfgPath == "" {
		var err error
		cfgPath, err = config.FindConfigFile()
		if err != nil {
			return cli.Exit(err, 1)
		}
	}

	cfg, err := config.LoadFile(cfgPath)
	if err != nil {
		return cli.Exit(err, 1)
	}

	s, err := server.New(cfg)
	if err != nil {
		return cli.Exit(err, 1)
	}
	defer s.Close()

	err = s.Serve(ctx)
	if err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}
