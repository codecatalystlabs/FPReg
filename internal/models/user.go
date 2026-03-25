package models

import "github.com/google/uuid"

type Role string

const (
	RoleSuperAdmin              Role = "superadmin"
	RoleFacilityAdmin           Role = "facility_admin"
	RoleFacilityUser            Role = "facility_user"
	RoleReviewer                Role = "reviewer"
	RoleDistrictBiostatistician Role = "district_biostatistician"
)

type User struct {
	BaseModel
	Email      string    `gorm:"size:200;uniqueIndex;not null" json:"email"`
	Password   string    `gorm:"size:255;not null" json:"-"`
	FullName   string    `gorm:"size:200;not null" json:"full_name"`
	Role       Role      `gorm:"size:40;not null;default:'facility_user'" json:"role"`
	FacilityID *uuid.UUID `gorm:"type:uuid;index" json:"facility_id,omitempty"`
	Facility   *Facility  `gorm:"foreignKey:FacilityID" json:"facility,omitempty"`
	// District is set for district_biostatistician (no facility). Must match facilities.district for scoped access.
	District string `gorm:"size:100" json:"district,omitempty"`
	IsActive bool   `gorm:"default:true" json:"is_active"`
}
