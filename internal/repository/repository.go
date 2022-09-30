package repository

import "gorm.io/gorm"

type Repository struct {
	Incident      *IncidentRepository
	IncidentEvent *IncidentEventRepository
	EventCache    *EventCacheRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Incident:      NewIncidentRepository(db),
		IncidentEvent: NewIncidentEventRepository(db),
		EventCache:    NewEventCacheRepository(db),
	}
}
