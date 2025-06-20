package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"http-from-tcp/internal/server"
)

const port = 42069

func main() {
	loggingLevel := new(slog.LevelVar)
	slogOpts := slog.HandlerOptions{Level: loggingLevel}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slogOpts))

	slog.SetDefault(logger)
	loggingLevel.Set(slog.LevelInfo)

	if err := run(); err != nil {
		slog.Error(err.Error())

		os.Exit(1)
	}
}

func run() error {
	// Set up logging

	server, err := server.Serve(port)
	if err != nil {
		return fmt.Errorf("error starting the server: %w", err)
	}
	defer server.Close()

	// TODO: Use context instead
	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer stop()

	<-ctx.Done()
	stop()

	slog.Info("Server gracefully stopped.")

	return nil
}
