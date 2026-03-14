package handler

import (
	"net/http"

	"fpreg/internal/middleware"
	"fpreg/internal/repository"
	"fpreg/internal/service"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegistrationHandler struct {
	regSvc *service.RegistrationService
}

func NewRegistrationHandler(regSvc *service.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{regSvc: regSvc}
}

// CreateRegistration godoc
// @Summary      Create a new FP register entry
// @Tags         registrations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body service.CreateRegistrationInput true "Registration data"
// @Success      201  {object} utils.APIResponse
// @Failure      422  {object} utils.APIError
// @Router       /api/v1/registrations [post]
func (h *RegistrationHandler) Create(c *gin.Context) {
	var input service.CreateRegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := middleware.GetUserID(c)
	facilityID := middleware.GetScopedFacilityID(c)
	if facilityID == nil {
		utils.RespondError(c, http.StatusBadRequest, "No facility associated with your account")
		return
	}
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	reg, errs := h.regSvc.Create(input, *facilityID, userID, ip, ua)
	if errs != nil {
		utils.RespondValidationError(c, errs)
		return
	}
	utils.RespondCreated(c, reg)
}

// ListRegistrations godoc
// @Summary      List FP register entries
// @Tags         registrations
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page"
// @Param        per_page query int false "Per page"
// @Param        visit_date query string false "Filter by visit date (YYYY-MM-DD)"
// @Param        search query string false "Search by name, NIN, or client number"
// @Param        sex query string false "Filter by sex (M/F)"
// @Param        date_from query string false "From date"
// @Param        date_to query string false "To date"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/registrations [get]
func (h *RegistrationHandler) List(c *gin.Context) {
	page, perPage := utils.GetPagination(c)

	f := repository.RegistrationFilter{
		FacilityID: middleware.GetScopedFacilityID(c),
		VisitDate:  c.Query("visit_date"),
		Search:     c.Query("search"),
		Sex:        c.Query("sex"),
		DateFrom:   c.Query("date_from"),
		DateTo:     c.Query("date_to"),
	}

	items, total, err := h.regSvc.List(page, perPage, f)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list registrations")
		return
	}

	utils.RespondPaginated(c, items, utils.Meta{
		Page: page, PerPage: perPage, Total: total,
		TotalPages: utils.CalcTotalPages(total, perPage),
	})
}

// GetRegistration godoc
// @Summary      Get a single FP register entry
// @Tags         registrations
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Registration ID"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/registrations/{id} [get]
func (h *RegistrationHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	reg, err := h.regSvc.GetByID(id)
	if err != nil {
		utils.RespondNotFound(c, "Registration")
		return
	}

	facilityID := middleware.GetScopedFacilityID(c)
	if facilityID != nil && reg.FacilityID != *facilityID {
		utils.RespondForbidden(c)
		return
	}

	utils.RespondOK(c, reg)
}

// UpdateRegistration godoc
// @Summary      Update an FP register entry
// @Tags         registrations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Registration ID"
// @Param        body body service.CreateRegistrationInput true "Updated data"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/registrations/{id} [put]
func (h *RegistrationHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	var input service.CreateRegistrationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	userID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	reg, errs := h.regSvc.Update(id, input, userID, ip, ua)
	if errs != nil {
		utils.RespondValidationError(c, errs)
		return
	}
	utils.RespondOK(c, reg)
}

// DeleteRegistration godoc
// @Summary      Soft-delete an FP register entry
// @Tags         registrations
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Registration ID"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/registrations/{id} [delete]
func (h *RegistrationHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}

	userID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	if err := h.regSvc.Delete(id, userID, ip, ua); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondMessage(c, "Registration deleted")
}
