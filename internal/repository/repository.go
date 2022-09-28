package repository

import "gorm.io/gorm"

type Query struct {
	Offset uint
	Limit  uint
}

type Repository struct {
	Incident   *IncidentRepository
	Event      *EventRepository
	EventCache *EventCacheRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Incident:   NewIncidentRepository(db),
		Event:      NewEventRepository(db),
		EventCache: NewEventCacheRepository(db),
	}
}
