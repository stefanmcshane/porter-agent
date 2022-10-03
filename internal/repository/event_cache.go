package repository

import (
	"time"

	"github.com/porter-dev/porter-agent/internal/models"
	"gorm.io/gorm"
)

type EventCacheRepository struct {
	db *gorm.DB
}

// NewEventCacheRepository returns pointer to repo along with the db
func NewEventCacheRepository(db *gorm.DB) *EventCacheRepository {
	return &EventCacheRepository{db}
}

func (r *EventCacheRepository) CreateEventCache(cache *models.EventCache) (*models.EventCache, error) {
	if err := r.db.Create(cache).Error; err != nil {
		return nil, err
	}

	return cache, nil
}

func (r *EventCacheRepository) ListEventCachesForEvent(uid string) ([]*models.EventCache, error) {
	var caches []*models.EventCache

	if err := r.db.Where("event_uid = ? AND timestamp >= ?", uid, time.Now().Add(-time.Hour)).
		Order("timestamp desc").
		Find(&caches).Error; err != nil {
		return nil, err
	}

	return caches, nil
}

func (r *EventCacheRepository) ListEventCachesForPod(name, namespace string) ([]*models.EventCache, error) {
	var caches []*models.EventCache

	if err := r.db.Where("pod_name = ? AND pod_namespace AND timestamp >= ?", name, namespace, time.Now().Add(-time.Hour)).
		Order("timestamp desc").
		Find(&caches).Error; err != nil {
		return nil, err
	}

	return caches, nil
}
