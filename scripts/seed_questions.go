package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log/slog"
	"os"

	"quizbattle/internal/logger"
	"quizbattle/internal/storage"
	"quizbattle/internal/types"
)

const defaultBankPath = "questions/bank.json"

func main() {
	addr := flag.String("addr", "localhost:6379", "Redis 地址,形如 host:port")
	password := flag.String("password", "", "Redis 密码,本地无密码时留空")
	flag.Parse()

	baseLog := logger.New()

	rds, err := storage.New(*addr, *password, baseLog)
	if err != nil {
		baseLog.Error("connect redis fail", slog.String("addr", *addr), slog.Any("err", err))
		os.Exit(1)
	}
	defer func() {
		if cerr := rds.Close(); cerr != nil {
			baseLog.Error("close redis fail", slog.Any("err", cerr))
		}
	}()

	data, err := os.ReadFile(defaultBankPath)
	if err != nil {
		baseLog.Error("read bank file fail", slog.String("path", defaultBankPath), slog.Any("err", err))
		os.Exit(1)
	}

	var qs []types.Question
	if err := json.Unmarshal(data, &qs); err != nil {
		baseLog.Error("unmarshal bank fail", slog.String("path", defaultBankPath), slog.Any("err", err))
		os.Exit(1)
	}

	var (
		inserted int
		skipped  int
		failed   int
	)

	for _, q := range qs {
		ctx := logger.WithLogID(context.Background(), "seed-"+q.ID)
		log := logger.FromContext(ctx, baseLog)

		err := rds.SaveQuestion(ctx, q)
		switch {
		case err == nil:
			log.Info("inserted", slog.String("questionID", q.ID))
			inserted++
		case errors.Is(err, storage.ErrQuestionExists):
			log.Info("skipped (already exists)", slog.String("questionID", q.ID))
			skipped++
		default:
			log.Error("save failed", slog.String("questionID", q.ID), slog.Any("err", err))
			failed++
		}
	}

	baseLog.Info("seed done",
		slog.String("bank", defaultBankPath),
		slog.Int("inserted", inserted),
		slog.Int("skipped", skipped),
		slog.Int("failed", failed),
		slog.Int("total", len(qs)),
	)

	if failed > 0 {
		os.Exit(1)
	}
}
