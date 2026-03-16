package repository

import (
	"fpreg/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FacilityRepository struct {
	db *gorm.DB
}

func NewFacilityRepository(db *gorm.DB) *FacilityRepository {
	return &FacilityRepository{db: db}
}

func (r *FacilityRepository) Create(f *models.Facility) error {
	return r.db.Create(f).Error
}

func (r *FacilityRepository) FindByID(id uuid.UUID) (*models.Facility, error) {
	var f models.Facility
	err := r.db.First(&f, "id = ?", id).Error
	return &f, err
}

func (r *FacilityRepository) FindByCode(code string) (*models.Facility, error) {
	var f models.Facility
	err := r.db.First(&f, "code = ?", code).Error
	return &f, err
}

func (r *FacilityRepository) FindByUID(uid string) (*models.Facility, error) {
	if uid == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var f models.Facility
	err := r.db.First(&f, "uid = ?", uid).Error
	return &f, err
}

func (r *FacilityRepository) UpsertByUID(f *models.Facility) error {
	existing, err := r.FindByUID(f.UID)
	if err == nil {
		f.ID = existing.ID
		return r.db.Save(f).Error
	}
	return r.db.Create(f).Error
}

func (r *FacilityRepository) Update(f *models.Facility) error {
	return r.db.Save(f).Error
}

func (r *FacilityRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.Facility{}, "id = ?", id).Error
}

func (r *FacilityRepository) List(page, perPage int) ([]models.Facility, int64, error) {
	var facilities []models.Facility
	var total int64
	r.db.Model(&models.Facility{}).Count(&total)
	err := r.db.Offset((page - 1) * perPage).Limit(perPage).
		Order("name ASC").
		Find(&facilities).Error
	return facilities, total, err
}
