package handler

import (
	"net/http"

	"fpreg/internal/middleware"
	"fpreg/internal/models"
	"fpreg/internal/repository"
	"fpreg/internal/service"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuditHandler struct {
	auditSvc     *service.AuditService
	facilityRepo *repository.FacilityRepository
	userSvc      *service.UserService
}

func NewAuditHandler(auditSvc *service.AuditService, facilityRepo *repository.FacilityRepository, userSvc *service.UserService) *AuditHandler {
	return &AuditHandler{auditSvc: auditSvc, facilityRepo: facilityRepo, userSvc: userSvc}
}

// ListAuditLogs godoc
// @Summary      List audit logs (admin only)
// @Tags         audit
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page"
// @Param        per_page query int false "Per page"
// @Param        user_id query string false "Filter by user ID"
// @Param        action query string false "Filter by action"
// @Param        entity query string false "Filter by entity"
// @Param        date_from query string false "From date (YYYY-MM-DD)"
// @Param        date_to query string false "To date (YYYY-MM-DD)"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/audit-logs [get]
func (h *AuditHandler) List(c *gin.Context) {
	page, perPage := utils.GetPagination(c)

	f := repository.AuditFilter{
		Action:   c.Query("action"),
		Entity:   c.Query("entity"),
		DateFrom: c.Query("date_from"),
		DateTo:   c.Query("date_to"),
	}
	roleVal, _ := c.Get("user_role")
	actorRole, _ := roleVal.(models.Role)
	if actorRole == models.RoleDistrictBiostatistician {
		d := middleware.GetUserDistrict(c)
		if d == "" {
			utils.RespondError(c, http.StatusForbidden, "No district assigned to your account")
			return
		}
		if uid := c.Query("user_id"); uid != "" {
			id, err := uuid.Parse(uid)
			if err != nil {
				utils.RespondError(c, http.StatusBadRequest, "Invalid user_id")
				return
			}
			if !h.userSvc.CanActorAccessUser(middleware.GetUserID(c), id) {
				utils.RespondForbidden(c)
				return
			}
			f.UserID = &id
		}
		if fid := c.Query("facility_id"); fid != "" {
			id, err := uuid.Parse(fid)
			if err != nil {
				utils.RespondError(c, http.StatusBadRequest, "Invalid facility_id")
				return
			}
			ok, ferr := h.facilityRepo.FacilityBelongsToDistrict(id, d)
			if ferr != nil || !ok {
				utils.RespondForbidden(c)
				return
			}
			f.FacilityID = &id
			f.DistrictScope = ""
		} else if f.UserID == nil {
			f.DistrictScope = d
		}
	} else {
		if uid := c.Query("user_id"); uid != "" {
			id, err := uuid.Parse(uid)
			if err == nil {
				f.UserID = &id
			}
		}
		if fid := c.Query("facility_id"); fid != "" {
			id, err := uuid.Parse(fid)
			if err == nil {
				f.FacilityID = &id
			}
		}
	}

	logs, total, err := h.auditSvc.List(page, perPage, f)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to list audit logs")
		return
	}

	utils.RespondPaginated(c, logs, utils.Meta{
		Page: page, PerPage: perPage, Total: total,
		TotalPages: utils.CalcTotalPages(total, perPage),
	})
}
