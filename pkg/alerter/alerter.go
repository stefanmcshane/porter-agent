package alerter

import (
	"strings"
	"time"

	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/logger"
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/pkg/httpclient"
)

type JobAlertConfiguration string

const (
	JobAlertConfigurationEvery JobAlertConfiguration = "every"
	JobAlertConfigurationDaily JobAlertConfiguration = "daily"
)

type AlertConfiguration struct {
	DefaultJobAlertConfiguration JobAlertConfiguration
}

type Alerter struct {
	AlertConfiguration *AlertConfiguration
	Client             *httpclient.Client
	Repository         *repository.Repository
	Logger             *logger.Logger
}

func (a *Alerter) HandleIncident(incident *models.Incident) error {
	switch incident.Severity {
	case types.SeverityCritical:
		if a.shouldAlertCritical(incident) {
			err := a.Client.NotifyNew(incident.ToAPIType())
			if err != nil {
				return err
			}

			return a.updateLastAlerted(incident)
		}

		return nil
	case types.SeverityNormal:
		if a.shouldAlertNormal(incident) {
			err := a.Client.NotifyNew(incident.ToAPIType())

			if err != nil {
				return err
			}

			return a.updateLastAlerted(incident)
		}
	}

	return nil
}

func (a *Alerter) HandleResolved(incident *models.Incident) error {
	switch incident.Severity {
	case types.SeverityCritical:
		// if this is a critical incident, alert immediately
		return a.Client.NotifyResolved(incident.ToAPIType())
	case types.SeverityNormal:
		// if this is a non-critical incident do nothing
	}

	return nil
}

// for critical incidents, alert every hour
func (a *Alerter) shouldAlertCritical(incident *models.Incident) bool {
	if incident.LastAlerted == nil {
		return true
	}

	elapsedTime := time.Now().Sub(*incident.LastAlerted)
	elapsedHours := elapsedTime.Truncate(time.Hour)

	return elapsedHours >= 1
}

// for non-critical incidents, alert every day
func (a *Alerter) shouldAlertNormal(incident *models.Incident) bool {
	if incident.LastAlerted == nil {
		return true
	}

	// if this is a job alert, check the alerter configuration
	if strings.ToLower(string(incident.InvolvedObjectKind)) == "job" && a.AlertConfiguration.DefaultJobAlertConfiguration == JobAlertConfigurationEvery {
		return true
	}

	elapsedTime := time.Now().Sub(*incident.LastAlerted)
	elapsedHours := elapsedTime.Truncate(time.Hour)

	return elapsedHours >= 24
}

func (a *Alerter) updateLastAlerted(incident *models.Incident) error {
	now := time.Now()

	incident.LastAlerted = &now
	incident, err := a.Repository.Incident.UpdateIncident(incident)

	return err
}
