package handler

import (
	"encoding/csv"
	"net/http"
	"strings"

	"fpreg/internal/middleware"
	"fpreg/internal/models"
	"fpreg/internal/service"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

// CreateUser godoc
// @Summary      Create a user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body service.CreateUserInput true "User data"
// @Success      201  {object} utils.APIResponse
// @Failure      422  {object} utils.APIError
// @Router       /api/v1/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var input service.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	actorID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	user, errs := h.userSvc.Create(input, actorID, ip, ua)
	if errs != nil {
		utils.RespondValidationError(c, errs)
		return
	}
	utils.RespondCreated(c, user)
}

// ListUsers godoc
// @Summary      List users
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page"
// @Param        per_page query int false "Per page"
// @Param        facility_id query string false "Filter by facility"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/users [get]
func (h *UserHandler) List(c *gin.Context) {
	page, perPage := utils.GetPagination(c)

	var facilityID *uuid.UUID
	if fid := c.Query("facility_id"); fid != "" {
		id, err := uuid.Parse(fid)
		if err == nil {
			facilityID = &id
		}
	}

	scopedFID := middleware.GetScopedFacilityID(c)
	if scopedFID != nil {
		facilityID = scopedFID
	}

	users, total, err := h.userSvc.List(page, perPage, facilityID)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list users")
		return
	}

	utils.RespondPaginated(c, users, utils.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: utils.CalcTotalPages(total, perPage),
	})
}

// GetUser godoc
// @Summary      Get user by ID
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/users/{id} [get]
func (h *UserHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}
	user, err := h.userSvc.GetByID(id)
	if err != nil {
		utils.RespondNotFound(c, "User")
		return
	}
	utils.RespondOK(c, user)
}

// UpdateUser godoc
// @Summary      Update user
// @Tags         users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID"
// @Param        body body service.CreateUserInput true "User data"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}
	var input service.CreateUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	actorID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	user, err := h.userSvc.Update(id, input, actorID, ip, ua)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondOK(c, user)
}

// ImportUsers godoc
// @Summary      Bulk import users from CSV
// @Tags         users
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "CSV file with columns: full_name,email,password,role,facility_id"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/users/import [post]
func (h *UserHandler) Import(c *gin.Context) {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil || len(rows) < 2 {
		utils.RespondError(c, http.StatusBadRequest, "Invalid or empty CSV file")
		return
	}

	header := rows[0]
	fullNameIdx, emailIdx, passwordIdx, roleIdx, facilityIDIdx := -1, -1, -1, -1, -1
	for i, h := range header {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "full_name":
			fullNameIdx = i
		case "email":
			emailIdx = i
		case "password":
			passwordIdx = i
		case "role":
			roleIdx = i
		case "facility_id":
			facilityIDIdx = i
		}
	}
	if fullNameIdx < 0 || emailIdx < 0 || passwordIdx < 0 || roleIdx < 0 {
		utils.RespondError(c, http.StatusBadRequest, "CSV must have columns: full_name,email,password,role[,facility_id]")
		return
	}

	actorID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")
	roleVal, _ := c.Get("user_role")
	actorRole, _ := roleVal.(models.Role)
	actorFacilityID := middleware.GetFacilityID(c)

	imported := 0
	failed := 0

	for i, row := range rows[1:] {
		if len(row) <= emailIdx {
			failed++
			continue
		}
		input := service.CreateUserInput{
			Email:    row[emailIdx],
			Password: row[passwordIdx],
			FullName: row[fullNameIdx],
			Role:     row[roleIdx],
		}

		// Facility: for superadmin we accept facility_id from CSV; for facility_admin we force their own facility
		if actorRole == models.RoleFacilityAdmin {
			input.FacilityID = actorFacilityID
		} else if facilityIDIdx >= 0 && facilityIDIdx < len(row) && row[facilityIDIdx] != "" {
			if fid, err := uuid.Parse(row[facilityIDIdx]); err == nil {
				input.FacilityID = &fid
			}
		}

		if _, errs := h.userSvc.Create(input, actorID, ip, ua); errs != nil {
			failed++
			_ = i // row index reserved for future detailed reporting
			continue
		}
		imported++
	}

	resp := gin.H{"imported": imported, "failed": failed}
	utils.RespondOK(c, resp)
}

// DeactivateUser godoc
// @Summary      Deactivate user
// @Tags         users
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/users/{id}/deactivate [patch]
func (h *UserHandler) Deactivate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	actorID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	if err := h.userSvc.Deactivate(id, actorID, ip, ua); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondMessage(c, "User deactivated")
}
