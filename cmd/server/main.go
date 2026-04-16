package main

import (
	"booking/internal/app"
	"booking/internal/container"
	"context"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	container := container.NewContainer(ctx)
	app := app.NewApp(container)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := app.Run(ctx); err != nil {
			container.Logger.Error("failed to start server")
			os.Exit(1)
		}
	}()

	<-sigChan
	container.Logger.Info("received shutdown signal")
	cancel()
	container.Storage.Close()
	container.Logger.Info("shutting down gracefully")
}
