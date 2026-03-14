package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"fpreg/internal/models"
	"fpreg/internal/repository"
	"fpreg/internal/utils"

	"github.com/google/uuid"
)

type RegistrationService struct {
	regRepo      *repository.RegistrationRepository
	clientNumRepo *repository.ClientNumberRepository
	facilityRepo *repository.FacilityRepository
	auditSvc     *AuditService
}

func NewRegistrationService(
	regRepo *repository.RegistrationRepository,
	clientNumRepo *repository.ClientNumberRepository,
	facilityRepo *repository.FacilityRepository,
	auditSvc *AuditService,
) *RegistrationService {
	return &RegistrationService{
		regRepo:       regRepo,
		clientNumRepo: clientNumRepo,
		facilityRepo:  facilityRepo,
		auditSvc:      auditSvc,
	}
}

type CreateRegistrationInput struct {
	VisitDate   string `json:"visit_date"`
	IsVisitor   bool   `json:"is_visitor"`
	NIN         string `json:"nin"`
	Surname     string `json:"surname"`
	GivenName   string `json:"given_name"`
	PhoneNumber string `json:"phone_number"`
	Village     string `json:"village"`
	Parish      string `json:"parish"`
	Subcounty   string `json:"subcounty"`
	District    string `json:"district"`
	Sex         string `json:"sex"`
	Age         int    `json:"age"`

	IsNewUser      bool   `json:"is_new_user"`
	IsRevisit      bool   `json:"is_revisit"`
	PreviousMethod string `json:"previous_method"`

	HTSCode              string `json:"hts_code"`
	CounselingIndividual bool   `json:"counseling_individual"`
	CounselingAsCouple   bool   `json:"counseling_as_couple"`
	CounselingOM         bool   `json:"counseling_om"`
	CounselingSE         bool   `json:"counseling_se"`
	CounselingWD         bool   `json:"counseling_wd"`
	CounselingMS         bool   `json:"counseling_ms"`

	IsSwitching     bool   `json:"is_switching"`
	SwitchingReason string `json:"switching_reason"`

	PillsCOCCycles int `json:"pills_coc_cycles"`
	PillsPOPCycles int `json:"pills_pop_cycles"`
	PillsECPPieces int `json:"pills_ecp_pieces"`

	CondomsMaleUnits   int `json:"condoms_male_units"`
	CondomsFemaleUnits int `json:"condoms_female_units"`

	InjectableDMPAIMDoses   int `json:"injectable_dmpa_im_doses"`
	InjectableDMPASCPADoses int `json:"injectable_dmpa_sc_pa_doses"`
	InjectableDMPASCSIDoses int `json:"injectable_dmpa_sc_si_doses"`

	Implant3Years bool `json:"implant_3_years"`
	Implant5Years bool `json:"implant_5_years"`

	IUDCopperT        bool `json:"iud_copper_t"`
	IUDHormonal3Years bool `json:"iud_hormonal_3_years"`
	IUDHormonal5Years bool `json:"iud_hormonal_5_years"`

	TubalLigation bool `json:"tubal_ligation"`
	Vasectomy     bool `json:"vasectomy"`

	FAMStandardDays bool `json:"fam_standard_days"`
	FAMLAM          bool `json:"fam_lam"`
	FAMTwoDay       bool `json:"fam_two_day"`

	PostPartumFPTiming   string `json:"postpartum_fp_timing"`
	PostAbortionFPTiming string `json:"post_abortion_fp_timing"`

	ImplantRemovalReason string `json:"implant_removal_reason"`
	ImplantRemovalTiming string `json:"implant_removal_timing"`
	IUDRemovalReason     string `json:"iud_removal_reason"`
	IUDRemovalTiming     string `json:"iud_removal_timing"`

	SideEffects string `json:"side_effects"`

	CervicalScreeningMethod string `json:"cervical_screening_method"`
	CervicalCancerStatus    string `json:"cervical_cancer_status"`
	CervicalCancerTreatment string `json:"cervical_cancer_treatment"`
	BreastCancerScreening   string `json:"breast_cancer_screening"`

	ScreenedForSTI *bool `json:"screened_for_sti"`

	ReferralNumber string `json:"referral_number"`
	ReferralReason string `json:"referral_reason"`
	Remarks        string `json:"remarks"`
}

func (s *RegistrationService) Create(input CreateRegistrationInput, facilityID, userID uuid.UUID, ip, ua string) (*models.FPRegistration, []utils.ErrorDetail) {
	errs := s.validateInput(input)
	if len(errs) > 0 {
		return nil, errs
	}

	facility, err := s.facilityRepo.FindByID(facilityID)
	if err != nil {
		return nil, []utils.ErrorDetail{{Message: "Facility not found"}}
	}

	tx := s.regRepo.DB().Begin()

	serialNum, err := s.regRepo.NextSerialNumber(tx, facilityID, input.VisitDate)
	if err != nil {
		tx.Rollback()
		return nil, []utils.ErrorDetail{{Message: "Failed to generate serial number"}}
	}

	var clientNumber *string
	if !input.IsVisitor {
		visitDate, _ := time.Parse("2006-01-02", input.VisitDate)
		cn, err := s.clientNumRepo.NextClientNumber(tx, facilityID, facility.ClientCodePrefix, visitDate)
		if err != nil {
			tx.Rollback()
			return nil, []utils.ErrorDetail{{Message: "Failed to generate client number"}}
		}
		clientNumber = &cn
	}

	reg := models.FPRegistration{
		FacilityID:   facilityID,
		CreatedBy:    userID,
		VisitDate:    input.VisitDate,
		SerialNumber: serialNum,
		ClientNumber: clientNumber,
		IsVisitor:    input.IsVisitor,

		NIN:         strings.TrimSpace(input.NIN),
		Surname:     strings.TrimSpace(input.Surname),
		GivenName:   strings.TrimSpace(input.GivenName),
		PhoneNumber: strings.TrimSpace(input.PhoneNumber),
		Village:     strings.TrimSpace(input.Village),
		Parish:      strings.TrimSpace(input.Parish),
		Subcounty:   strings.TrimSpace(input.Subcounty),
		District:    strings.TrimSpace(input.District),
		Sex:         strings.ToUpper(strings.TrimSpace(input.Sex)),
		Age:         input.Age,

		IsNewUser:      input.IsNewUser,
		IsRevisit:      input.IsRevisit,
		PreviousMethod: input.PreviousMethod,
		HTSCode:        input.HTSCode,

		CounselingIndividual: input.CounselingIndividual,
		CounselingAsCouple:   input.CounselingAsCouple,
		CounselingOM:         input.CounselingOM,
		CounselingSE:         input.CounselingSE,
		CounselingWD:         input.CounselingWD,
		CounselingMS:         input.CounselingMS,

		IsSwitching:     input.IsSwitching,
		SwitchingReason: input.SwitchingReason,

		PillsCOCCycles: input.PillsCOCCycles,
		PillsPOPCycles: input.PillsPOPCycles,
		PillsECPPieces: input.PillsECPPieces,

		CondomsMaleUnits:   input.CondomsMaleUnits,
		CondomsFemaleUnits: input.CondomsFemaleUnits,

		InjectableDMPAIMDoses:   input.InjectableDMPAIMDoses,
		InjectableDMPASCPADoses: input.InjectableDMPASCPADoses,
		InjectableDMPASCSIDoses: input.InjectableDMPASCSIDoses,

		Implant3Years:     input.Implant3Years,
		Implant5Years:     input.Implant5Years,
		IUDCopperT:        input.IUDCopperT,
		IUDHormonal3Years: input.IUDHormonal3Years,
		IUDHormonal5Years: input.IUDHormonal5Years,
		TubalLigation:     input.TubalLigation,
		Vasectomy:         input.Vasectomy,

		FAMStandardDays: input.FAMStandardDays,
		FAMLAM:          input.FAMLAM,
		FAMTwoDay:       input.FAMTwoDay,

		PostPartumFPTiming:   input.PostPartumFPTiming,
		PostAbortionFPTiming: input.PostAbortionFPTiming,

		ImplantRemovalReason: input.ImplantRemovalReason,
		ImplantRemovalTiming: input.ImplantRemovalTiming,
		IUDRemovalReason:     input.IUDRemovalReason,
		IUDRemovalTiming:     input.IUDRemovalTiming,

		SideEffects: input.SideEffects,

		CervicalCancerScreeningMethod: input.CervicalScreeningMethod,
		CervicalCancerStatus:          input.CervicalCancerStatus,
		CervicalCancerTreatment:       input.CervicalCancerTreatment,
		BreastCancerScreening:         input.BreastCancerScreening,

		ScreenedForSTI: input.ScreenedForSTI,
		ReferralNumber: input.ReferralNumber,
		ReferralReason: input.ReferralReason,
		Remarks:        input.Remarks,
	}

	if err := s.regRepo.Create(tx, &reg); err != nil {
		tx.Rollback()
		return nil, []utils.ErrorDetail{{Message: "Failed to save registration"}}
	}

	tx.Commit()

	s.auditSvc.LogWithValues(&userID, &facilityID, models.AuditCreate,
		"fp_registration", reg.ID.String(), ip, ua, nil, reg)

	return &reg, nil
}

func (s *RegistrationService) Update(id uuid.UUID, input CreateRegistrationInput, userID uuid.UUID, ip, ua string) (*models.FPRegistration, []utils.ErrorDetail) {
	existing, err := s.regRepo.FindByID(id)
	if err != nil {
		return nil, []utils.ErrorDetail{{Message: "Registration not found"}}
	}

	errs := s.validateInput(input)
	if len(errs) > 0 {
		return nil, errs
	}

	oldJSON, _ := json.Marshal(existing)

	existing.NIN = strings.TrimSpace(input.NIN)
	existing.Surname = strings.TrimSpace(input.Surname)
	existing.GivenName = strings.TrimSpace(input.GivenName)
	existing.PhoneNumber = strings.TrimSpace(input.PhoneNumber)
	existing.Village = strings.TrimSpace(input.Village)
	existing.Parish = strings.TrimSpace(input.Parish)
	existing.Subcounty = strings.TrimSpace(input.Subcounty)
	existing.District = strings.TrimSpace(input.District)
	existing.Sex = strings.ToUpper(strings.TrimSpace(input.Sex))
	existing.Age = input.Age
	existing.IsNewUser = input.IsNewUser
	existing.IsRevisit = input.IsRevisit
	existing.PreviousMethod = input.PreviousMethod
	existing.HTSCode = input.HTSCode
	existing.CounselingIndividual = input.CounselingIndividual
	existing.CounselingAsCouple = input.CounselingAsCouple
	existing.CounselingOM = input.CounselingOM
	existing.CounselingSE = input.CounselingSE
	existing.CounselingWD = input.CounselingWD
	existing.CounselingMS = input.CounselingMS
	existing.IsSwitching = input.IsSwitching
	existing.SwitchingReason = input.SwitchingReason
	existing.PillsCOCCycles = input.PillsCOCCycles
	existing.PillsPOPCycles = input.PillsPOPCycles
	existing.PillsECPPieces = input.PillsECPPieces
	existing.CondomsMaleUnits = input.CondomsMaleUnits
	existing.CondomsFemaleUnits = input.CondomsFemaleUnits
	existing.InjectableDMPAIMDoses = input.InjectableDMPAIMDoses
	existing.InjectableDMPASCPADoses = input.InjectableDMPASCPADoses
	existing.InjectableDMPASCSIDoses = input.InjectableDMPASCSIDoses
	existing.Implant3Years = input.Implant3Years
	existing.Implant5Years = input.Implant5Years
	existing.IUDCopperT = input.IUDCopperT
	existing.IUDHormonal3Years = input.IUDHormonal3Years
	existing.IUDHormonal5Years = input.IUDHormonal5Years
	existing.TubalLigation = input.TubalLigation
	existing.Vasectomy = input.Vasectomy
	existing.FAMStandardDays = input.FAMStandardDays
	existing.FAMLAM = input.FAMLAM
	existing.FAMTwoDay = input.FAMTwoDay
	existing.PostPartumFPTiming = input.PostPartumFPTiming
	existing.PostAbortionFPTiming = input.PostAbortionFPTiming
	existing.ImplantRemovalReason = input.ImplantRemovalReason
	existing.ImplantRemovalTiming = input.ImplantRemovalTiming
	existing.IUDRemovalReason = input.IUDRemovalReason
	existing.IUDRemovalTiming = input.IUDRemovalTiming
	existing.SideEffects = input.SideEffects
	existing.CervicalCancerScreeningMethod = input.CervicalScreeningMethod
	existing.CervicalCancerStatus = input.CervicalCancerStatus
	existing.CervicalCancerTreatment = input.CervicalCancerTreatment
	existing.BreastCancerScreening = input.BreastCancerScreening
	existing.ScreenedForSTI = input.ScreenedForSTI
	existing.ReferralNumber = input.ReferralNumber
	existing.ReferralReason = input.ReferralReason
	existing.Remarks = input.Remarks

	if err := s.regRepo.Update(existing); err != nil {
		return nil, []utils.ErrorDetail{{Message: "Failed to update registration"}}
	}

	var oldMap interface{}
	_ = json.Unmarshal(oldJSON, &oldMap)
	s.auditSvc.LogWithValues(&userID, &existing.FacilityID, models.AuditUpdate,
		"fp_registration", existing.ID.String(), ip, ua, oldMap, existing)

	return existing, nil
}

func (s *RegistrationService) GetByID(id uuid.UUID) (*models.FPRegistration, error) {
	return s.regRepo.FindByID(id)
}

func (s *RegistrationService) List(page, perPage int, f repository.RegistrationFilter) ([]models.FPRegistration, int64, error) {
	return s.regRepo.List(page, perPage, f)
}

func (s *RegistrationService) Delete(id uuid.UUID, userID uuid.UUID, ip, ua string) error {
	reg, err := s.regRepo.FindByID(id)
	if err != nil {
		return errors.New("registration not found")
	}
	if err := s.regRepo.SoftDelete(id); err != nil {
		return err
	}
	s.auditSvc.LogWithValues(&userID, &reg.FacilityID, models.AuditDelete,
		"fp_registration", id.String(), ip, ua, reg, nil)
	return nil
}

func (s *RegistrationService) validateInput(input CreateRegistrationInput) []utils.ErrorDetail {
	var errs []utils.ErrorDetail

	if strings.TrimSpace(input.VisitDate) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "visit_date", Message: "Visit date is required"})
	} else if !utils.IsValidDate(input.VisitDate) {
		errs = append(errs, utils.ErrorDetail{Field: "visit_date", Message: "Visit date must be YYYY-MM-DD"})
	}
	if strings.TrimSpace(input.Surname) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "surname", Message: "Surname is required"})
	}
	if strings.TrimSpace(input.GivenName) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "given_name", Message: "Given name is required"})
	}
	if !utils.IsValidSex(input.Sex) {
		errs = append(errs, utils.ErrorDetail{Field: "sex", Message: "Sex must be M or F"})
	}
	if input.Age < 0 || input.Age > 120 {
		errs = append(errs, utils.ErrorDetail{Field: "age", Message: "Age must be between 0 and 120"})
	}
	if input.IsNewUser && input.IsRevisit {
		errs = append(errs, utils.ErrorDetail{Field: "is_new_user", Message: "Cannot be both new user and revisit"})
	}
	if !input.IsNewUser && !input.IsRevisit {
		errs = append(errs, utils.ErrorDetail{Field: "is_new_user", Message: "Must be either new user or revisit"})
	}
	if input.IsRevisit && strings.TrimSpace(input.PreviousMethod) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "previous_method", Message: "Previous method required for revisit"})
	}
	if input.IsSwitching && strings.TrimSpace(input.SwitchingReason) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "switching_reason", Message: "Reason required when switching method"})
	}

	return errs
}
