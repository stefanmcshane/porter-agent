//go:generate mockgen -source=event.go -destination=mocks/event.go
package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/utils"
	"gorm.io/gorm"
)

// EventRepository wraps all actions related to a stored porter event
type EventRepository interface {
	CreateEvent(event *models.Event) (*models.Event, error)
	ReadEvent(id uint) (*models.Event, error)
	UpdateEvent(event *models.Event) (*models.Event, error)
	DeleteEvent(uid string) error
	ListEvents(filter *utils.ListEventsFilter, opts ...utils.QueryOption) ([]*models.Event, *utils.PaginatedResult, error)
}

type eventRepository struct {
	db *gorm.DB
}

// NewEventRepository returns pointer to repo along with the db
func NewEventRepository(db *gorm.DB) EventRepository {
	return eventRepository{db}
}

func (r eventRepository) CreateEvent(event *models.Event) (*models.Event, error) {
	if err := r.db.Create(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r eventRepository) ReadEvent(id uint) (*models.Event, error) {
	event := &models.Event{}

	if err := r.db.Where("id = ?", id).First(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r eventRepository) UpdateEvent(event *models.Event) (*models.Event, error) {
	if err := r.db.Save(event).Error; err != nil {
		return nil, err
	}

	return event, nil
}

func (r eventRepository) ListEvents(
	filter *utils.ListEventsFilter,
	opts ...utils.QueryOption,
) ([]*models.Event, *utils.PaginatedResult, error) {
	var events []*models.Event

	db := r.db.Model(&models.Event{})

	if filter.Type != nil {
		db = db.Where("type = ?", *filter.Type)
	}

	if filter.ReleaseName != nil {
		db = db.Where("release_name = ?", *filter.ReleaseName)
	}

	if filter.ReleaseNamespace != nil {
		db = db.Where("release_namespace = ?", *filter.ReleaseNamespace)
	}

	if filter.AdditionalQueryMeta != nil {
		db = db.Where("additional_query_meta = ?", *filter.AdditionalQueryMeta)
	}

	paginatedResult := &utils.PaginatedResult{}

	db = db.Scopes(utils.Paginate(opts, db, paginatedResult))

	if err := db.Find(&events).Error; err != nil {
		return nil, nil, err
	}

	return events, paginatedResult, nil
}

func (r eventRepository) DeleteEvent(uid string) error {
	incident := &models.Event{}

	if err := r.db.Where("unique_id = ?", uid).First(incident).Error; err != nil {
		return err
	}

	if err := r.db.Delete(incident).Error; err != nil {
		return err
	}

	return nil
}
