package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/utils"
	"gorm.io/gorm"
)

type IncidentRepository struct {
	db *gorm.DB
}

// NewSessionRepository returns pointer to repo along with the db
func NewIncidentRepository(db *gorm.DB) *IncidentRepository {
	return &IncidentRepository{db}
}

func (r *IncidentRepository) CreateIncident(incident *models.Incident) (*models.Incident, error) {
	if err := r.db.Create(incident).Error; err != nil {
		return nil, err
	}

	return incident, nil
}

func (r *IncidentRepository) ReadIncident(uid string) (*models.Incident, error) {
	incident := &models.Incident{}

	if err := r.db.Where("unique_id = ?", uid).First(incident).Error; err != nil {
		return nil, err
	}

	return incident, nil
}

func (r *IncidentRepository) ListIncidents(opts ...utils.QueryOption) ([]*models.Incident, error) {
	var incidents []*models.Incident

	if err := r.db.Scopes(utils.Paginate(opts)).Find(&incidents).Error; err != nil {
		return nil, err
	}

	return incidents, nil
}

func (r *IncidentRepository) DeleteIncident(uid string) error {
	incident := &models.Incident{}

	if err := r.db.Where("unique_id = ?", uid).First(&incident).Error; err != nil {
		return err
	}

	if err := r.db.Delete(incident).Error; err != nil {
		return err
	}

	return nil
}
