package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:  "yamabiko",
		Usage: "Yamabiko is skk dictionary server",
		Commands: []*cli.Command{
			{
				Name:   "server",
				Usage:  "Run the Yamabiko server",
				Action: serverCmd,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "config",
						Aliases: []string{"c"},
						Usage:   "Path to the configuration file",
					},
				},
			},
		},
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if err := cmd.Run(ctx, os.Args); err != nil {
		code := 1
		var exitCoder cli.ExitCoder
		if errors.As(err, &exitCoder) {
			code = exitCoder.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(code)
	}
}
