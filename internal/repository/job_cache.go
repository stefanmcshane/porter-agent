package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"gorm.io/gorm"
)

type JobCacheRepository struct {
	db *gorm.DB
}

// NewJobCacheRepository returns pointer to repo along with the db
func NewJobCacheRepository(db *gorm.DB) *JobCacheRepository {
	return &JobCacheRepository{db}
}

func (r *JobCacheRepository) CreateJobCache(cache *models.JobCache) (*models.JobCache, error) {
	if err := r.db.Create(cache).Error; err != nil {
		return nil, err
	}

	return cache, nil
}

func (r *JobCacheRepository) ListJobCaches(podName, podNamespace, reason string) ([]*models.JobCache, error) {
	var caches []*models.JobCache

	if err := r.db.Where("pod_name = ? AND pod_namespace = ? AND reason = ?", podName, podNamespace, reason).
		Order("timestamp desc").
		Find(&caches).Error; err != nil {
		return nil, err
	}

	return caches, nil
}
