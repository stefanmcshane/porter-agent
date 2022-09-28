package repository

import (
	"github.com/porter-dev/porter-agent/pkg/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB, debug bool) error {
	instanceDB := db

	if debug {
		instanceDB = instanceDB.Debug()
	}

	return instanceDB.AutoMigrate(
		&models.Incident{},
	)
}
