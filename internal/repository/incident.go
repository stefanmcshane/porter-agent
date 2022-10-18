package repository

import (
	"github.com/porter-dev/porter-agent/internal/models"
	"github.com/porter-dev/porter-agent/internal/utils"
	"gorm.io/gorm"
)

type IncidentRepository struct {
	db *gorm.DB
}

// NewIncidentRepository returns pointer to repo along with the db
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

	if err := r.db.Debug().Preload("Events", func(db *gorm.DB) *gorm.DB {
		return db.Order("incident_events.last_seen DESC")
	}).Where("unique_id = ?", uid).First(incident).Error; err != nil {
		return nil, err
	}

	return incident, nil
}

func (r *IncidentRepository) UpdateIncident(incident *models.Incident) (*models.Incident, error) {
	// TODO: store incident events and make sure they're de-duped
	if err := r.db.Save(incident).Error; err != nil {
		return nil, err
	}

	return incident, nil
}

func (r *IncidentRepository) ListIncidents(
	filter *utils.ListIncidentsFilter,
	opts ...utils.QueryOption,
) ([]*models.Incident, *utils.PaginatedResult, error) {
	var incidents []*models.Incident

	db := r.db.Model(&models.Incident{})

	if filter.Status != nil {
		db = db.Where("incident_status = ?", *filter.Status)
	}

	if filter.ReleaseName != nil {
		db = db.Where("release_name = ?", *filter.ReleaseName)
	}

	if filter.ReleaseNamespace != nil {
		db = db.Where("release_namespace = ?", *filter.ReleaseNamespace)
	}

	paginatedResult := &utils.PaginatedResult{}

	db = db.Scopes(utils.Paginate(opts, db, paginatedResult))

	if err := db.Preload("Events", func(db *gorm.DB) *gorm.DB {
		return db.Order("incident_events.last_seen DESC")
	}).Find(&incidents).Error; err != nil {
		return nil, nil, err
	}

	return incidents, paginatedResult, nil
}

func (r *IncidentRepository) DeleteIncident(uid string) error {
	incident := &models.Incident{}

	if err := r.db.Where("unique_id = ?", uid).First(incident).Error; err != nil {
		return err
	}

	if err := r.db.Delete(incident).Error; err != nil {
		return err
	}

	return nil
}
