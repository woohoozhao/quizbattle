package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"quizbattle/internal/logger"
	"quizbattle/pkg/protocol"
	"time"

	"github.com/coder/websocket"
)

func main() {
	addr := flag.String("addr", "ws://localhost:17000", "ws server")
	name := flag.String("name", "anon", "player name")
	flag.Parse()

	log := logger.New()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, *addr, nil)
	if err != nil {
		log.Error("dial fail", slog.Any("err", err))
		os.Exit(1)
	}
	defer func() {
		err := conn.Close(websocket.StatusNormalClosure, "")
		if err != nil {
			log.Error("close fail", slog.Any("err", err))
		}
	}()

	join, _ := json.Marshal(protocol.Envelope{Type: protocol.TypeJoinQueue, Name: *name})
	if err := conn.Write(ctx, websocket.MessageText, join); err != nil {
		log.Error("join send fail", slog.Any("err", err))
		os.Exit(1)
	}

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			log.Info("read end", slog.Any("err", err))
			return
		}
		fmt.Printf("recv: %s\n", data)
	}
}
