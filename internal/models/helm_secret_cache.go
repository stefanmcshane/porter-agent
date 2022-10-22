package models

import (
	"time"

	"gorm.io/gorm"
)

type HelmSecretCache struct {
	gorm.Model

	Name      string
	Namespace string
	Revision  string
	Timestamp *time.Time
}
