package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, debug bool) error {
	instanceDB := db

	if debug {
		instanceDB = instanceDB.Debug()
	}

	return instanceDB.AutoMigrate(
		&models.Alert{},
		&models.Incident{},
		&models.IncidentEvent{},
		&models.EventCache{},
		&models.Event{},
		&models.HelmSecretCache{},
		&models.JobCache{},
	)
}
