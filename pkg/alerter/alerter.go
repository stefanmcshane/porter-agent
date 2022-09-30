package alerter

import (
	"time"

	"github.com/porter-dev/porter-agent/api/server/types"
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/repository"
	"github.com/porter-dev/porter-agent/pkg/httpclient"
)

type Alerter struct {
	Client     *httpclient.Client
	Repository *repository.Repository
}

func (a *Alerter) HandleIncident(incident *models.Incident) error {
	switch incident.Severity {
	case types.SeverityCritical:
		if a.shouldAlertCritical(incident) {
			err := a.Client.NotifyNew(incident.ToAPITypeMeta())
			if err != nil {
				return err
			}

			return a.updateLastAlerted(incident)
		}

		return nil
	case types.SeverityNormal:
		if a.shouldAlertNormal(incident) {
			err := a.Client.NotifyNew(incident.ToAPITypeMeta())

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
		return a.Client.NotifyResolved(incident.ToAPITypeMeta())
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
