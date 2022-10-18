package repository

import "gorm.io/gorm"

type Repository struct {
	Alert           *AlertRepository
	Incident        *IncidentRepository
	IncidentEvent   *IncidentEventRepository
	EventCache      *EventCacheRepository
	Event           *EventRepository
	HelmSecretCache *HelmSecretCacheRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		Alert:           NewAlertRepository(db),
		Incident:        NewIncidentRepository(db),
		IncidentEvent:   NewIncidentEventRepository(db),
		EventCache:      NewEventCacheRepository(db),
		Event:           NewEventRepository(db),
		HelmSecretCache: NewHelmSecretCacheRepository(db),
	}
}
