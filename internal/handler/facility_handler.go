package handler

import (
	"net/http"

	"fpreg/internal/middleware"
	"fpreg/internal/models"
	"fpreg/internal/service"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FacilityHandler struct {
	facilitySvc *service.FacilityService
}

func NewFacilityHandler(facilitySvc *service.FacilityService) *FacilityHandler {
	return &FacilityHandler{facilitySvc: facilitySvc}
}

// CreateFacility godoc
// @Summary      Create a facility
// @Tags         facilities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body service.CreateFacilityInput true "Facility data"
// @Success      201  {object} utils.APIResponse
// @Router       /api/v1/facilities [post]
func (h *FacilityHandler) Create(c *gin.Context) {
	var input service.CreateFacilityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	actorID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	f, errs := h.facilitySvc.Create(input, actorID, ip, ua)
	if errs != nil {
		utils.RespondValidationError(c, errs)
		return
	}
	utils.RespondCreated(c, f)
}

// ListFacilities godoc
// @Summary      List facilities
// @Tags         facilities
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page"
// @Param        per_page query int false "Per page"
// @Param        search query string false "Search by name, code, district, or subcounty"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/facilities [get]
func (h *FacilityHandler) List(c *gin.Context) {
	// Allow large per_page for admin/dropdown use (default remains 25).
	page, perPage := utils.GetPaginationOrMax(c, 25, 10000)
	search := c.Query("search")

	var items []models.Facility
	var total int64
	var err error

	roleVal, _ := c.Get("user_role")
	role, _ := roleVal.(models.Role)
	if role == models.RoleDistrictBiostatistician {
		d := middleware.GetUserDistrict(c)
		if d == "" {
			utils.RespondError(c, http.StatusForbidden, "No district assigned to your account")
			return
		}
		items, total, err = h.facilitySvc.ListByDistrict(page, perPage, d, search)
	} else {
		items, total, err = h.facilitySvc.List(page, perPage, search)
	}
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list facilities")
		return
	}
	utils.RespondPaginated(c, items, utils.Meta{
		Page: page, PerPage: perPage, Total: total,
		TotalPages: utils.CalcTotalPages(total, perPage),
	})
}

// ListDistricts godoc
// @Summary      List distinct facility districts
// @Description  District values from the facilities table (for assigning district biostatisticians).
// @Tags         facilities
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/facilities/districts [get]
func (h *FacilityHandler) ListDistricts(c *gin.Context) {
	names, err := h.facilitySvc.ListDistinctDistricts()
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list districts")
		return
	}
	utils.RespondOK(c, names)
}

// GetFacility godoc
// @Summary      Get facility by ID
// @Tags         facilities
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Facility ID"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/facilities/{id} [get]
func (h *FacilityHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}
	f, err := h.facilitySvc.GetByID(id)
	if err != nil {
		utils.RespondNotFound(c, "Facility")
		return
	}
	roleVal, _ := c.Get("user_role")
	role, _ := roleVal.(models.Role)
	if role == models.RoleDistrictBiostatistician {
		d := middleware.GetUserDistrict(c)
		if d == "" {
			utils.RespondError(c, http.StatusForbidden, "No district assigned to your account")
			return
		}
		ok, err := h.facilitySvc.FacilityBelongsToDistrict(id, d)
		if err != nil || !ok {
			utils.RespondNotFound(c, "Facility")
			return
		}
	}
	utils.RespondOK(c, f)
}

// UpdateFacility godoc
// @Summary      Update facility
// @Tags         facilities
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Facility ID"
// @Param        body body service.CreateFacilityInput true "Facility data"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/facilities/{id} [put]
func (h *FacilityHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}
	var input service.CreateFacilityInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid request body")
		return
	}
	actorID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	f, err := h.facilitySvc.Update(id, input, actorID, ip, ua)
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondOK(c, f)
}

// DeleteFacility godoc
// @Summary      Delete facility
// @Tags         facilities
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Facility ID"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/facilities/{id} [delete]
func (h *FacilityHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, "Invalid ID")
		return
	}
	actorID := middleware.GetUserID(c)
	ip := utils.GetClientIP(c)
	ua := c.GetHeader("User-Agent")

	if err := h.facilitySvc.Delete(id, actorID, ip, ua); err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondMessage(c, "Facility deleted")
}
