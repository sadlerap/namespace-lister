package main

import (
	"log/slog"
	"net/http"

	"k8s.io/client-go/rest"
)

type getNamespaceHandler struct {
	cfg *rest.Config
	log *slog.Logger
}

func newGetNamespaceHandler(cfg *rest.Config, log *slog.Logger) http.Handler {
	return &getNamespaceHandler{
		cfg: cfg,
		log: log,
	}
}

func (h *getNamespaceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
