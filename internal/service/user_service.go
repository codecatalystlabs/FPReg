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

	actor, _ := s.repo.FindByID(actorID)
	var actorRole models.Role
	var actorFacilityID *uuid.UUID
	if actor != nil {
		actorRole = actor.Role
		actorFacilityID = actor.FacilityID
	}

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

	// Facility admin can only create users with roles below them (facility_user or reviewer)
	// and only within their own facility.
	if actorRole == models.RoleFacilityAdmin {
		if role == models.RoleSuperAdmin || role == models.RoleFacilityAdmin {
			errs = append(errs, utils.ErrorDetail{Field: "role", Message: "Facility admin can only assign Facility User or Reviewer roles"})
		}
		if actorFacilityID == nil {
			errs = append(errs, utils.ErrorDetail{Field: "facility_id", Message: "Your account is not linked to a facility"})
		} else {
			// Force created user into actor's facility
			input.FacilityID = actorFacilityID
		}
	}

	// Non-superadmin (including facility_admin) must always belong to a facility
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

	actor, _ := s.repo.FindByID(actorID)
	var actorRole models.Role
	var actorFacilityID *uuid.UUID
	if actor != nil {
		actorRole = actor.Role
		actorFacilityID = actor.FacilityID
	}

	// Only superadmin can modify a superadmin user
	if user.Role == models.RoleSuperAdmin && actorRole != models.RoleSuperAdmin {
		return nil, errors.New("only superadmin can modify a superadmin user")
	}

	if strings.TrimSpace(input.FullName) != "" {
		user.FullName = strings.TrimSpace(input.FullName)
	}
	newRole := user.Role
	if input.Role != "" {
		newRole = models.Role(input.Role)
	}

	// Facility admin can only assign roles below them and only within their own facility
	if actorRole == models.RoleFacilityAdmin {
		if newRole == models.RoleSuperAdmin || newRole == models.RoleFacilityAdmin {
			return nil, errors.New("facility admin can only assign Facility User or Reviewer roles")
		}
		if actorFacilityID == nil {
			return nil, errors.New("your account is not linked to a facility")
		}
		// They can only manage users in their own facility
		if user.FacilityID == nil || *user.FacilityID != *actorFacilityID {
			return nil, errors.New("you can only manage users in your facility")
		}
		// Force user to remain in actor's facility
		input.FacilityID = actorFacilityID
	}

	user.Role = newRole

	// Always apply facility_id on update so superadmin can move user to another facility or clear facility
	user.FacilityID = input.FacilityID
	if user.Role != models.RoleSuperAdmin && user.FacilityID == nil {
		return nil, errors.New("facility is required for non-superadmin users")
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
