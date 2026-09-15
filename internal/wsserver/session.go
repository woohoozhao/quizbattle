package wsserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"quizbattle/pkg/protocol"

	"github.com/coder/websocket"
)

func runSession(ctx context.Context, conn *websocket.Conn, id uint64, log *slog.Logger) {
	slogLog := log.With(slog.Uint64("session", id))
	defer func() {
		err := conn.Close(websocket.StatusNormalClosure, "bye")
		if err != nil {
			slog.Error(err.Error())
		}
	}()

	slogLog.Info("connected")

	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			slogLog.Error("read end", slog.Any("err", err))
			return
		}
		var env protocol.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			sendErr(ctx, conn, protocol.CodeBadFrame, err.Error())
			continue
		}

		switch env.Type {
		case protocol.TypeJoinQueue:
			slogLog.Info("joined", slog.String("name", env.Name))
		case protocol.TypeHeartbeat:
		default:
			sendErr(ctx, conn, protocol.CodeUnknownType, env.Type)
		}
	}
}

func sendErr(ctx context.Context, conn *websocket.Conn, code int, msg string) {
	data, _ := json.Marshal(protocol.Envelope{Type: protocol.TypeError, Code: code, Msg: msg})
	err := conn.Write(ctx, websocket.MessageText, data)
	if err != nil {
		slog.Error("send error", slog.Any("err", err))
	}
}
