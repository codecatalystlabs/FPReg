package database

import (
	"log"

	"fpreg/internal/models"

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
}
