package repository

import (
	"testing"

	"time"

	"github.com/porter-dev/porter-agent/internal/models"
)

func TestReadIncident(t *testing.T) {
	tester := &tester{
		dbFileName: "./incident_test.db",
	}

	setupTestEnv(tester, t)
	defer cleanup(tester, t)

	oneHourAgo := time.Now().Add(-1 * time.Hour)
	twoHourAgo := time.Now().Add(-2 * time.Hour)
	threeHourAgo := time.Now().Add(-3 * time.Hour)

	incident := &models.Incident{
		UniqueID:         "very-unique-id",
		ReleaseName:      "release",
		ReleaseNamespace: "namespace",
		Events: []models.IncidentEvent{
			{
				UniqueID: "unique-event-id-1",
				LastSeen: &oneHourAgo,
			},
			{
				UniqueID: "unique-event-id-3",
				LastSeen: &threeHourAgo,
			},
			{
				UniqueID: "unique-event-id-2",
				LastSeen: &twoHourAgo,
			},
		},
	}

	incident, err := tester.repo.Incident.CreateIncident(incident)

	if err != nil {
		t.Fatalf("Expected no error after creating incident, got %v", err)
	}

	incident, err = tester.repo.Incident.ReadIncident(incident.UniqueID)

	if err != nil {
		t.Fatalf("Expected no error after reading incident, got %v", err)
	}
}
