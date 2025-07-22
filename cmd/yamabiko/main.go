package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:          "yamabiko",
		Short:        "Yamabiko is skk dictionary server",
		SilenceUsage: true,
	}

	cmd.AddCommand(serverCmd())

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer cancel()

	if err := cmd.ExecuteContext(ctx); err != nil {
		code := 1
		var exitErr exitCoder
		if errors.As(err, &exitErr) {
			code = exitErr.ExitCode()
		}
		os.Exit(code)
	}
}
