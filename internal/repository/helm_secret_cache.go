package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"gorm.io/gorm"
)

type HelmSecretCacheRepository struct {
	db *gorm.DB
}

// NewHelmSecretCacheRepository returns pointer to repo along with the db
func NewHelmSecretCacheRepository(db *gorm.DB) *HelmSecretCacheRepository {
	return &HelmSecretCacheRepository{db}
}

func (r *HelmSecretCacheRepository) CreateHelmSecretCache(cache *models.HelmSecretCache) (*models.HelmSecretCache, error) {
	if err := r.db.Create(cache).Error; err != nil {
		return nil, err
	}

	return cache, nil
}

func (r *HelmSecretCacheRepository) ListHelmSecretCachesForRevision(revision, name, namespace string) ([]*models.HelmSecretCache, error) {
	var caches []*models.HelmSecretCache

	if err := r.db.Where("revision = ? AND name = ? AND namespace = ?", revision, name, namespace).
		Order("timestamp desc").
		Find(&caches).Error; err != nil {
		return nil, err
	}

	return caches, nil
}
