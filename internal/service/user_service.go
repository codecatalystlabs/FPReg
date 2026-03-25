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
	repo         *repository.UserRepository
	facilityRepo *repository.FacilityRepository
	auditSvc     *AuditService
}

func NewUserService(repo *repository.UserRepository, facilityRepo *repository.FacilityRepository, auditSvc *AuditService) *UserService {
	return &UserService{repo: repo, facilityRepo: facilityRepo, auditSvc: auditSvc}
}

type CreateUserInput struct {
	Email      string     `json:"email"`
	Password   string     `json:"password"`
	FullName   string     `json:"full_name"`
	Role       string     `json:"role"`
	FacilityID *uuid.UUID `json:"facility_id"`
	District   string     `json:"district"`
}

func (s *UserService) userBelongsToDistrict(u *models.User, district string) bool {
	d := strings.TrimSpace(district)
	if d == "" || u == nil {
		return false
	}
	if u.Role == models.RoleDistrictBiostatistician {
		return strings.EqualFold(strings.TrimSpace(u.District), d)
	}
	if u.FacilityID == nil {
		return false
	}
	ok, err := s.facilityRepo.FacilityBelongsToDistrict(*u.FacilityID, d)
	return err == nil && ok
}

// CanActorAccessUser returns whether actor may view or manage the target user via admin APIs.
func (s *UserService) CanActorAccessUser(actorID, targetID uuid.UUID) bool {
	if actorID == targetID {
		return true
	}
	actor, errA := s.repo.FindByID(actorID)
	target, errT := s.repo.FindByID(targetID)
	if errA != nil || errT != nil || actor == nil || target == nil {
		return false
	}
	switch actor.Role {
	case models.RoleSuperAdmin:
		return true
	case models.RoleFacilityAdmin:
		if actor.FacilityID == nil || target.FacilityID == nil {
			return false
		}
		return *actor.FacilityID == *target.FacilityID
	case models.RoleDistrictBiostatistician:
		if target.Role == models.RoleSuperAdmin || target.Role == models.RoleFacilityAdmin {
			return false
		}
		if target.Role == models.RoleDistrictBiostatistician {
			return actor.ID == target.ID
		}
		return s.userBelongsToDistrict(target, actor.District)
	default:
		return false
	}
}

func (s *UserService) Create(input CreateUserInput, actorID uuid.UUID, ip, ua string) (*models.User, []utils.ErrorDetail) {
	var errs []utils.ErrorDetail

	actor, _ := s.repo.FindByID(actorID)
	var actorRole models.Role
	var actorFacilityID *uuid.UUID
	var actorDistrict string
	if actor != nil {
		actorRole = actor.Role
		actorFacilityID = actor.FacilityID
		actorDistrict = strings.TrimSpace(actor.District)
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
		role != models.RoleFacilityUser && role != models.RoleReviewer &&
		role != models.RoleDistrictBiostatistician {
		errs = append(errs, utils.ErrorDetail{Field: "role", Message: "Invalid role"})
	}

	// District biostatistician: district required, no facility
	if role == models.RoleDistrictBiostatistician {
		if strings.TrimSpace(input.District) == "" {
			errs = append(errs, utils.ErrorDetail{Field: "district", Message: "District is required for district biostatistician"})
		}
		if input.FacilityID != nil {
			errs = append(errs, utils.ErrorDetail{Field: "facility_id", Message: "Facility must not be set for district biostatistician; use district only"})
		}
	} else if role != models.RoleSuperAdmin && input.FacilityID == nil {
		errs = append(errs, utils.ErrorDetail{Field: "facility_id", Message: "Facility is required for this role"})
	}

	// Facility admin can only create users with roles below them (facility_user or reviewer)
	// and only within their own facility.
	if actorRole == models.RoleFacilityAdmin {
		if role == models.RoleSuperAdmin || role == models.RoleFacilityAdmin || role == models.RoleDistrictBiostatistician {
			errs = append(errs, utils.ErrorDetail{Field: "role", Message: "Facility admin can only assign Facility User or Reviewer roles"})
		}
		if actorFacilityID == nil {
			errs = append(errs, utils.ErrorDetail{Field: "facility_id", Message: "Your account is not linked to a facility"})
		} else {
			input.FacilityID = actorFacilityID
		}
		input.District = ""
	}

	// District biostatistician can create facility_user / reviewer only, facility must be in their district
	if actorRole == models.RoleDistrictBiostatistician {
		if actorDistrict == "" {
			errs = append(errs, utils.ErrorDetail{Field: "district", Message: "Your account has no district assigned"})
		}
		if role != models.RoleFacilityUser && role != models.RoleReviewer {
			errs = append(errs, utils.ErrorDetail{Field: "role", Message: "You can only create Facility User or Reviewer accounts"})
		}
		if input.FacilityID == nil {
			errs = append(errs, utils.ErrorDetail{Field: "facility_id", Message: "Facility is required"})
		} else if actorDistrict != "" {
			ok, ferr := s.facilityRepo.FacilityBelongsToDistrict(*input.FacilityID, actorDistrict)
			if ferr != nil || !ok {
				errs = append(errs, utils.ErrorDetail{Field: "facility_id", Message: "Facility must belong to your district"})
			}
		}
		input.District = ""
	}

	// Only superadmin may create another superadmin or district biostatistician
	if actorRole != models.RoleSuperAdmin {
		if role == models.RoleSuperAdmin || role == models.RoleDistrictBiostatistician {
			errs = append(errs, utils.ErrorDetail{Field: "role", Message: "Only superadmin can assign this role"})
		}
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
	if role == models.RoleDistrictBiostatistician {
		user.District = strings.TrimSpace(input.District)
		user.FacilityID = nil
	} else {
		user.District = ""
	}

	if err := s.repo.Create(&user); err != nil {
		return nil, []utils.ErrorDetail{{Message: "Failed to create user"}}
	}

	s.auditSvc.Log(&actorID, user.FacilityID, models.AuditCreate,
		"user", user.ID.String(), ip, ua, "Created user: "+user.Email)

	return &user, nil
}

func (s *UserService) GetByID(id uuid.UUID) (*models.User, error) {
	return s.repo.FindByID(id)
}

func (s *UserService) List(page, perPage int, facilityID *uuid.UUID, districtScope string) ([]models.User, int64, error) {
	return s.repo.List(page, perPage, facilityID, districtScope)
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

	if !s.CanActorAccessUser(actorID, id) {
		return nil, errors.New("forbidden")
	}

	if user.Role == models.RoleSuperAdmin && actorRole != models.RoleSuperAdmin {
		return nil, errors.New("only superadmin can modify a superadmin user")
	}

	if strings.TrimSpace(input.Email) != "" && input.Email != user.Email {
		if !utils.IsValidEmail(input.Email) {
			return nil, errors.New("valid email is required")
		}
		if existing, _ := s.repo.FindByEmail(input.Email); existing != nil && existing.ID != uuid.Nil && existing.ID != user.ID {
			return nil, errors.New("email already in use")
		}
		user.Email = strings.TrimSpace(input.Email)
	}

	if strings.TrimSpace(input.FullName) != "" {
		user.FullName = strings.TrimSpace(input.FullName)
	}

	newRole := user.Role
	if input.Role != "" {
		newRole = models.Role(input.Role)
	}

	// District biostatistician updating own profile: email / name / password only
	if actor != nil && actor.ID == user.ID && actorRole == models.RoleDistrictBiostatistician {
		if input.Role != "" && newRole != user.Role {
			return nil, errors.New("cannot change your own role")
		}
		if input.FacilityID != nil {
			return nil, errors.New("cannot set facility on a district biostatistician account")
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

	switch actorRole {
	case models.RoleFacilityAdmin:
		if newRole == models.RoleSuperAdmin || newRole == models.RoleFacilityAdmin || newRole == models.RoleDistrictBiostatistician {
			return nil, errors.New("facility admin can only assign Facility User or Reviewer roles")
		}
		if actorFacilityID == nil {
			return nil, errors.New("your account is not linked to a facility")
		}
		if user.FacilityID == nil || *user.FacilityID != *actorFacilityID {
			return nil, errors.New("you can only manage users in your facility")
		}
		input.FacilityID = actorFacilityID
		user.Role = newRole
		user.FacilityID = input.FacilityID
		user.District = ""

	case models.RoleDistrictBiostatistician:
		ad := strings.TrimSpace(actor.District)
		if ad == "" {
			return nil, errors.New("your account has no district assigned")
		}
		if user.Role == models.RoleDistrictBiostatistician && user.ID != actor.ID {
			return nil, errors.New("cannot modify another district biostatistician")
		}
		if newRole != models.RoleFacilityUser && newRole != models.RoleReviewer {
			return nil, errors.New("you can only assign Facility User or Reviewer roles")
		}
		if input.FacilityID == nil {
			return nil, errors.New("facility is required")
		}
		ok, ferr := s.facilityRepo.FacilityBelongsToDistrict(*input.FacilityID, ad)
		if ferr != nil || !ok {
			return nil, errors.New("facility must belong to your district")
		}
		user.Role = newRole
		user.FacilityID = input.FacilityID
		user.District = ""

	case models.RoleSuperAdmin:
		user.Role = newRole
		switch newRole {
		case models.RoleSuperAdmin:
			user.FacilityID = nil
			user.District = ""
		case models.RoleDistrictBiostatistician:
			d := strings.TrimSpace(input.District)
			if d == "" {
				return nil, errors.New("district is required for district biostatistician")
			}
			user.District = d
			user.FacilityID = nil
		default:
			user.District = ""
			if input.FacilityID != nil {
				user.FacilityID = input.FacilityID
			}
			if user.FacilityID == nil {
				return nil, errors.New("facility is required for this role")
			}
		}

	default:
		return nil, errors.New("forbidden")
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
	actor, _ := s.repo.FindByID(actorID)
	if actor == nil {
		return errors.New("forbidden")
	}

	switch actor.Role {
	case models.RoleFacilityAdmin:
		if actor.FacilityID == nil || user.FacilityID == nil || *user.FacilityID != *actor.FacilityID {
			return errors.New("you can only deactivate users in your facility")
		}
	case models.RoleDistrictBiostatistician:
		if user.Role == models.RoleSuperAdmin || user.Role == models.RoleFacilityAdmin || user.Role == models.RoleDistrictBiostatistician {
			return errors.New("cannot deactivate this user")
		}
		if !s.userBelongsToDistrict(user, actor.District) {
			return errors.New("forbidden")
		}
	case models.RoleSuperAdmin:
		// ok
	default:
		return errors.New("forbidden")
	}

	user.IsActive = false
	if err := s.repo.Update(user); err != nil {
		return err
	}
	s.auditSvc.Log(&actorID, user.FacilityID, models.AuditAdminAction,
		"user", user.ID.String(), ip, ua, "Deactivated user: "+user.Email)
	return nil
}
