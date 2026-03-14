package repository

import (
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

func (r *UserRepository) List(page, perPage int, facilityID *uuid.UUID) ([]models.User, int64, error) {
	var users []models.User
	var total int64
	q := r.db.Model(&models.User{})
	if facilityID != nil {
		q = q.Where("facility_id = ?", *facilityID)
	}
	q.Count(&total)
	err := q.Preload("Facility").
		Offset((page - 1) * perPage).Limit(perPage).
		Order("created_at DESC").
		Find(&users).Error
	return users, total, err
}
