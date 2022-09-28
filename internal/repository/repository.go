package repository

import "gorm.io/gorm"

type Query struct {
	Offset uint
	Limit  uint
}

type Repository struct {
	Incident *IncidentRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Incident: NewIncidentRepository(db),
	}
}
