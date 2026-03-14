package models

import (
	"time"

	"github.com/google/uuid"
)

type AuditAction string

const (
	AuditLogin         AuditAction = "LOGIN"
	AuditLoginFailed   AuditAction = "LOGIN_FAILED"
	AuditLogout        AuditAction = "LOGOUT"
	AuditTokenRefresh  AuditAction = "TOKEN_REFRESH"
	AuditCreate        AuditAction = "CREATE"
	AuditUpdate        AuditAction = "UPDATE"
	AuditDelete        AuditAction = "DELETE"
	AuditView          AuditAction = "VIEW"
	AuditExport        AuditAction = "EXPORT"
	AuditAdminAction   AuditAction = "ADMIN_ACTION"
)

type AuditLog struct {
	ID         uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID     *uuid.UUID  `gorm:"type:uuid;index" json:"user_id,omitempty"`
	FacilityID *uuid.UUID  `gorm:"type:uuid;index" json:"facility_id,omitempty"`
	Action     AuditAction `gorm:"size:30;not null;index" json:"action"`
	Entity     string      `gorm:"size:50;index" json:"entity"`
	EntityID   string      `gorm:"size:50" json:"entity_id,omitempty"`
	OldValues  string      `gorm:"type:text" json:"old_values,omitempty"`
	NewValues  string      `gorm:"type:text" json:"new_values,omitempty"`
	IPAddress  string      `gorm:"size:45" json:"ip_address"`
	UserAgent  string      `gorm:"size:500" json:"user_agent"`
	Detail     string      `gorm:"size:500" json:"detail,omitempty"`
	CreatedAt  time.Time   `gorm:"index" json:"created_at"`
}
