package handler

import (
	"errors"
	"net/http"
	"sort"

	"fpreg/internal/middleware"
	"fpreg/internal/models"
	"fpreg/internal/repository"
	"fpreg/internal/service"
	"fpreg/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type FPReportSyncRequest struct {
	Period      string   `json:"period"`
	FacilityIDs []string `json:"facility_ids"`
	Force       bool     `json:"force"`
}

type FPReportHandler struct {
	reportSvc    *service.FPReportService
	facilityRepo *repository.FacilityRepository
	dhisRepo     *repository.DHIS2Repository
	syncSvc      *service.DHIS2SyncService
}

func NewFPReportHandler(
	reportSvc *service.FPReportService,
	facilityRepo *repository.FacilityRepository,
	dhisRepo *repository.DHIS2Repository,
	syncSvc *service.DHIS2SyncService,
) *FPReportHandler {
	return &FPReportHandler{
		reportSvc:    reportSvc,
		facilityRepo: facilityRepo,
		dhisRepo:     dhisRepo,
		syncSvc:      syncSvc,
	}
}

func (h *FPReportHandler) resolveReportFacilityIDs(c *gin.Context, explicit []uuid.UUID) ([]uuid.UUID, error) {
	roleVal, _ := c.Get("user_role")
	role, _ := roleVal.(models.Role)

	if role == models.RoleSuperAdmin {
		if len(explicit) > 0 {
			return explicit, nil
		}
		return []uuid.UUID{}, nil
	}

	if scoped := middleware.GetScopedFacilityID(c); scoped != nil {
		return []uuid.UUID{*scoped}, nil
	}

	if role == models.RoleDistrictBiostatistician {
		d := middleware.GetUserDistrict(c)
		if d == "" {
			return nil, errors.New("no district assigned to your account")
		}
		if len(explicit) > 0 {
			for _, id := range explicit {
				ok, err := h.facilityRepo.FacilityBelongsToDistrict(id, d)
				if err != nil || !ok {
					return nil, errors.New("one or more facilities are not in your district")
				}
			}
			return explicit, nil
		}
		return h.facilityRepo.FindIDsByDistrict(d)
	}

	return nil, errors.New("forbidden")
}

// MonthlyReport godoc
// @Summary      Monthly FP methods report
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        period query string true "Period in YYYYMM format"
// @Param        facility_id query string false "Facility ID"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/reports/family-planning/monthly [get]
func (h *FPReportHandler) Monthly(c *gin.Context) {
	period := c.Query("period")
	if len(period) != 6 {
		utils.RespondError(c, http.StatusBadRequest, "period must be in YYYYMM format")
		return
	}

	var explicit []uuid.UUID
	if fid := c.Query("facility_id"); fid != "" {
		if id, err := uuid.Parse(fid); err == nil {
			explicit = append(explicit, id)
		}
	}
	facilityIDs, err := h.resolveReportFacilityIDs(c, explicit)
	if err != nil {
		utils.RespondError(c, http.StatusForbidden, err.Error())
		return
	}

	rows, err := h.reportSvc.AggregateForPeriod(period, facilityIDs)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, "Failed to aggregate report")
		return
	}

	type cell struct {
		LocalKey  string `json:"local_indicator_key"`
		Method    string `json:"method_code"`
		VisitType string `json:"visit_type"`
		AgeGroup  string `json:"age_group"`
		Value     int    `json:"value"`
	}
	type facilityReport struct {
		FacilityID   uuid.UUID `json:"facility_id"`
		FacilityName string    `json:"facility_name"`
		OrgUnitUID   string    `json:"orgunit_uid"`
		Cells        []cell    `json:"cells"`
	}

	facMap := map[uuid.UUID]*facilityReport{}
	for _, r := range rows {
		fr, ok := facMap[r.FacilityID]
		if !ok {
			f, _ := h.facilityRepo.FindByID(r.FacilityID)
			name := ""
			org := ""
			if f != nil {
				name = f.Name
				org = f.UID
			}
			fr = &facilityReport{
				FacilityID:   r.FacilityID,
				FacilityName: name,
				OrgUnitUID:   org,
				Cells:        []cell{},
			}
			facMap[r.FacilityID] = fr
		}

		method, visitType, ageGroup := service.ParseLocalIndicatorKey(r.LocalIndicatorKey)
		fr.Cells = append(fr.Cells, cell{
			LocalKey:  r.LocalIndicatorKey,
			Method:    method,
			VisitType: visitType,
			AgeGroup:  ageGroup,
			Value:     r.Value,
		})
	}

	list := make([]facilityReport, 0, len(facMap))
	for _, fr := range facMap {
		list = append(list, *fr)
	}
	// Stable, useful order: most non-zero activity first (UI used to pick random map key).
	sort.Slice(list, func(i, j int) bool {
		ti, tj := 0, 0
		for _, c := range list[i].Cells {
			ti += c.Value
		}
		for _, c := range list[j].Cells {
			tj += c.Value
		}
		if ti != tj {
			return ti > tj
		}
		return list[i].FacilityName < list[j].FacilityName
	})

	out := struct {
		Period     string           `json:"period"`
		Facilities []facilityReport `json:"facilities"`
	}{
		Period:     period,
		Facilities: list,
	}

	utils.RespondOK(c, out)
}

// PayloadPreview godoc
// @Summary      Preview DHIS2 payload for monthly FP report
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        period query string true "Period in YYYYMM format"
// @Param        facility_id query string false "Facility ID"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/reports/family-planning/payload-preview [get]
func (h *FPReportHandler) PayloadPreview(c *gin.Context) {
	period := c.Query("period")
	if len(period) != 6 {
		utils.RespondError(c, http.StatusBadRequest, "period must be in YYYYMM format")
		return
	}

	var explicit []uuid.UUID
	if fid := c.Query("facility_id"); fid != "" {
		if id, err := uuid.Parse(fid); err == nil {
			explicit = append(explicit, id)
		}
	}
	facilityIDs, err := h.resolveReportFacilityIDs(c, explicit)
	if err != nil {
		utils.RespondError(c, http.StatusForbidden, err.Error())
		return
	}

	previews, err := h.syncSvc.BuildPreview(period, facilityIDs)
	if err != nil {
		utils.RespondError(c, http.StatusInternalServerError, err.Error())
		return
	}
	utils.RespondOK(c, previews)
}

// Sync godoc
// @Summary      Sync monthly FP report to DHIS2
// @Tags         reports
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body body handler.FPReportSyncRequest true "Sync payload"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/reports/family-planning/sync [post]
func (h *FPReportHandler) Sync(c *gin.Context) {
	var req FPReportSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Period) != 6 {
		utils.RespondError(c, http.StatusBadRequest, "Invalid body or period")
		return
	}

	var explicit []uuid.UUID
	for _, s := range req.FacilityIDs {
		if id, err := uuid.Parse(s); err == nil {
			explicit = append(explicit, id)
		}
	}
	facilityIDs, err := h.resolveReportFacilityIDs(c, explicit)
	if err != nil {
		utils.RespondError(c, http.StatusForbidden, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	outcome, err := h.syncSvc.Sync(req.Period, facilityIDs, req.Force, userID.String())
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondOK(c, outcome)
}
