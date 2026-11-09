package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/lewtec/lewkit/x/cmd"
	_ "github.com/lewtec/lewkit/x/driver/exec/prelude"
	"github.com/lewtec/lewkit/x/logging"
	"github.com/lucasew/env-mcp/internal/cli"
)

func main() {
	slog.SetDefault(slog.New(logging.NewHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, os.Args[1:]); err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	app, err := cmd.Parse[cmd.App[cli.Root]](cli.InsertDashes(args)...)
	if err != nil {
		return err
	}
	return app.Run(ctx)
}
