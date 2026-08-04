package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"highway-to-golang/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))

	if err := app.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}
