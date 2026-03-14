package service

import (
	"errors"
	"strings"

	"fpreg/internal/models"
	"fpreg/internal/repository"
	"fpreg/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo     *repository.UserRepository
	auditSvc *AuditService
}

func NewUserService(repo *repository.UserRepository, auditSvc *AuditService) *UserService {
	return &UserService{repo: repo, auditSvc: auditSvc}
}

type CreateUserInput struct {
	Email      string     `json:"email"`
	Password   string     `json:"password"`
	FullName   string     `json:"full_name"`
	Role       string     `json:"role"`
	FacilityID *uuid.UUID `json:"facility_id"`
}

func (s *UserService) Create(input CreateUserInput, actorID uuid.UUID, ip, ua string) (*models.User, []utils.ErrorDetail) {
	var errs []utils.ErrorDetail

	if !utils.IsValidEmail(input.Email) {
		errs = append(errs, utils.ErrorDetail{Field: "email", Message: "Valid email is required"})
	}
	if len(input.Password) < 8 {
		errs = append(errs, utils.ErrorDetail{Field: "password", Message: "Password must be at least 8 characters"})
	}
	if strings.TrimSpace(input.FullName) == "" {
		errs = append(errs, utils.ErrorDetail{Field: "full_name", Message: "Full name is required"})
	}

	role := models.Role(input.Role)
	if role != models.RoleSuperAdmin && role != models.RoleFacilityAdmin &&
		role != models.RoleFacilityUser && role != models.RoleReviewer {
		errs = append(errs, utils.ErrorDetail{Field: "role", Message: "Invalid role"})
	}

	if role != models.RoleSuperAdmin && input.FacilityID == nil {
		errs = append(errs, utils.ErrorDetail{Field: "facility_id", Message: "Facility is required for non-superadmin users"})
	}

	if len(errs) > 0 {
		return nil, errs
	}

	if existing, _ := s.repo.FindByEmail(input.Email); existing != nil && existing.ID != uuid.Nil {
		return nil, []utils.ErrorDetail{{Field: "email", Message: "Email already in use"}}
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	user := models.User{
		Email:      input.Email,
		Password:   string(hash),
		FullName:   strings.TrimSpace(input.FullName),
		Role:       role,
		FacilityID: input.FacilityID,
		IsActive:   true,
	}

	if err := s.repo.Create(&user); err != nil {
		return nil, []utils.ErrorDetail{{Message: "Failed to create user"}}
	}

	s.auditSvc.Log(&actorID, input.FacilityID, models.AuditCreate,
		"user", user.ID.String(), ip, ua, "Created user: "+user.Email)

	return &user, nil
}

func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) List(page, perPage int, facilityID *uuid.UUID) ([]models.User, int64, error) {
	return s.repo.List(page, perPage, facilityID)
}

func (s *UserService) Update(id uuid.UUID, input CreateUserInput, actorID uuid.UUID, ip, ua string) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if strings.TrimSpace(input.FullName) != "" {
		user.FullName = strings.TrimSpace(input.FullName)
	}
	if input.Role != "" {
		user.Role = models.Role(input.Role)
	}
	if input.FacilityID != nil {
		user.FacilityID = input.FacilityID
	}
	if input.Password != "" && len(input.Password) >= 8 {
		hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		user.Password = string(hash)
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	s.auditSvc.Log(&actorID, user.FacilityID, models.AuditUpdate,
		"user", user.ID.String(), ip, ua, "Updated user: "+user.Email)

	return user, nil
}

func (s *UserService) Deactivate(id uuid.UUID, actorID uuid.UUID, ip, ua string) error {
	user, err := s.repo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	user.IsActive = false
	if err := s.repo.Update(user); err != nil {
		return err
	}
	s.auditSvc.Log(&actorID, user.FacilityID, models.AuditAdminAction,
		"user", user.ID.String(), ip, ua, "Deactivated user: "+user.Email)
	return nil
}
