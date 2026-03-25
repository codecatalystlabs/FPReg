package repository

import (
	"errors"
	"strings"

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
	err := r.db.Where("local_indicator_key = ?", item.LocalIndicatorKey).
		Order("updated_at DESC NULLS LAST, created_at DESC NULLS LAST, id ASC").
		First(&existing).Error
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
	// Usable for sync when both DHIS2 UIDs are set (active flag optional once configured).
	err := r.db.Where("local_indicator_key = ?", key).
		Where("TRIM(COALESCE(dhis2_data_element_uid, '')) <> ''").
		Where("TRIM(COALESCE(dhis2_cat_option_combo_uid, '')) <> ''").
		Order("updated_at DESC NULLS LAST, created_at DESC NULLS LAST, id ASC").
		First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// SeedFPIndicatorMappingStubs inserts template rows for keys that do not exist yet (inactive, empty DHIS2 UIDs).
func (r *DHIS2Repository) SeedFPIndicatorMappingStubs(stubs []models.DHIS2MappingItem) (inserted int, err error) {
	for i := range stubs {
		s := &stubs[i]
		var n int64
		if err := r.db.Model(&models.DHIS2MappingItem{}).Where("local_indicator_key = ?", s.LocalIndicatorKey).Count(&n).Error; err != nil {
			return inserted, err
		}
		if n > 0 {
			continue
		}
		if err := r.db.Create(s).Error; err != nil {
			return inserted, err
		}
		inserted++
	}
	return inserted, nil
}

// SyncOrgUnitMappingsFromFacilities upserts one row per facility that has a non-empty uid
// (DHIS2 org unit). Matches: local_facility_id = facilities.id, local_facility_name = name,
// dhis2 org unit = facilities.uid.
func (r *DHIS2Repository) SyncOrgUnitMappingsFromFacilities() (int, error) {
	var facs []models.Facility
	if err := r.db.Order("name").Find(&facs).Error; err != nil {
		return 0, err
	}
	n := 0
	for _, f := range facs {
		uid := strings.TrimSpace(f.UID)
		if uid == "" {
			continue
		}
		var existing models.OrgUnitMapping
		err := r.db.Where("local_facility_id = ?", f.ID).First(&existing).Error
		m := models.OrgUnitMapping{
			LocalFacilityID:   f.ID,
			LocalFacilityName: f.Name,
			DHIS2OrgUnitUID:   uid,
			Active:            true,
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := r.db.Create(&m).Error; err != nil {
				return n, err
			}
		} else if err != nil {
			return n, err
		} else {
			m.ID = existing.ID
			if err := r.db.Save(&m).Error; err != nil {
				return n, err
			}
		}
		n++
	}
	return n, nil
}

// FindOrgUnitMappingByFacilityID returns the active mapping for a local facility, if any.
func (r *DHIS2Repository) FindOrgUnitMappingByFacilityID(facilityID uuid.UUID) (*models.OrgUnitMapping, error) {
	var m models.OrgUnitMapping
	err := r.db.Where("local_facility_id = ? AND active = ?", facilityID, true).First(&m).Error
	if err != nil {
		return nil, err
	}
	return &m, nil
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
