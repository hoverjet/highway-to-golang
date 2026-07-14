package main

import (
	"log/slog"
	"os"

	"highway-to-golang/internal/app"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if err := app.Run(logger); err != nil {
		logger.Error("application failed", "error", err)
		os.Exit(1)
	}
}
