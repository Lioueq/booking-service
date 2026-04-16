package container

import (
	"booking/internal/config"
	"booking/internal/logger"
	"booking/internal/repository/storage"
	"context"
	"log/slog"
	"os"
)

type Container struct {
	Config  *config.Config
	Logger  *slog.Logger
	Storage *storage.Storage
}

func NewContainer(ctx context.Context) *Container {
	cfg := config.MustLoad()
	log := logger.NewLogger(cfg.Env)
	pgstorage, err := storage.NewStorage(ctx)
	if err != nil {
		log.Error("failed to connect to database")
		os.Exit(2)
	}
	return &Container{
		Config:  cfg,
		Logger:  log,
		Storage: pgstorage,
	}
}
