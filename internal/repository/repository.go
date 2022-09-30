package repository

import "gorm.io/gorm"

type Query struct {
	Offset uint
	Limit  uint
}

type Repository struct {
	Incident   *IncidentRepository
	Event      *IncidentEventRepository
	EventCache *EventCacheRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Incident:   NewIncidentRepository(db),
		Event:      NewIncidentEventRepository(db),
		EventCache: NewEventCacheRepository(db),
	}
}
