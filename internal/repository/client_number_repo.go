package repository

import (
	"fmt"
	"time"

	"fpreg/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ClientNumberRepository struct {
	db *gorm.DB
}

func NewClientNumberRepository(db *gorm.DB) *ClientNumberRepository {
	return &ClientNumberRepository{db: db}
}

// NextClientNumber atomically generates the next client number for a facility on a given date.
// Format: {prefix}{YYMMDD}{NNN} e.g. DHC260314001
// Uses SELECT FOR UPDATE to prevent duplicates under concurrent requests.
func (r *ClientNumberRepository) NextClientNumber(tx *gorm.DB, facilityID uuid.UUID, prefix string, date time.Time) (string, error) {
	dateOnly := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	datePart := date.Format("060102")

	var seq models.ClientNumberSeq
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("facility_id = ? AND seq_date = ?", facilityID, dateOnly).
		First(&seq).Error

	if err == gorm.ErrRecordNotFound {
		seq = models.ClientNumberSeq{
			FacilityID: facilityID,
			SeqDate:    dateOnly,
			LastSeq:    1,
		}
		if err := tx.Create(&seq).Error; err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	} else {
		seq.LastSeq++
		if err := tx.Save(&seq).Error; err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s%s%03d", prefix, datePart, seq.LastSeq), nil
}
