package main

import (
	"cmp"
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/client-go/rest"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func main() {
	l := buildLogger()
	if err := run(l); err != nil {
		l.Error("error running the server", "error", err)
		os.Exit(1)
	}
}

func run(l *slog.Logger) error {
	log.SetLogger(logr.FromSlogHandler(l.Handler()))

	// get k8s rest config
	cfg, err := config.GetConfig()
	if err != nil {
		return err
	}

	// setup context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// build http server
	s := buildServer(cfg, l)

	// HTTP Server graceful shutdown
	go func() {
		<-ctx.Done()

		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if err := s.Shutdown(sctx); err != nil {
			l.Error("error gracefully shutting down the HTTP server", "error", err)
			os.Exit(1)
		}
	}()

	// start server
	return s.ListenAndServe()
}

func buildServer(cfg *rest.Config, l *slog.Logger) *http.Server {
	// configure the server
	h := http.NewServeMux()
	h.Handle("GET /api/v1/namespaces", newListNamespacesHandler(rest.CopyConfig(cfg), l))
	h.Handle("GET /api/v1/namespaces/{name}", newGetNamespaceHandler(rest.CopyConfig(cfg), l))
	return &http.Server{
		Addr:              cmp.Or(os.Getenv("ADDRESS"), DefaultAddr),
		Handler:           h,
		ReadHeaderTimeout: 3 * time.Second,
	}
}
