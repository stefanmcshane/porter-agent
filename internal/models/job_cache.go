package models

import (
	"time"

	"gorm.io/gorm"
)

type JobCache struct {
	gorm.Model

	PodName      string
	PodNamespace string
	Timestamp    *time.Time
	Reason       string
}
