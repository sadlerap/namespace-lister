package main

import (
	openshiftapi "github.com/openshift/api"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func buildImpersonatingClient(cfg *rest.Config, username string) (client.Client, error) {
	cfg = rest.CopyConfig(cfg)
	cfg.Impersonate.UserName = username

	s := runtime.NewScheme()
	if err := openshiftapi.Install(s); err != nil {
		return nil, err
	}
	return client.New(cfg, client.Options{Scheme: s})
}
