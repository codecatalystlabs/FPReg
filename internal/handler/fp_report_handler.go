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

	var facilityIDs []uuid.UUID
	if fid := c.Query("facility_id"); fid != "" {
		if id, err := uuid.Parse(fid); err == nil {
			facilityIDs = append(facilityIDs, id)
		}
	}

	// Scope facilities for non-superadmin
	if scoped := middleware.GetScopedFacilityID(c); scoped != nil {
		// Non-superadmin are already scoped to a single facility
		facilityIDs = []uuid.UUID{*scoped}
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

	out := struct {
		Period     string           `json:"period"`
		Facilities []facilityReport `json:"facilities"`
	}{
		Period:     period,
		Facilities: []facilityReport{},
	}
	for _, fr := range facMap {
		out.Facilities = append(out.Facilities, *fr)
	}

	utils.RespondOK(c, out)
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

	var facilityIDs []uuid.UUID
	if fid := c.Query("facility_id"); fid != "" {
		if id, err := uuid.Parse(fid); err == nil {
			facilityIDs = append(facilityIDs, id)
		}
	}
	if scoped := middleware.GetScopedFacilityID(c); scoped != nil {
		facilityIDs = []uuid.UUID{*scoped}
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
// @Param        body body struct{Period string `+"`json:\"period\"`"+`; FacilityIDs []string `+"`json:\"facility_ids\"`"+`; Force bool `+"`json:\"force\"`"+`} true "Sync payload"
// @Success      200  {object} utils.APIResponse
// @Router       /api/v1/reports/family-planning/sync [post]
func (h *FPReportHandler) Sync(c *gin.Context) {
	var req struct {
		Period      string   `json:"period"`
		FacilityIDs []string `json:"facility_ids"`
		Force       bool     `json:"force"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || len(req.Period) != 6 {
		utils.RespondError(c, http.StatusBadRequest, "Invalid body or period")
		return
	}

	var facilityIDs []uuid.UUID
	for _, s := range req.FacilityIDs {
		if id, err := uuid.Parse(s); err == nil {
			facilityIDs = append(facilityIDs, id)
		}
	}
	if scoped := middleware.GetScopedFacilityID(c); scoped != nil {
		// Non-superadmin can only sync their own facility
		facilityIDs = []uuid.UUID{*scoped}
	}

	userID := middleware.GetUserID(c)
	logs, err := h.syncSvc.Sync(req.Period, facilityIDs, req.Force, userID.String())
	if err != nil {
		utils.RespondError(c, http.StatusBadRequest, err.Error())
		return
	}
	utils.RespondOK(c, logs)
}
