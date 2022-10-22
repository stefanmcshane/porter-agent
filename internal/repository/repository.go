package repository

import "gorm.io/gorm"

type Repository struct {
	DB *gorm.DB

	Alert           *AlertRepository
	Incident        *IncidentRepository
	IncidentEvent   *IncidentEventRepository
	EventCache      *EventCacheRepository
	Event           *EventRepository
	HelmSecretCache *HelmSecretCacheRepository
	JobCache        *JobCacheRepository
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{
		DB:              db,
		Alert:           NewAlertRepository(db),
		Incident:        NewIncidentRepository(db),
		IncidentEvent:   NewIncidentEventRepository(db),
		EventCache:      NewEventCacheRepository(db),
		Event:           NewEventRepository(db),
		HelmSecretCache: NewHelmSecretCacheRepository(db),
		JobCache:        NewJobCacheRepository(db),
	}
}
