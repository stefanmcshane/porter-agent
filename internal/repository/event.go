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

func (r *EventRepository) CreateEvent(event *models.Event) (*models.Event, error) {
	if err := r.db.Create(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) ReadEvent(id uint) (*models.Event, error) {
	event := &models.Event{}

	if err := r.db.Where("id = ?", id).First(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) UpdateEvent(event *models.Event) (*models.Event, error) {
	if err := r.db.Save(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r *EventRepository) ListEvents(
	filter *utils.ListIncidentsFilter,
	opts ...utils.QueryOption,
) ([]*models.Event, *utils.PaginatedResult, error) {
	var events []*models.Event

	db := r.db.Model(&models.Event{})

	if filter.ReleaseName != nil {
		db = db.Where("release_name = ?", *filter.ReleaseName)
	}

	if filter.ReleaseNamespace != nil {
		db = db.Where("release_namespace = ?", *filter.ReleaseNamespace)
	}

	paginatedResult := &utils.PaginatedResult{}

	db = db.Scopes(utils.Paginate(opts, db, paginatedResult))

	if err := db.Find(&events).Error; err != nil {
		return nil, nil, err
	}

	return events, paginatedResult, nil
}

func (r *EventRepository) DeleteEvent(uid string) error {
	incident := &models.Event{}

	if err := r.db.Where("unique_id = ?", uid).First(incident).Error; err != nil {
		return err
	}

	if err := r.db.Delete(incident).Error; err != nil {
		return err
	}

	return nil
}
