package models

import (
	"time"

	"gorm.io/gorm"
)

type EventCache struct {
	gorm.Model

	EventUID string

	PodName      string
	PodNamespace string
	Timestamp    *time.Time
	Revision     string
}
