package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsServerBuilder struct {
	reg  *prometheus.Registry
	addr string

	tlsEnabled bool
	tlsConfig  []func(*tls.Config)
}

type MetricsServer struct {
	*http.Server
}

func NewMetricsServerBuilder() *MetricsServerBuilder {
	return &MetricsServerBuilder{}
}

func (ms *MetricsServerBuilder) WithRegistry(reg *prometheus.Registry) *MetricsServerBuilder {
	ms.reg = reg
	return ms
}

func (ms *MetricsServerBuilder) WithTLS(flag bool) *MetricsServerBuilder {
	ms.tlsEnabled = flag
	return ms
}

func (ms *MetricsServerBuilder) WithTLSConfig(funcs ...func(*tls.Config)) *MetricsServerBuilder {
	ms.tlsConfig = funcs
	return ms
}

func (ms *MetricsServerBuilder) WithAddress(addr string) *MetricsServerBuilder {
	ms.addr = addr
	return ms
}

func (ms *MetricsServerBuilder) Build() MetricsServer {
	h := http.NewServeMux()
	h.Handle("/metrics", promhttp.HandlerFor(ms.reg, promhttp.HandlerOpts{
		Registry: ms.reg,
	}))

	server := &http.Server{
		Addr:        ms.addr,
		Handler:     h,
		IdleTimeout: 30 * time.Second,
		ReadTimeout: 3 * time.Second,
	}

	if ms.tlsEnabled {
		server.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12,
		}

		for _, fun := range ms.tlsConfig {
			fun(server.TLSConfig)
		}
	}

	return MetricsServer{
		Server: server,
	}
}

func (ms *MetricsServer) Serve(ctx context.Context) {
	var err error
	logger := getLoggerFromContext(ctx)

	// HTTP Server graceful shutdown
	go func() {
		<-ctx.Done()

		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		//nolint:contextcheck
		if err := ms.Server.Shutdown(sctx); err != nil {
			logger.Error("error gracefully shutting down metrics server", "error", err)
			os.Exit(1)
		}
	}()

	if ms.Server.TLSConfig != nil {
		logger.Info("Serving metrics over https", "address", ms.Server.Addr)
		err = ms.Server.ListenAndServeTLS("", "")
	} else {
		logger.Info("Serving metrics over http", "address", ms.Server.Addr)
		err = ms.Server.ListenAndServe()
	}

	if errors.Is(err, http.ErrServerClosed) {
		logger.Info("gracefully shutting down metrics server")
	} else {
		logger.Error("metrics server closed unexpectedly", "error", err)
	}
}
