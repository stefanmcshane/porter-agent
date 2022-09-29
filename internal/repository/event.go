package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/utils"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository returns pointer to repo along with the db
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db}
}

func (r *EventRepository) CreateEvent(event *models.IncidentEvent) (*models.IncidentEvent, error) {
	if err := r.db.Create(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) ReadEvent(uid string) (*models.IncidentEvent, error) {
	event := &models.IncidentEvent{}

	if err := r.db.Preload("Logs").Where("unique_id = ?", uid).First(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) ListEvents(filter *utils.ListIncidentEventsFilter, opts ...utils.QueryOption) ([]*models.IncidentEvent, error) {
	var events []*models.IncidentEvent

	db := r.db.Scopes(utils.Paginate(opts))

	if filter.IncidentID != nil {
		db = db.Where("incident_id = ?", *filter.IncidentID)
	}

	if filter.PodName != nil {
		db = db.Where("pod_name = ?", *filter.PodName)
	}

	if filter.PodNamespace != nil {
		db = db.Where("pod_namespace = ?", *filter.PodNamespace)
	}

	if filter.Summary != nil {
		db = db.Where("summary = ?", *filter.Summary)
	}

	if filter.IsPrimaryCause != nil {
		db = db.Where("is_primary_cause = ?", *filter.IsPrimaryCause)
	}

	if err := db.Find(&events).Error; err != nil {
		return nil, err
	}

	return events, nil
}

func (r *EventRepository) DeleteEvent(uid string) error {
	event := &models.IncidentEvent{}

	if err := r.db.Where("unique_id = ?", uid).First(event).Error; err != nil {
		return err
	}

	if err := r.db.Delete(event).Error; err != nil {
		return err
	}

	return nil
}
