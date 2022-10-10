package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"gorm.io/gorm"
)

type AlertRepository struct {
	db *gorm.DB
}

// NewAlertRepository returns pointer to repo along with the db
func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db}
}

func (r *AlertRepository) CreateAlert(alert *models.Alert) (*models.Alert, error) {
	if err := r.db.Create(alert).Error; err != nil {
		return nil, err
	}

	return alert, nil
}

func (r *AlertRepository) ListAlertsByPodName(triggeringPodName string) ([]*models.Alert, error) {
	alerts := make([]*models.Alert, 0)

	if err := r.db.Where("triggering_pod_name = ?", triggeringPodName).Find(&alerts).Error; err != nil {
		return nil, err
	}

	return alerts, nil
}
