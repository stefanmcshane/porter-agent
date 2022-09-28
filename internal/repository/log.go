package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/utils"
	"gorm.io/gorm"
)

type LogRepository struct {
	db *gorm.DB
}

// NewLogRepository returns pointer to repo along with the db
func NewLogRepository(db *gorm.DB) *LogRepository {
	return &LogRepository{db}
}

func (r *LogRepository) CreateLog(log *models.Log) (*models.Log, error) {
	if err := r.db.Create(log).Error; err != nil {
		return nil, err
	}

	return log, nil
}

func (r *LogRepository) ReadLog(uid string) (*models.Log, error) {
	log := &models.Log{}

	if err := r.db.Where("unique_id = ?", uid).First(log).Error; err != nil {
		return nil, err
	}

	return log, nil
}

func (r *LogRepository) ListLogs(opts ...utils.QueryOption) ([]*models.Log, error) {
	var logs []*models.Log

	if err := r.db.Scopes(utils.Paginate(opts)).Find(&logs).Error; err != nil {
		return nil, err
	}

	return logs, nil
}

func (r *LogRepository) DeleteLog(uid string) error {
	log := &models.Log{}

	if err := r.db.Where("unique_id = ?", uid).First(log).Error; err != nil {
		return err
	}

	if err := r.db.Delete(log).Error; err != nil {
		return err
	}

	return nil
}
