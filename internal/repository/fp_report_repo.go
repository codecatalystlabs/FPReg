package repository

import (
	"time"

	"fpreg/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FPReportRepository struct {
	db *gorm.DB
}

func NewFPReportRepository(db *gorm.DB) *FPReportRepository {
	return &FPReportRepository{db: db}
}

// ListForRange returns all FPRegistration rows between [from, to) optionally filtered by facility IDs.
func (r *FPReportRepository) ListForRange(from, to time.Time, facilityIDs []uuid.UUID) ([]models.FPRegistration, error) {
	var regs []models.FPRegistration

	q := r.db.Model(&models.FPRegistration{}).
		Where("visit_date >= ? AND visit_date < ?", from.Format("2006-01-02"), to.Format("2006-01-02"))

	if len(facilityIDs) > 0 {
		q = q.Where("facility_id IN ?", facilityIDs)
	}

	err := q.Find(&regs).Error
	return regs, err
}

