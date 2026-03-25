package repository

import (
	"strings"

	"fpreg/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RegistrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

func (r *RegistrationRepository) Create(tx *gorm.DB, reg *models.FPRegistration) error {
	if tx == nil {
		tx = r.db
	}
	return tx.Create(reg).Error
}

func (r *RegistrationRepository) FindByID(id uuid.UUID) (*models.FPRegistration, error) {
	var reg models.FPRegistration
	err := r.db.Preload("Facility").Preload("Creator").
		First(&reg, "id = ?", id).Error
	return &reg, err
}

func (r *RegistrationRepository) Update(reg *models.FPRegistration) error {
	return r.db.Save(reg).Error
}

func (r *RegistrationRepository) SoftDelete(id uuid.UUID) error {
	return r.db.Delete(&models.FPRegistration{}, "id = ?", id).Error
}

type RegistrationFilter struct {
	FacilityID *uuid.UUID
	// District filters registrations to facilities in this district (trimmed, case-insensitive match).
	District  string
	VisitDate string
	Search    string
	Sex       string
	IsNewUser *bool
	DateFrom  string
	DateTo    string
}

func (r *RegistrationRepository) List(page, perPage int, f RegistrationFilter) ([]models.FPRegistration, int64, error) {
	var items []models.FPRegistration
	var total int64

	q := r.db.Model(&models.FPRegistration{})

	if f.FacilityID != nil {
		q = q.Where("facility_id = ?", *f.FacilityID)
	}
	if strings.TrimSpace(f.District) != "" {
		d := strings.TrimSpace(f.District)
		q = q.Where("facility_id IN (?)",
			r.db.Model(&models.Facility{}).Select("id").Where("LOWER(TRIM(district)) = LOWER(?)", d))
	}
	if f.VisitDate != "" {
		q = q.Where("visit_date = ?", f.VisitDate)
	}
	if f.Sex != "" {
		q = q.Where("sex = ?", f.Sex)
	}
	if f.IsNewUser != nil {
		q = q.Where("is_new_user = ?", *f.IsNewUser)
	}
	if f.DateFrom != "" {
		q = q.Where("visit_date >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		q = q.Where("visit_date <= ?", f.DateTo)
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		q = q.Where("surname ILIKE ? OR given_name ILIKE ? OR client_number ILIKE ? OR nin ILIKE ?",
			like, like, like, like)
	}

	q.Count(&total)

	err := q.Preload("Facility").Preload("Creator").
		Offset((page - 1) * perPage).Limit(perPage).
		Order("visit_date DESC, serial_number DESC").
		Find(&items).Error

	return items, total, err
}

func (r *RegistrationRepository) NextSerialNumber(tx *gorm.DB, facilityID uuid.UUID, visitDate string) (int, error) {
	if tx == nil {
		tx = r.db
	}
	var maxSerial *int
	err := tx.Model(&models.FPRegistration{}).
		Where("facility_id = ? AND visit_date = ?", facilityID, visitDate).
		Select("MAX(serial_number)").
		Scan(&maxSerial).Error
	if err != nil {
		return 0, err
	}
	if maxSerial == nil {
		return 1, nil
	}
	return *maxSerial + 1, nil
}

func (r *RegistrationRepository) DB() *gorm.DB {
	return r.db
}
