package service

import (
	"strings"

	"fpreg/internal/models"
	"fpreg/internal/repository"
	"fpreg/internal/utils"

	"github.com/google/uuid"
)

type FacilityService struct {
	repo     *repository.FacilityRepository
	auditSvc *AuditService
}

func NewFacilityService(repo *repository.FacilityRepository, auditSvc *AuditService) *FacilityService {
	return &FacilityService{repo: repo, auditSvc: auditSvc}
}

type CreateFacilityInput struct {
	Name             string `json:"name"`
	Code             string `json:"code"`
	Level            string `json:"level"`
	Subcounty        string `json:"subcounty"`
	HSD              string `json:"hsd"`
	District         string `json:"district"`
	ClientCodePrefix string `json:"client_code_prefix"`
}

func (s *FacilityService) Create(input CreateFacilityInput, actorID uuid.UUID, ip, ua string) (*models.Facility, []utils.ErrorDetail) {
	var errs []utils.ErrorDetail

	if strings.TrimSpace(input.Name) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "name", Message: "Facility name is required"})
	}
	if strings.TrimSpace(input.Code) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "code", Message: "Facility code is required"})
	}
	if strings.TrimSpace(input.ClientCodePrefix) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "client_code_prefix", Message: "Client code prefix is required"})
	}
	if len(input.ClientCodePrefix) > 5 {
		errs = append(errs, utils.ErrorDetail{Field: "client_code_prefix", Message: "Client code prefix must be 5 characters or less"})
	}

	if len(errs) > 0 {
		return nil, errs
	}

	if existing, _ := s.repo.FindByCode(input.Code); existing != nil && existing.ID != uuid.Nil {
		return nil, []utils.ErrorDetail{{Field: "code", Message: "Facility code already exists"}}
	}

	f := models.Facility{
		Name:             strings.TrimSpace(input.Name),
		Code:             strings.ToUpper(strings.TrimSpace(input.Code)),
		Level:            strings.TrimSpace(input.Level),
		Subcounty:        strings.TrimSpace(input.Subcounty),
		HSD:              strings.TrimSpace(input.HSD),
		District:         strings.TrimSpace(input.District),
		ClientCodePrefix: strings.ToUpper(strings.TrimSpace(input.ClientCodePrefix)),
	}

	if err := s.repo.Create(&f); err != nil {
		return nil, []utils.ErrorDetail{{Message: "Failed to create facility"}}
	}

	s.auditSvc.Log(&actorID, nil, models.AuditCreate,
		"facility", f.ID.String(), ip, ua, "Created facility: "+f.Name)

	return &f, nil
}

func (s *FacilityService) GetByID(id uuid.UUID) (*models.Facility, error) {
	return s.repo.FindByID(id)
}

func (s *FacilityService) List(page, perPage int, search string) ([]models.Facility, int64, error) {
	return s.repo.List(page, perPage, search)
}

// ListByDistrict lists facilities in a single district (for district_biostatistician UI).
func (s *FacilityService) ListByDistrict(page, perPage int, district, search string) ([]models.Facility, int64, error) {
	return s.repo.ListByDistrict(page, perPage, district, search)
}

// ListDistinctDistricts returns district names present on at least one facility.
func (s *FacilityService) ListDistinctDistricts() ([]string, error) {
	return s.repo.ListDistinctDistricts()
}

func (s *FacilityService) Update(id uuid.UUID, input CreateFacilityInput, actorID uuid.UUID, ip, ua string) (*models.Facility, error) {
	f, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if input.Name != "" {
		f.Name = strings.TrimSpace(input.Name)
	}
	if input.Level != "" {
		f.Level = strings.TrimSpace(input.Level)
	}
	if input.Subcounty != "" {
		f.Subcounty = strings.TrimSpace(input.Subcounty)
	}
	if input.HSD != "" {
		f.HSD = strings.TrimSpace(input.HSD)
	}
	if input.District != "" {
		f.District = strings.TrimSpace(input.District)
	}
	if input.ClientCodePrefix != "" {
		f.ClientCodePrefix = strings.ToUpper(strings.TrimSpace(input.ClientCodePrefix))
	}

	if err := s.repo.Update(f); err != nil {
		return nil, err
	}

	s.auditSvc.Log(&actorID, nil, models.AuditUpdate,
		"facility", f.ID.String(), ip, ua, "Updated facility: "+f.Name)
	return f, nil
}

func (s *FacilityService) Delete(id uuid.UUID, actorID uuid.UUID, ip, ua string) error {
	f, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	s.auditSvc.Log(&actorID, nil, models.AuditDelete,
		"facility", id.String(), ip, ua, "Deleted facility: "+f.Name)
	return s.repo.Delete(id)
}
