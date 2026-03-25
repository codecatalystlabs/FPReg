package database

import (
	"log"

	"fpreg/internal/models"
	"fpreg/internal/repository"
	"fpreg/internal/service"

	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`)

	err := db.AutoMigrate(
		&models.Facility{},
		&models.User{},
		&models.RefreshToken{},
		&models.OptionSet{},
		&models.ClientNumberSeq{},
		&models.FPRegistration{},
		&models.AuditLog{},
		&models.DHIS2MappingItem{},
		&models.DHIS2SyncSetting{},
		&models.OrgUnitMapping{},
		&models.ReportSyncLog{},
		&models.ReportCellSyncStatus{},
		&models.AggregationExclusionLog{},
	)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	createIndexes(db)

	dhisRepo := repository.NewDHIS2Repository(db)
	if n, err := dhisRepo.SyncOrgUnitMappingsFromFacilities(); err != nil {
		log.Printf("Warning: org_unit_mappings sync from facilities: %v", err)
	} else if n > 0 {
		log.Printf("org_unit_mappings: upserted %d row(s) from facilities (uid non-empty)", n)
	}

	if n, err := dhisRepo.SeedFPIndicatorMappingStubs(service.StubDHIS2MappingItems()); err != nil {
		log.Printf("Warning: dhis2_mapping_item stubs: %v", err)
	} else if n > 0 {
		log.Printf("dhis2_mapping_item: inserted %d stub row(s) — set DHIS2 data element + category option combo UIDs to enable sync", n)
	}

	if items, err := service.OfficialDHIS2MappingItemsFromEmbeddedCSV(); err != nil {
		log.Printf("Warning: dhis2 official CSV mapping: %v", err)
	} else {
		ok := 0
		for i := range items {
			if err := dhisRepo.UpsertMappingItem(&items[i]); err != nil {
				log.Printf("Warning: dhis2_mapping_item upsert %q: %v", items[i].LocalIndicatorKey, err)
			} else {
				ok++
			}
		}
		if ok > 0 {
			log.Printf("dhis2_mapping_item: upserted %d row(s) from embedded dhis2_mapping_official.csv", ok)
		}
	}

	log.Println("Database migration completed")
}

func createIndexes(db *gorm.DB) {
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_client_number_seq_unique
		ON client_number_seqs (facility_id, seq_date)`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_option_set_cat_code
		ON option_sets (category, code) WHERE deleted_at IS NULL`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_fp_reg_visit_date
		ON fp_registrations (facility_id, visit_date)`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_fp_reg_client_number
		ON fp_registrations (client_number) WHERE client_number IS NOT NULL`)
	db.Exec(`CREATE INDEX IF NOT EXISTS idx_audit_created
		ON audit_logs (created_at DESC)`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_org_unit_mappings_local_facility
		ON org_unit_mappings (local_facility_id)`)
}
