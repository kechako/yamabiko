package main

import (
	"context"

	"github.com/kechako/yamabiko/config"
	"github.com/kechako/yamabiko/server"
	"github.com/urfave/cli/v3"
)

func serverCmd(ctx context.Context, cmd *cli.Command) error {
	cfg, err := config.LoadFile(cmd.String("config"))
	if err != nil {
		return err
	}

	s, err := server.New(cfg)
	if err != nil {
		return err
	}
	defer s.Close()

	return s.Serve(ctx)
}
