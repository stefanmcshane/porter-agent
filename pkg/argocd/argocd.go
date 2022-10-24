//go:generate mockgen -source argocd.go -destination mocks/argocd.go
package argocd

import (
	"context"
	"fmt"
	"net/http"

	argocd "github.com/stefanmcshane/go-argocd/argocd"
	"github.com/stefanmcshane/go-argocd/argocd/client"
)

// ArgoCD abstracts the underlying module to allow for easier testing
type ArgoCD interface {
	Sync(ctx context.Context, app Application) error
	Rollback(ctx context.Context, app Application) error
}

// ArgoCDConfig contains all details for connecting to ArgoCD
type ArgoCDConfig struct {
	Host      string
	Port      string
	Username  string
	Password  string
	Scheme    string
	Transport *http.Transport
	Debug     bool
}

// argoCDClient wraps the ArgoCD go module
type argoCDClient struct {
	client *client.ConsolidateServices
}

// NewArgoCDClient creates an implementation of ArgoCD
func NewArgoCDClient(ctx context.Context, conf ArgoCDConfig) (ArgoCD, error) {
	var cli argoCDClient

	argoConf := argocd.ArgoCDClientOpts{
		Host:      conf.Host,
		Port:      conf.Port,
		Username:  conf.Username,
		Password:  conf.Password,
		Scheme:    conf.Scheme,
		Debug:     conf.Debug,
		Transport: conf.Transport,
	}

	argoClient, err := argocd.NewArgoCD(ctx, argoConf)
	if err != nil {
		return cli, fmt.Errorf("unable to connect to argo: %v", err)
	}
	cli.client = argoClient

	return cli, nil
}
