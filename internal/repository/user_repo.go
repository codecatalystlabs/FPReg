package repository

import (
	"strings"

	"fpreg/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) FindByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Facility").First(&user, "id = ?", id).Error
	return &user, err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Preload("Facility").First(&user, "email = ?", email).Error
	return &user, err
}

func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

// List returns users. If facilityID is set, filter by that facility only.
// If districtScope is non-empty, return users whose facility is in that district, or district_biostatisticians assigned to that district.
func (r *UserRepository) List(page, perPage int, facilityID *uuid.UUID, districtScope string) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	q := r.db.Model(&models.User{})
	if facilityID != nil {
		q = q.Where("facility_id = ?", *facilityID)
	} else if strings.TrimSpace(districtScope) != "" {
		d := strings.TrimSpace(districtScope)
		sub := r.db.Model(&models.Facility{}).Select("id").Where("LOWER(TRIM(district)) = LOWER(?)", d)
		q = q.Where(
			r.db.Where("facility_id IN (?)", sub).
				Or("(role = ? AND LOWER(TRIM(district)) = LOWER(?))", models.RoleDistrictBiostatistician, d),
		)
	}
	q.Count(&total)
	err := q.Preload("Facility").
		Offset((page - 1) * perPage).Limit(perPage).
		Order("created_at DESC").
		Find(&users).Error
	return users, total, err
}
