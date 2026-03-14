package database

import (
	"log"

	"fpreg/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func Seed(db *gorm.DB, adminEmail, adminPassword, adminName string) {
	seedOptionSets(db)
	seedDefaultFacility(db)
	seedAdminUser(db, adminEmail, adminPassword, adminName)
	log.Println("Seed data loaded")
}

func seedDefaultFacility(db *gorm.DB) {
	var count int64
	db.Model(&models.Facility{}).Count(&count)
	if count > 0 {
		return
	}
	facility := models.Facility{
		Name:             "Demo Health Centre IV",
		Code:             "DEMO-HC4",
		Level:            "HC IV",
		Subcounty:        "Central Division",
		HSD:              "Central HSD",
		District:         "Kampala",
		ClientCodePrefix: "DHC",
	}
	db.Create(&facility)
}

func seedAdminUser(db *gorm.DB, email, password, name string) {
	var count int64
	db.Model(&models.User{}).Where("role = ?", models.RoleSuperAdmin).Count(&count)
	if count > 0 {
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		Email:    email,
		Password: string(hash),
		FullName: name,
		Role:     models.RoleSuperAdmin,
		IsActive: true,
	}
	db.Create(&user)
}

func seedOptionSets(db *gorm.DB) {
	var count int64
	db.Model(&models.OptionSet{}).Count(&count)
	if count > 0 {
		return
	}

	sets := []models.OptionSet{
		// HTS Codes
		{Category: "hts_code", Code: "NEG", Label: "Negative", Description: "Client counselled, tested, and given HIV Negative results", SortOrder: 1},
		{Category: "hts_code", Code: "POS", Label: "Positive", Description: "Client counselled, tested, and given HIV Positive results", SortOrder: 2},
		{Category: "hts_code", Code: "INC", Label: "Inconclusive", Description: "Client counselled, tested, and given HIV Inconclusive results", SortOrder: 3},
		{Category: "hts_code", Code: "UKN", Label: "Unknown", Description: "Client was not tested, HIV status unknown", SortOrder: 4},

		// Switching reason codes
		{Category: "switching_reason", Code: "SE", Label: "Side Effects", SortOrder: 1},
		{Category: "switching_reason", Code: "OS", Label: "Stock Outs", SortOrder: 2},
		{Category: "switching_reason", Code: "CFI", Label: "Change in Fertility Intention", SortOrder: 3},
		{Category: "switching_reason", Code: "O", Label: "Others", SortOrder: 4},

		// Postpartum FP timing codes
		{Category: "postpartum_fp_timing", Code: "1", Label: "Within 48 hours", SortOrder: 1},
		{Category: "postpartum_fp_timing", Code: "2", Label: "3 days to 3 weeks", SortOrder: 2},
		{Category: "postpartum_fp_timing", Code: "3", Label: "4 weeks to 6 weeks", SortOrder: 3},
		{Category: "postpartum_fp_timing", Code: "4", Label: "7 weeks to 5 months", SortOrder: 4},
		{Category: "postpartum_fp_timing", Code: "5", Label: "6 months to 12 months", SortOrder: 5},

		// Post-abortion FP timing codes
		{Category: "post_abortion_fp_timing", Code: "1", Label: "Within 48 hours", SortOrder: 1},
		{Category: "post_abortion_fp_timing", Code: "2", Label: "3 days to 1 week", SortOrder: 2},
		{Category: "post_abortion_fp_timing", Code: "3", Label: "8 days to 2 weeks", SortOrder: 3},

		// LARC removal reason codes
		{Category: "larc_removal_reason", Code: "1", Label: "On-schedule", SortOrder: 1},
		{Category: "larc_removal_reason", Code: "2", Label: "Side Effects", SortOrder: 2},
		{Category: "larc_removal_reason", Code: "3", Label: "Change in Fertility Intention", SortOrder: 3},
		{Category: "larc_removal_reason", Code: "4", Label: "Gender Based Violence", SortOrder: 4},
		{Category: "larc_removal_reason", Code: "5", Label: "Others (Specify in Remarks)", SortOrder: 5},

		// LARC removal timing codes
		{Category: "larc_removal_timing", Code: "1", Label: "Within 3 months", SortOrder: 1},
		{Category: "larc_removal_timing", Code: "2", Label: "4 months to 6 months", SortOrder: 2},
		{Category: "larc_removal_timing", Code: "3", Label: "7 months to 12 months", SortOrder: 3},
		{Category: "larc_removal_timing", Code: "4", Label: "13 months to 18 months", SortOrder: 4},
		{Category: "larc_removal_timing", Code: "5", Label: "19 months to 35 months", SortOrder: 5},
		{Category: "larc_removal_timing", Code: "6", Label: "At 36 months", SortOrder: 6},
		{Category: "larc_removal_timing", Code: "7", Label: "Above 36 months", SortOrder: 7},

		// Side effect codes
		{Category: "side_effect", Code: "IB", Label: "Irregular Bleeding", Description: "Bleeding at the wrong times", SortOrder: 1},
		{Category: "side_effect", Code: "HB", Label: "Heavy Bleeding", SortOrder: 2},
		{Category: "side_effect", Code: "NB", Label: "No Monthly Bleeding", Description: "Lack of periods", SortOrder: 3},
		{Category: "side_effect", Code: "NV", Label: "Nausea, Vomiting or Dizziness", SortOrder: 4},
		{Category: "side_effect", Code: "ISR", Label: "Injection Site Reaction", Description: "For injectables", SortOrder: 5},
		{Category: "side_effect", Code: "INSR", Label: "Insertion Site Reaction", Description: "Abscess, infection etc. (for implants)", SortOrder: 6},
		{Category: "side_effect", Code: "HE", Label: "Headache", SortOrder: 7},
		{Category: "side_effect", Code: "MO", Label: "Mood Changes", SortOrder: 8},
		{Category: "side_effect", Code: "BR", Label: "Breast Tenderness", SortOrder: 9},
		{Category: "side_effect", Code: "CR", Label: "Cramping or Mild Abdominal Pain", SortOrder: 10},
		{Category: "side_effect", Code: "SAP", Label: "Severe Lower Abdominal Pain", SortOrder: 11},
		{Category: "side_effect", Code: "WG", Label: "Weight Gain", SortOrder: 12},
		{Category: "side_effect", Code: "SP", Label: "Suspected Pregnancy", SortOrder: 13},
		{Category: "side_effect", Code: "AC", Label: "Acne", SortOrder: 14},
		{Category: "side_effect", Code: "ETC", Label: "Others (Specify in Remarks)", SortOrder: 15},

		// Cervical cancer screening method
		{Category: "cervical_screening_method", Code: "1", Label: "HPV", SortOrder: 1},
		{Category: "cervical_screening_method", Code: "2", Label: "VIA", SortOrder: 2},
		{Category: "cervical_screening_method", Code: "3", Label: "Pap Smear", SortOrder: 3},

		// Cervical cancer status
		{Category: "cervical_cancer_status", Code: "1", Label: "Negative", SortOrder: 1},
		{Category: "cervical_cancer_status", Code: "2", Label: "Positive", SortOrder: 2},
		{Category: "cervical_cancer_status", Code: "3", Label: "Suspicious of Cancer", SortOrder: 3},
		{Category: "cervical_cancer_status", Code: "4", Label: "Invasive Cancer", SortOrder: 4},
		{Category: "cervical_cancer_status", Code: "5", Label: "Not Eligible", SortOrder: 5},
		{Category: "cervical_cancer_status", Code: "6", Label: "Not Done", SortOrder: 6},

		// Cervical cancer treatment
		{Category: "cervical_cancer_treatment", Code: "1", Label: "Thermocoagulation", SortOrder: 1},
		{Category: "cervical_cancer_treatment", Code: "2", Label: "LEEP", SortOrder: 2},
		{Category: "cervical_cancer_treatment", Code: "3", Label: "Cryotherapy", SortOrder: 3},

		// Breast cancer screening
		{Category: "breast_cancer_screening", Code: "FOM", Label: "Finding of Normal", Description: "No swellings, pain, abnormal discharge", SortOrder: 1},
		{Category: "breast_cancer_screening", Code: "SS", Label: "Suspicious Signs", Description: "Abnormal discharge (pus or blood)", SortOrder: 2},

		// Previous method / FP method codes
		{Category: "fp_method", Code: "COC", Label: "Combined Oral Contraceptives", SortOrder: 1},
		{Category: "fp_method", Code: "POP", Label: "Progestogen Only Pill", SortOrder: 2},
		{Category: "fp_method", Code: "ECP", Label: "Emergency Contraceptive Pill", SortOrder: 3},
		{Category: "fp_method", Code: "MC", Label: "Male Condom", SortOrder: 4},
		{Category: "fp_method", Code: "FC", Label: "Female Condom", SortOrder: 5},
		{Category: "fp_method", Code: "DMPA_IM", Label: "DMPA-IM (Injectable 3 Months)", SortOrder: 6},
		{Category: "fp_method", Code: "DMPA_SC", Label: "DMPA-SC (Injectable 3 Months)", SortOrder: 7},
		{Category: "fp_method", Code: "IMP3", Label: "Implant (3 Years)", SortOrder: 8},
		{Category: "fp_method", Code: "IMP5", Label: "Implant (5 Years)", SortOrder: 9},
		{Category: "fp_method", Code: "IUD_CU", Label: "IUD Copper-T", SortOrder: 10},
		{Category: "fp_method", Code: "IUD_H3", Label: "IUD Hormonal (3 Years)", SortOrder: 11},
		{Category: "fp_method", Code: "IUD_H5", Label: "IUD Hormonal (5 Years)", SortOrder: 12},
		{Category: "fp_method", Code: "TL", Label: "Tubal Ligation", SortOrder: 13},
		{Category: "fp_method", Code: "VAS", Label: "Vasectomy", SortOrder: 14},
		{Category: "fp_method", Code: "SDM", Label: "Standard Days Method", SortOrder: 15},
		{Category: "fp_method", Code: "LAM", Label: "Lactational Amenorrhea Method", SortOrder: 16},
		{Category: "fp_method", Code: "TDM", Label: "Two Day Method", SortOrder: 17},

		// Sex
		{Category: "sex", Code: "M", Label: "Male", SortOrder: 1},
		{Category: "sex", Code: "F", Label: "Female", SortOrder: 2},
	}

	db.CreateInBatches(sets, 50)
}
