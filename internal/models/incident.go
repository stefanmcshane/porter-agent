package models

import "gorm.io/gorm"

type Incident struct {
	gorm.Model

	UniqueID string `gorm:"unique"`

	ReleaseName string
	ChartName   string

	Events []Event
}
