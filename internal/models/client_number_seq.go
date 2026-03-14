package models

import (
	"time"

	"github.com/google/uuid"
)

type ClientNumberSeq struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	FacilityID uuid.UUID `gorm:"type:uuid;not null" json:"facility_id"`
	SeqDate    time.Time `gorm:"type:date;not null" json:"seq_date"`
	LastSeq    int       `gorm:"not null;default:0" json:"last_seq"`
}
