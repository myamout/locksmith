package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type ServerConfig struct {
	Port int
}

type LocksmithServer struct {
	ctx context.Context
	cfg ServerConfig
}

func NewLocksmithServer(ctx context.Context, cfg ServerConfig) *LocksmithServer {
	return &LocksmithServer{
		ctx: ctx,
		cfg: cfg,
	}
}

func (ls *LocksmithServer) Start() error {
	r := ls.getRouter()

	server := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", ls.cfg.Port),
		Handler: r,
	}

	serverCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	<-serverCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return err
	}

	return nil
}

func (ls *LocksmithServer) getRouter() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	return r
}
