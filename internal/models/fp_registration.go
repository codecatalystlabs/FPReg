package models

import "github.com/google/uuid"

type FPRegistration struct {
	BaseModel

	// Facility & user context
	FacilityID uuid.UUID `gorm:"type:uuid;not null;index" json:"facility_id"`
	Facility   Facility  `gorm:"foreignKey:FacilityID" json:"facility,omitempty"`
	CreatedBy  uuid.UUID `gorm:"type:uuid;not null;index" json:"created_by"`
	Creator    User      `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`

	// Column 1 – Serial number (resets monthly)
	VisitDate    string `gorm:"size:10;not null;index" json:"visit_date"`
	SerialNumber int    `gorm:"not null" json:"serial_number"`

	// Column 2 – Client number (auto-generated, nullable for visitors)
	ClientNumber *string `gorm:"size:30;index" json:"client_number,omitempty"`
	IsVisitor    bool    `gorm:"default:false" json:"is_visitor"`

	// Column 3 – NIN
	NIN string `gorm:"size:30" json:"nin,omitempty"`

	// Column 4 – Client name & contact
	Surname     string `gorm:"size:100;not null" json:"surname"`
	GivenName   string `gorm:"size:100;not null" json:"given_name"`
	PhoneNumber string `gorm:"size:20" json:"phone_number,omitempty"`

	// Column 5 – Physical address
	Village   string `gorm:"size:100" json:"village,omitempty"`
	Parish    string `gorm:"size:100" json:"parish,omitempty"`
	Subcounty string `gorm:"size:100" json:"subcounty,omitempty"`
	District  string `gorm:"size:100" json:"district,omitempty"`

	// Column 6 – Sex
	Sex string `gorm:"size:1;not null" json:"sex"`

	// Column 7 – Age
	Age int `gorm:"not null" json:"age"`

	// Column 8 – New user (first time FP user)
	IsNewUser bool `gorm:"default:false" json:"is_new_user"`

	// Column 9 – Revisit
	IsRevisit bool `gorm:"default:false" json:"is_revisit"`

	// Column 10 – Previous method used (if revisit)
	PreviousMethod string `gorm:"size:50" json:"previous_method,omitempty"`

	// Column 11 – HTS code
	HTSCode string `gorm:"size:10" json:"hts_code,omitempty"`

	// Column 12 – FP counseling
	CounselingIndividual bool `gorm:"default:false" json:"counseling_individual"`
	CounselingAsCouple   bool `gorm:"default:false" json:"counseling_as_couple"`
	CounselingOM         bool `gorm:"default:false" json:"counseling_om"`
	CounselingSE         bool `gorm:"default:false" json:"counseling_se"`
	CounselingWD         bool `gorm:"default:false" json:"counseling_wd"`
	CounselingMS         bool `gorm:"default:false" json:"counseling_ms"`

	// Column 13 – Switching method
	IsSwitching     bool   `gorm:"default:false" json:"is_switching"`
	SwitchingReason string `gorm:"size:10" json:"switching_reason,omitempty"`

	// Column 14 – Oral pills
	PillsCOCCycles int `gorm:"default:0" json:"pills_coc_cycles"`
	PillsPOPCycles int `gorm:"default:0" json:"pills_pop_cycles"`
	PillsECPPieces int `gorm:"default:0" json:"pills_ecp_pieces"`

	// Column 15 – Condoms
	CondomsMaleUnits   int `gorm:"default:0" json:"condoms_male_units"`
	CondomsFemaleUnits int `gorm:"default:0" json:"condoms_female_units"`

	// Column 16 – Injectables
	InjectableDMPAIMDoses  int `gorm:"default:0" json:"injectable_dmpa_im_doses"`
	InjectableDMPASCPADoses int `gorm:"default:0" json:"injectable_dmpa_sc_pa_doses"`
	InjectableDMPASCSIDoses int `gorm:"default:0" json:"injectable_dmpa_sc_si_doses"`

	// Column 17 – Implants
	Implant3Years bool `gorm:"default:false" json:"implant_3_years"`
	Implant5Years bool `gorm:"default:false" json:"implant_5_years"`

	// Column 18 – IUDs
	IUDCopperT         bool `gorm:"default:false" json:"iud_copper_t"`
	IUDHormonal3Years  bool `gorm:"default:false" json:"iud_hormonal_3_years"`
	IUDHormonal5Years  bool `gorm:"default:false" json:"iud_hormonal_5_years"`

	// Column 19 – Sterilization
	TubalLigation bool `gorm:"default:false" json:"tubal_ligation"`
	Vasectomy     bool `gorm:"default:false" json:"vasectomy"`

	// Column 20 – Fertility awareness methods
	FAMStandardDays bool `gorm:"default:false" json:"fam_standard_days"`
	FAMLAM          bool `gorm:"default:false" json:"fam_lam"`
	FAMTwoDay       bool `gorm:"default:false" json:"fam_two_day"`

	// Column 21 – Post-pregnancy FP
	PostPartumFPTiming  string `gorm:"size:5" json:"postpartum_fp_timing,omitempty"`
	PostAbortionFPTiming string `gorm:"size:5" json:"post_abortion_fp_timing,omitempty"`

	// Column 22 – LARC removal
	ImplantRemovalReason string `gorm:"size:5" json:"implant_removal_reason,omitempty"`
	ImplantRemovalTiming string `gorm:"size:5" json:"implant_removal_timing,omitempty"`
	IUDRemovalReason     string `gorm:"size:5" json:"iud_removal_reason,omitempty"`
	IUDRemovalTiming     string `gorm:"size:5" json:"iud_removal_timing,omitempty"`

	// Column 23 – Side effects (comma-separated codes)
	SideEffects string `gorm:"size:200" json:"side_effects,omitempty"`

	// Column 24 – Cancer screening
	CervicalCancerScreeningMethod string `gorm:"size:5" json:"cervical_screening_method,omitempty"`
	CervicalCancerStatus          string `gorm:"size:5" json:"cervical_cancer_status,omitempty"`
	CervicalCancerTreatment       string `gorm:"size:5" json:"cervical_cancer_treatment,omitempty"`
	BreastCancerScreening         string `gorm:"size:10" json:"breast_cancer_screening,omitempty"`

	// Column 25 – STI screening
	ScreenedForSTI *bool `gorm:"" json:"screened_for_sti,omitempty"`

	// Column 26 – Referral
	ReferralNumber string `gorm:"size:50" json:"referral_number,omitempty"`
	ReferralReason string `gorm:"size:300" json:"referral_reason,omitempty"`

	// Column 27 – Remarks
	Remarks string `gorm:"type:text" json:"remarks,omitempty"`
}
