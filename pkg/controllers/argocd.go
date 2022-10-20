package controllers

import (
	"context"
	"crypto/tls"
	"net/http"

	argocd "github.com/stefanmcshane/go-argocd/argocd"
	"github.com/stefanmcshane/go-argocd/argocd/client/application_service"
)

type ArgoCD struct {
	// Client argocdclient.
}

func NewArgoCDClient() (ArgoCD, error) {
	var cli ArgoCD

	argoConf := argocd.ArgoCDClientOpts{
		Host:      "localhost",
		Port:      "8080",
		APIToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJhcmdvY2QiLCJzdWIiOiJwb3J0ZXI6YXBpS2V5IiwibmJmIjoxNjY2MjIyODc5LCJpYXQiOjE2NjYyMjI4NzksImp0aSI6IjAwN2Y2NzA1LWE2ZmMtNDI3My1iNjlmLWEwZTFiMzcyOGQ1OCJ9.5t_P5mPMksbqZIj59gn-ToFpsSRDpQQZSBylNVHpLLQ",
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Debug:     false,
	}

	argoClient, err := argocd.NewArgoCDWithAPIKey(argoConf)
	if err != nil {
		return cli, err
	}

	params := application_service.ApplicationServiceListParams{
		Context: context.Background(),
	}
	resp, err := argoClient.ApplicationService.ApplicationServiceList(&params)
	if err != nil {
		return cli, err
	}

	_ = resp
	return cli, nil
}
