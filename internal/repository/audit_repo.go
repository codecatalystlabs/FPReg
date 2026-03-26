package repository

import (
	"strings"

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
	UserID        *uuid.UUID
	FacilityID    *uuid.UUID
	DistrictScope string // non-empty: only logs tied to facilities in this district or users assigned to that district
	Action        string
	Entity        string
	DateFrom      string
	DateTo        string
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
	if strings.TrimSpace(f.DistrictScope) != "" {
		d := strings.TrimSpace(f.DistrictScope)
		q = q.Where(`(
			facility_id IN (SELECT id FROM facilities WHERE LOWER(TRIM(district)) = LOWER(?))
			OR user_id IN (
				SELECT id FROM users WHERE facility_id IN (
					SELECT id FROM facilities WHERE LOWER(TRIM(district)) = LOWER(?)
				)
				OR (role = ? AND LOWER(TRIM(COALESCE(district, ''))) = LOWER(?))
			)
		)`, d, d, models.RoleDistrictBiostatistician, d)
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
