package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Coastal-Programs/notion-cli/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	err := cli.ExecuteContext(ctx)
	os.Exit(cli.ExitCode(err))
}
