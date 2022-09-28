package models

import "gorm.io/gorm"

type Log struct {
	gorm.Model

	UniqueID string `gorm:"unique"`
}
