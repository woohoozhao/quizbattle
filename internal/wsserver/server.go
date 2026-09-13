package wsserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync/atomic"

	"github.com/coder/websocket"
)

type Server struct {
	addr string
	log  *slog.Logger
	srv  *http.Server
	next atomic.Uint64
}

func New(addr string, log *slog.Logger) *Server {
	s := &Server{
		addr: addr,
		log:  log,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handle)
	s.srv = &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	return s
}

func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.srv.Shutdown(context.Background())
	}()
	s.log.Info("ws server listening", slog.String("addr", s.addr))
	if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})

	if err != nil {
		s.log.Error("ws accept fail", slog.Any("err", err))
		return
	}

	id := s.next.Add(1)
	runSession(r.Context(), conn, id, s.log)
}
