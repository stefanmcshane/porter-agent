package config

import (
	"os"

	"github.com/porter-dev/porter-agent/internal/envconf"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/pkg/argocd"
	"github.com/porter-dev/porter-agent/pkg/logstore"
	"github.com/porter-dev/porter/api/server/shared/apierrors/alerter"
	"github.com/porter-dev/porter/pkg/logger"
)

type Config struct {
	// Logger for logging
	Logger *logger.Logger

	// Alerter to send alerts to a third-party aggregator
	Alerter alerter.Alerter

	Repository *repository.Repository

	LogStore logstore.LogStore

	ArgoCD argocd.ArgoCDConfig
}

func GetConfig(envConf *envconf.EnvDecoderConf, repo *repository.Repository, ls logstore.LogStore) (*Config, error) {
	var err error

	res := &Config{
		Logger:     logger.New(envConf.Debug, os.Stdout),
		Alerter:    alerter.NoOpAlerter{},
		Repository: repo,
	}

	res.Alerter = alerter.NoOpAlerter{}

	if envConf.SentryDSN != "" {
		res.Alerter, err = alerter.NewSentryAlerter(envConf.SentryDSN, envConf.SentryEnv)

		if err != nil {
			return nil, err
		}
	}

	res.LogStore = ls

	return res, nil
}
