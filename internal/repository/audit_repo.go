package repository

import (
	"fpreg/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *models.AuditLog) error {
	return r.db.Create(log).Error
}

type AuditFilter struct {
	UserID     *uuid.UUID
	FacilityID *uuid.UUID
	Action     string
	Entity     string
	DateFrom   string
	DateTo     string
}

func (r *AuditRepository) List(page, perPage int, f AuditFilter) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	q := r.db.Model(&models.AuditLog{})

	if f.UserID != nil {
		q = q.Where("user_id = ?", *f.UserID)
	}
	if f.FacilityID != nil {
		q = q.Where("facility_id = ?", *f.FacilityID)
	}
	if f.Action != "" {
		q = q.Where("action = ?", f.Action)
	}
	if f.Entity != "" {
		q = q.Where("entity = ?", f.Entity)
	}
	if f.DateFrom != "" {
		q = q.Where("created_at >= ?", f.DateFrom)
	}
	if f.DateTo != "" {
		q = q.Where("created_at <= ?", f.DateTo+" 23:59:59")
	}

	q.Count(&total)

	err := q.Offset((page - 1) * perPage).Limit(perPage).
		Order("created_at DESC").
		Find(&logs).Error

	return logs, total, err
}
