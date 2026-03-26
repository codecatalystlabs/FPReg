package models

import (
	"time"

	"github.com/google/uuid"
)

type DHIS2MappingItem struct {
	ID                     uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LocalIndicatorKey      string    `gorm:"size:100;index;not null" json:"local_indicator_key"`
	MethodCode             string    `gorm:"size:20;not null" json:"method_code"`
	MethodName             string    `gorm:"size:200;not null" json:"method_name"`
	Subgroup               string    `gorm:"size:20" json:"subgroup"`
	VisitType              string    `gorm:"size:20;not null" json:"visit_type"`    // NEW or REVISIT
	AgeGroup               string    `gorm:"size:20;not null" json:"age_group"`     // BELOW_15, 15_19, 20_24, …
	DHIS2DataElementUID    string    `gorm:"size:64" json:"dhis2_data_element_uid"` // nullable – required for active rows
	DHIS2CatOptionComboUID string    `gorm:"size:64" json:"dhis2_cat_option_combo_uid"`
	Active                 bool      `gorm:"default:true" json:"active"`
	Notes                  string    `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt              time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type DHIS2SyncSetting struct {
	ID                   uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	DHIS2BaseURL         string    `gorm:"size:255;not null" json:"dhis2_base_url"`
	AuthType             string    `gorm:"size:20;not null;default:basic" json:"auth_type"` // basic|token
	Username             string    `gorm:"size:100" json:"username,omitempty"`
	PasswordEncrypted    string    `gorm:"size:255" json:"password_encrypted,omitempty"`
	TokenEncrypted       string    `gorm:"size:255" json:"token_encrypted,omitempty"`
	AttributeOptionCombo string    `gorm:"size:64;not null" json:"attributeoptioncombo_uid"`
	DatasetUID           string    `gorm:"size:64" json:"dataset_uid,omitempty"`
	Active               bool      `gorm:"default:true" json:"active"`
	CreatedAt            time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type OrgUnitMapping struct {
	ID                uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	LocalFacilityID   uuid.UUID `gorm:"type:uuid;not null;index" json:"local_facility_id"`
	LocalFacilityName string    `gorm:"size:200;not null" json:"local_facility_name"`
	DHIS2OrgUnitUID   string    `gorm:"size:64;not null" json:"dhis2_orgunit_uid"`
	Active            bool      `gorm:"default:true" json:"active"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type ReportSyncLog struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Period             string     `gorm:"size:6;not null;index" json:"period"` // YYYYMM
	OrgUnitUID         string     `gorm:"size:64;not null;index" json:"orgunit_uid"`
	LocalFacilityID    *uuid.UUID `gorm:"type:uuid" json:"local_facility_id,omitempty"`
	TriggerType        string     `gorm:"size:20;not null" json:"trigger_type"` // automatic|manual
	SyncScope          string     `gorm:"size:50;not null" json:"sync_scope"`   // e.g. monthly_fp_methods
	PayloadJSON        string     `gorm:"type:text;not null" json:"payload_json"`
	ResponseStatusCode int        `json:"response_status_code"`
	ResponseBody       string     `gorm:"type:text" json:"response_body,omitempty"`
	Success            bool       `gorm:"default:false" json:"success"`
	ForceResync        bool       `gorm:"default:false" json:"force_resync"`
	InitiatedBy        string     `gorm:"size:100" json:"initiated_by,omitempty"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
}

type ReportCellSyncStatus struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Period            string     `gorm:"size:6;not null;index" json:"period"`
	LocalIndicatorKey string     `gorm:"size:100;not null;index" json:"local_indicator_key"`
	OrgUnitUID        string     `gorm:"size:64;not null;index" json:"orgunit_uid"`
	Value             int        `gorm:"not null" json:"value"`
	LastSyncedAt      *time.Time `json:"last_synced_at,omitempty"`
	LastSyncLogID     *uuid.UUID `gorm:"type:uuid" json:"last_sync_log_id,omitempty"`
	SyncStatus        string     `gorm:"size:20;not null;default:'pending'" json:"sync_status"` // pending|synced|failed
	Checksum          string     `gorm:"size:64" json:"checksum,omitempty"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

type AggregationExclusionLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	RawRecordID *uuid.UUID `gorm:"type:uuid" json:"raw_record_id,omitempty"`
	Reason      string     `gorm:"size:200;not null" json:"reason"`
	Period      string     `gorm:"size:6;not null" json:"period"`
	FacilityID  *uuid.UUID `gorm:"type:uuid" json:"facility_id,omitempty"`
	Details     string     `gorm:"type:text" json:"details,omitempty"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"created_at"`
}
