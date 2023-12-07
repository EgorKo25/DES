package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/EgorKo25/DES/internal/cache"
	"github.com/EgorKo25/DES/internal/config"
	"github.com/EgorKo25/DES/internal/logger"
	"github.com/EgorKo25/DES/internal/server/service"
	"github.com/EgorKo25/DES/internal/workers"

	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	logs := logger.NewLogger()
	log := logs["logger"]
	log.Info("logger was init successful")

	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal("cannot load config",
			zap.NamedError("error", err),
		)
	}

	cacher := cache.NewCache(ctx, cfg.CacheClearInterval)
	log.Info("cacher init successful")

	channel := make(chan chan []byte, cfg.ChannelSize)

	_ = workers.NewWorkerPull(ctx, channel,
		cfg.WorkerConfig.MaxWorkers,
		cfg.WorkerConfig.TimeoutConnection,
		cfg.WorkerConfig.MaxTimeForResponse,
		cfg.WorkerConfig.Authentication.Login,
		cfg.WorkerConfig.Authentication.Password,
		log,
		logs["http"],
	)

	log.Info("worker pull init successful")

	server := service.NewExtServer(
		channel,
		log,
		logs["grpc"],
		cacher,
	)

	s, err := server.StartServer(cfg.ServiceConfig.IP, cfg.ServiceConfig.PORT)
	if err != nil {
		log.Fatal("cannot start grpc server",
			zap.NamedError("error", err),
		)
	}
	log.Info("server started",
		zap.String("host", cfg.ServiceConfig.IP),
		zap.String("port", cfg.ServiceConfig.PORT),
	)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	sig := <-interrupt
	log.Info("Received signal", zap.String("signal", sig.String()))
	s.GracefulStop()
}
