package repository

import (
	"errors"

	"fpreg/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DHIS2Repository struct {
	db *gorm.DB
}

func NewDHIS2Repository(db *gorm.DB) *DHIS2Repository {
	return &DHIS2Repository{db: db}
}

// Settings
func (r *DHIS2Repository) GetActiveSettings() (*models.DHIS2SyncSetting, error) {
	var s models.DHIS2SyncSetting
	err := r.db.Where("active = ?", true).Order("updated_at DESC").First(&s).Error
	return &s, err
}

func (r *DHIS2Repository) UpsertSettings(s *models.DHIS2SyncSetting) error {
	// Keep only one active at a time
	if s.Active {
		r.db.Model(&models.DHIS2SyncSetting{}).Where("id <> ?", s.ID).Update("active", false)
	}
	if s.ID == uuid.Nil {
		return r.db.Create(s).Error
	}
	return r.db.Save(s).Error
}

// Mapping Items
func (r *DHIS2Repository) UpsertMappingItem(item *models.DHIS2MappingItem) error {
	if item.LocalIndicatorKey == "" {
		return errors.New("local_indicator_key is required")
	}
	var existing models.DHIS2MappingItem
	err := r.db.First(&existing, "local_indicator_key = ?", item.LocalIndicatorKey).Error
	if err == nil {
		item.ID = existing.ID
		return r.db.Save(item).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(item).Error
	}
	return err
}

func (r *DHIS2Repository) FindMappingByLocalKey(key string) (*models.DHIS2MappingItem, error) {
	var item models.DHIS2MappingItem
	err := r.db.First(&item, "local_indicator_key = ? AND active = true", key).Error
	return &item, err
}

func (r *DHIS2Repository) ListMappings(page, perPage int, search string) ([]models.DHIS2MappingItem, int64, error) {
	var items []models.DHIS2MappingItem
	var total int64
	q := r.db.Model(&models.DHIS2MappingItem{})
	if search != "" {
		like := "%" + search + "%"
		q = q.Where("local_indicator_key ILIKE ? OR method_code ILIKE ? OR method_name ILIKE ?", like, like, like)
	}
	q.Count(&total)
	err := q.Order("method_code ASC, visit_type ASC, age_group ASC").
		Offset((page - 1) * perPage).Limit(perPage).
		Find(&items).Error
	return items, total, err
}

// Sync Logs + Status
func (r *DHIS2Repository) CreateSyncLog(log *models.ReportSyncLog) error {
	return r.db.Create(log).Error
}

func (r *DHIS2Repository) ListSyncLogs(page, perPage int, period, orgUnit, trigger string, success *bool) ([]models.ReportSyncLog, int64, error) {
	var items []models.ReportSyncLog
	var total int64
	q := r.db.Model(&models.ReportSyncLog{})
	if period != "" {
		q = q.Where("period = ?", period)
	}
	if orgUnit != "" {
		q = q.Where("org_unit_uid = ?", orgUnit)
	}
	if trigger != "" {
		q = q.Where("trigger_type = ?", trigger)
	}
	if success != nil {
		q = q.Where("success = ?", *success)
	}
	q.Count(&total)
	err := q.Order("created_at DESC").
		Offset((page - 1) * perPage).Limit(perPage).
		Find(&items).Error
	return items, total, err
}

func (r *DHIS2Repository) GetCellStatus(period, orgUnit, localKey string) (*models.ReportCellSyncStatus, error) {
	var s models.ReportCellSyncStatus
	err := r.db.First(&s, "period = ? AND org_unit_uid = ? AND local_indicator_key = ?", period, orgUnit, localKey).Error
	return &s, err
}

func (r *DHIS2Repository) UpsertCellStatus(s *models.ReportCellSyncStatus) error {
	var existing models.ReportCellSyncStatus
	err := r.db.First(&existing, "period = ? AND org_unit_uid = ? AND local_indicator_key = ?", s.Period, s.OrgUnitUID, s.LocalIndicatorKey).Error
	if err == nil {
		s.ID = existing.ID
		return r.db.Save(s).Error
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.Create(s).Error
	}
	return err
}

func (r *DHIS2Repository) CreateExclusion(log *models.AggregationExclusionLog) error {
	return r.db.Create(log).Error
}

