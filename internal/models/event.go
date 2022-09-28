package models

import "gorm.io/gorm"

type SeverityType string

const (
	SeverityCritical SeverityType = "critical"
	SeverityNormal   SeverityType = "normal"
)

type Event struct {
	gorm.Model

	UniqueID string `gorm:"unique"`

	Resource  string
	Namespace string
	Severity  SeverityType

	Logs []Log
}
