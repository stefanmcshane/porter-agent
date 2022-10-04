package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/utils"
	"gorm.io/gorm"
)

type IncidentEventRepository struct {
	db *gorm.DB
}

// NewIncidentEventRepository returns pointer to repo along with the db
func NewIncidentEventRepository(db *gorm.DB) *IncidentEventRepository {
	return &IncidentEventRepository{db}
}

func (r *IncidentEventRepository) CreateEvent(event *models.IncidentEvent) (*models.IncidentEvent, error) {
	if err := r.db.Create(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r *IncidentEventRepository) ReadEvent(uid string) (*models.IncidentEvent, error) {
	event := &models.IncidentEvent{}

	if err := r.db.Where("unique_id = ?", uid).First(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r *IncidentEventRepository) ListEvents(
	filter *utils.ListIncidentEventsFilter,
	opts ...utils.QueryOption,
) ([]*models.IncidentEvent, *utils.PaginatedResult, error) {
	var events []*models.IncidentEvent

	db := r.db.Model(&models.IncidentEvent{})

	if filter.IncidentID != nil {
		db = db.Where("id = ?", *filter.IncidentID)
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

	paginatedResult := &utils.PaginatedResult{}

	db = db.Scopes(utils.Paginate(opts, db, paginatedResult))

	if err := db.Find(&events).Error; err != nil {
		return nil, nil, err
	}

	return events, paginatedResult, nil
}

func (r *IncidentEventRepository) DeleteEvent(uid string) error {
	event := &models.IncidentEvent{}

	if err := r.db.Where("unique_id = ?", uid).First(event).Error; err != nil {
		return err
	}

	if err := r.db.Delete(event).Error; err != nil {
		return err
	}

	return nil
}
