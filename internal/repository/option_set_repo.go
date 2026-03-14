package repository

import (
	"fpreg/internal/models"

	"gorm.io/gorm"
)

type OptionSetRepository struct {
	db *gorm.DB
}

func NewOptionSetRepository(db *gorm.DB) *OptionSetRepository {
	return &OptionSetRepository{db: db}
}

func (r *OptionSetRepository) FindByCategory(category string) ([]models.OptionSet, error) {
	var items []models.OptionSet
	err := r.db.Where("category = ? AND is_active = true", category).
		Order("sort_order ASC").Find(&items).Error
	return items, err
}

func (r *OptionSetRepository) FindAll() ([]models.OptionSet, error) {
	var items []models.OptionSet
	err := r.db.Where("is_active = true").
		Order("category ASC, sort_order ASC").Find(&items).Error
	return items, err
}

func (r *OptionSetRepository) FindAllGrouped() (map[string][]models.OptionSet, error) {
	items, err := r.FindAll()
	if err != nil {
		return nil, err
	}
	grouped := make(map[string][]models.OptionSet)
	for _, item := range items {
		grouped[item.Category] = append(grouped[item.Category], item)
	}
	return grouped, nil
}

func (r *OptionSetRepository) Categories() ([]string, error) {
	var categories []string
	err := r.db.Model(&models.OptionSet{}).
		Where("is_active = true").
		Distinct("category").
		Pluck("category", &categories).Error
	return categories, err
}
