package main

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	openshiftapiv1 "github.com/openshift/api/project/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var _ http.Handler = &listNamespacesHandler{}

type listNamespacesHandler struct {
	cfg *rest.Config
	log *slog.Logger
}

func newListNamespacesHandler(cfg *rest.Config, log *slog.Logger) http.Handler {
	return &listNamespacesHandler{
		cfg: cfg,
		log: log,
	}
}

func (h *listNamespacesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// build impersonating client
	cli, err := buildImpersonatingClient(h.cfg, r.Header.Get("X-Username"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// retrieve projects as the user
	pp := openshiftapiv1.ProjectList{}
	if err := cli.List(r.Context(), &pp); err != nil {
		serr := &kerrors.StatusError{}
		if errors.As(err, &serr) {
			w.WriteHeader(int(serr.Status().Code))
			w.Write([]byte(serr.Error()))
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	// map projects to namespaces
	nr := len(pp.Items)
	nn := make([]corev1.Namespace, nr, nr)
	for _, p := range pp.Items {
		n := convertProjectToNamespace(&p)
		nn = append(nn, n)
	}

	// build response
	// for PoC limited to JSON
	l := corev1.NamespaceList{TypeMeta: metav1.TypeMeta{Kind: "List", APIVersion: "v1"}, Items: nn}
	b, err := json.Marshal(l)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	w.Write(b)
}
