package repository

import (
	"strings"

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

func (r *FacilityRepository) FindByIDs(ids []uuid.UUID) ([]models.Facility, error) {
	var items []models.Facility
	if len(ids) == 0 {
		return items, nil
	}
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
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

func (r *FacilityRepository) List(page, perPage int, search string) ([]models.Facility, int64, error) {
	var facilities []models.Facility
	var total int64

	q := r.db.Model(&models.Facility{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where(
			r.db.Where("name ILIKE ?", like).
				Or("code ILIKE ?", like).
				Or("district ILIKE ?", like).
				Or("subcounty ILIKE ?", like),
		)
	}

	q.Count(&total)

	err := q.Order("name ASC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&facilities).Error
	return facilities, total, err
}

// ListByDistrict lists facilities whose district matches (case-insensitive, trimmed).
func (r *FacilityRepository) ListByDistrict(page, perPage int, district, search string) ([]models.Facility, int64, error) {
	var facilities []models.Facility
	var total int64
	d := strings.TrimSpace(district)
	q := r.db.Model(&models.Facility{}).Where("LOWER(TRIM(district)) = LOWER(?)", d)
	if search != "" {
		like := "%" + search + "%"
		q = q.Where(
			r.db.Where("name ILIKE ?", like).
				Or("code ILIKE ?", like).
				Or("district ILIKE ?", like).
				Or("subcounty ILIKE ?", like),
		)
	}
	q.Count(&total)
	err := q.Order("name ASC").
		Offset((page - 1) * perPage).
		Limit(perPage).
		Find(&facilities).Error
	return facilities, total, err
}

// FindIDsByDistrict returns all facility IDs in the given district.
func (r *FacilityRepository) FindIDsByDistrict(district string) ([]uuid.UUID, error) {
	d := strings.TrimSpace(district)
	var ids []uuid.UUID
	err := r.db.Model(&models.Facility{}).
		Where("LOWER(TRIM(district)) = LOWER(?)", d).
		Pluck("id", &ids).Error
	return ids, err
}

// ListDistinctDistricts returns unique district names from facilities (trimmed),
// one canonical spelling per case-insensitive group, sorted A–Z.
func (r *FacilityRepository) ListDistinctDistricts() ([]string, error) {
	var rows []struct {
		D string `gorm:"column:d"`
	}
	err := r.db.Raw(`
		SELECT MIN(TRIM(district)) AS d
		FROM facilities
		WHERE district IS NOT NULL AND TRIM(district) <> ''
		GROUP BY LOWER(TRIM(district))
		ORDER BY MIN(TRIM(district)) ASC
	`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.D != "" {
			out = append(out, row.D)
		}
	}
	return out, nil
}

// FacilityBelongsToDistrict returns true if the facility's district matches (case-insensitive).
func (r *FacilityRepository) FacilityBelongsToDistrict(facilityID uuid.UUID, district string) (bool, error) {
	d := strings.TrimSpace(district)
	if d == "" {
		return false, nil
	}
	var f models.Facility
	err := r.db.Select("id", "district").First(&f, "id = ?", facilityID).Error
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(f.District), d), nil
}
