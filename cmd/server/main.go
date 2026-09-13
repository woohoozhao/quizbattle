package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"quizbattle/internal/logger"
	"quizbattle/internal/wsserver"
	"syscall"
)

func main() {
	log := logger.New()
	srv := wsserver.New(":17000", log)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := srv.Run(ctx); err != nil {
		log.Error("server fail", slog.Any("err", err))
		os.Exit(1)
	}
}
