package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"fpreg/internal/config"
	"fpreg/internal/models"
	"fpreg/internal/repository"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DHIS2SyncService struct {
	cfg        *config.Config
	fpReport   *FPReportService
	facilities *repository.FacilityRepository
	dhisRepo   *repository.DHIS2Repository
	httpClient *http.Client
}

type FPCellWithMapping struct {
	Period            string                   `json:"period"`
	FacilityID        uuid.UUID                `json:"facility_id"`
	OrgUnitUID        string                   `json:"orgunit_uid"`
	LocalIndicatorKey string                   `json:"local_indicator_key"`
	Value             int                      `json:"value"`
	Mapping           *models.DHIS2MappingItem `json:"mapping,omitempty"`
	ExcludedReason    string                   `json:"excluded_reason,omitempty"`
}

type dhis2DataValue struct {
	DataElement          string `json:"dataElement"`
	Period               string `json:"period"`
	OrgUnit              string `json:"orgUnit"`
	CategoryOptionCombo  string `json:"categoryOptionCombo"`
	AttributeOptionCombo string `json:"attributeOptionCombo"`
	Value                int    `json:"value"`
}

type dhis2Payload struct {
	DataValues []dhis2DataValue `json:"dataValues"`
}

// MissingMappingWithValue is an aggregated cell that has no usable DHIS2 mapping row yet.
type MissingMappingWithValue struct {
	LocalIndicatorKey string `json:"local_indicator_key"`
	Value             int    `json:"value"`
}

type SyncPreview struct {
	Period                   string                    `json:"period"`
	OrgUnitUID               string                    `json:"orgunit_uid"`
	Cells                    []FPCellWithMapping       `json:"cells"`
	DataValues               []dhis2DataValue          `json:"data_values"`
	MissingMappings          []string                  `json:"missing_mappings"`
	MissingMappingWithValues []MissingMappingWithValue `json:"missing_mapping_with_values,omitempty"`
	MissingOrgUnits          []string                  `json:"missing_org_units"`
}

// DHIS2SyncOutcome is returned from Sync for UI feedback when no HTTP batches ran.
type DHIS2SyncOutcome struct {
	Logs                     []models.ReportSyncLog    `json:"logs"`
	PreviewFacilityCount     int                       `json:"preview_facility_count"`
	TotalDataValues          int                       `json:"total_data_values"`
	PostedBatches            int                       `json:"posted_batches"`
	SkippedNoDataValues      int                       `json:"skipped_no_data_values"`
	SkippedAlreadySynced     int                       `json:"skipped_already_synced"`
	DetailMessages           []string                  `json:"detail_messages"`
	MissingMappingWithValues []MissingMappingWithValue `json:"missing_mapping_with_values,omitempty"`
}

func NewDHIS2SyncService(
	cfg *config.Config,
	fpReport *FPReportService,
	facilityRepo *repository.FacilityRepository,
	dhisRepo *repository.DHIS2Repository,
) *DHIS2SyncService {
	return &DHIS2SyncService{
		cfg:        cfg,
		fpReport:   fpReport,
		facilities: facilityRepo,
		dhisRepo:   dhisRepo,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// BuildPreview builds aggregated cells + DHIS2 dataValues for facilities/period.
// facilityIDs==nil or empty => all facilities.
func (s *DHIS2SyncService) BuildPreview(period string, facilityIDs []uuid.UUID) ([]SyncPreview, error) {
	rows, err := s.fpReport.AggregateForPeriod(period, facilityIDs)
	if err != nil {
		return nil, err
	}

	// Load active settings
	settings, err := s.dhisRepo.GetActiveSettings()
	if err != nil {
		return nil, fmt.Errorf("no active DHIS2 settings: %w", err)
	}
	attrOC := settings.AttributeOptionCombo

	// Group rows by facility
	type key struct {
		FacilityID uuid.UUID
	}
	grouped := map[uuid.UUID][]FPReportAggregationRow{}
	for _, r := range rows {
		grouped[r.FacilityID] = append(grouped[r.FacilityID], r)
	}

	var previews []SyncPreview

	for facID, facRows := range grouped {
		fac, err := s.facilities.FindByID(facID)
		orgUnit := ""
		if om, e := s.dhisRepo.FindOrgUnitMappingByFacilityID(facID); e == nil && om != nil {
			orgUnit = strings.TrimSpace(om.DHIS2OrgUnitUID)
		} else if e != nil && !errors.Is(e, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("org unit mapping for facility %s: %w", facID, e)
		}
		if orgUnit == "" && fac != nil {
			orgUnit = strings.TrimSpace(fac.UID)
		}
		if err != nil || fac == nil || orgUnit == "" {
			// all rows for this facility missing orgUnit
			for _, r := range facRows {
				_ = s.dhisRepo.CreateExclusion(&models.AggregationExclusionLog{
					RawRecordID: nil,
					Reason:      "MISSING_ORGUNIT_UID",
					Period:      period,
					FacilityID:  &facID,
					Details:     fmt.Sprintf(`{"local_indicator_key":"%s"}`, r.LocalIndicatorKey),
				})
			}
			continue
		}

		preview := SyncPreview{
			Period:     period,
			OrgUnitUID: orgUnit,
			Cells:      []FPCellWithMapping{},
		}
		var dataValues []dhis2DataValue
		missingMappings := map[string]bool{}

		for _, r := range facRows {
			cell := FPCellWithMapping{
				Period:            period,
				FacilityID:        facID,
				OrgUnitUID:        orgUnit,
				LocalIndicatorKey: r.LocalIndicatorKey,
				Value:             r.Value,
			}

			mapping, mErr := s.dhisRepo.FindMappingByLocalKey(r.LocalIndicatorKey)
			if mErr != nil {
				if !errors.Is(mErr, gorm.ErrRecordNotFound) {
					return nil, mErr
				}
				cell.ExcludedReason = "MISSING_MAPPING"
				missingMappings[r.LocalIndicatorKey] = true
				preview.MissingMappingWithValues = append(preview.MissingMappingWithValues, MissingMappingWithValue{
					LocalIndicatorKey: r.LocalIndicatorKey,
					Value:             r.Value,
				})
				preview.Cells = append(preview.Cells, cell)
				continue
			}
			cell.Mapping = mapping
			preview.Cells = append(preview.Cells, cell)

			if r.Value == 0 {
				// For now: skip zero values; can make this configurable later.
				continue
			}

			dv := dhis2DataValue{
				DataElement:          mapping.DHIS2DataElementUID,
				Period:               period,
				OrgUnit:              orgUnit,
				CategoryOptionCombo:  mapping.DHIS2CatOptionComboUID,
				AttributeOptionCombo: attrOC,
				Value:                r.Value,
			}
			dataValues = append(dataValues, dv)
		}

		for k := range missingMappings {
			preview.MissingMappings = append(preview.MissingMappings, k)
		}
		preview.DataValues = dataValues
		previews = append(previews, preview)
	}

	return previews, nil
}

// Sync posts the payloads to DHIS2 and writes logs + cell status.
// If force is false, already-synced cells with same checksum are skipped.
func (s *DHIS2SyncService) Sync(period string, facilityIDs []uuid.UUID, force bool, initiatedBy string) (DHIS2SyncOutcome, error) {
	out := DHIS2SyncOutcome{
		Logs:           []models.ReportSyncLog{},
		DetailMessages: []string{},
	}
	previews, err := s.BuildPreview(period, facilityIDs)
	if err != nil {
		return out, err
	}
	settings, err := s.dhisRepo.GetActiveSettings()
	if err != nil {
		return out, fmt.Errorf("no active DHIS2 settings: %w", err)
	}

	baseURL := settings.DHIS2BaseURL
	if baseURL == "" {
		return out, fmt.Errorf("dhis2 base url not configured")
	}

	out.PreviewFacilityCount = len(previews)
	if len(previews) == 0 {
		out.DetailMessages = append(out.DetailMessages,
			"No DHIS2 payloads were built: no aggregated registrations for this period in scope, or every facility is missing a resolvable DHIS2 org unit (facility uid / org_unit_mappings).")
		return out, nil
	}

	mergedMissing := map[string]int{}
	for _, p := range previews {
		for _, d := range p.MissingMappingWithValues {
			mergedMissing[d.LocalIndicatorKey] += d.Value
		}
	}
	if len(mergedMissing) > 0 {
		keys := make([]string, 0, len(mergedMissing))
		for k := range mergedMissing {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			out.MissingMappingWithValues = append(out.MissingMappingWithValues, MissingMappingWithValue{
				LocalIndicatorKey: k,
				Value:             mergedMissing[k],
			})
		}
		out.DetailMessages = append(out.DetailMessages,
			"Register aggregates exist, but DHIS2 cannot post until each local_indicator_key below has dhis2_data_element_uid and dhis2_cat_option_combo_uid set in dhis2_mapping_item (stub rows are created on first app migrate).")
	}

	for _, p := range previews {
		out.TotalDataValues += len(p.DataValues)
		if len(p.DataValues) == 0 {
			out.SkippedNoDataValues++
			nMiss := len(p.MissingMappings)
			if nMiss > 0 && len(p.MissingMappingWithValues) > 0 {
				var parts []string
				for i, d := range p.MissingMappingWithValues {
					if i >= 5 {
						parts = append(parts, fmt.Sprintf("… +%d more key(s)", nMiss-5))
						break
					}
					parts = append(parts, fmt.Sprintf("%s=%d", d.LocalIndicatorKey, d.Value))
				}
				out.DetailMessages = append(out.DetailMessages, fmt.Sprintf(
					"Org unit %s: 0 data values to send — %d key(s) lack DHIS2 UIDs in dhis2_mapping_item. Examples: %s",
					p.OrgUnitUID, nMiss, strings.Join(parts, "; ")))
			} else {
				out.DetailMessages = append(out.DetailMessages, fmt.Sprintf(
					"Org unit %s: 0 data values to send (no mapped non-zero cells, or all counts zero).",
					p.OrgUnitUID))
			}
			continue
		}

		filtered := []dhis2DataValue{}
		for _, dv := range p.DataValues {
			checksum := FPReportChecksum(dv.Period, dv.OrgUnit, "", dv.Value)
			if !force {
				status, err := s.dhisRepo.GetCellStatus(dv.Period, dv.OrgUnit, "")
				if err == nil && status.SyncStatus == "synced" && status.Checksum == checksum {
					continue
				}
			}
			filtered = append(filtered, dv)
		}
		if len(filtered) == 0 {
			out.SkippedAlreadySynced++
			out.DetailMessages = append(out.DetailMessages, fmt.Sprintf(
				"Org unit %s: built %d data value(s) but all were skipped as unchanged (already synced). Enable “Force resync” to resend.",
				p.OrgUnitUID, len(p.DataValues)))
			continue
		}

		payload := dhis2Payload{DataValues: filtered}
		bodyBytes, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/dataValueSets", strings.TrimRight(baseURL, "/")), bytes.NewReader(bodyBytes))
		if err != nil {
			return out, err
		}
		req.Header.Set("Content-Type", "application/json")
		if settings.AuthType == "basic" && settings.Username != "" {
			user := settings.Username
			pass := settings.PasswordEncrypted
			auth := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
			req.Header.Set("Authorization", "Basic "+auth)
		} else if settings.AuthType == "token" && settings.TokenEncrypted != "" {
			req.Header.Set("Authorization", "Bearer "+settings.TokenEncrypted)
		}

		resp, httpErr := s.httpClient.Do(req)
		statusCode := 0
		respBody := ""
		success := false
		if httpErr != nil {
			respBody = httpErr.Error()
		} else if resp != nil {
			statusCode = resp.StatusCode
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(resp.Body)
			_ = resp.Body.Close()
			respBody = buf.String()
			success = statusCode >= 200 && statusCode < 300
		}

		log := models.ReportSyncLog{
			Period:             period,
			OrgUnitUID:         p.OrgUnitUID,
			TriggerType:        "manual",
			SyncScope:          "monthly_fp_methods",
			PayloadJSON:        string(bodyBytes),
			ResponseStatusCode: statusCode,
			ResponseBody:       respBody,
			Success:            success,
			ForceResync:        force,
			InitiatedBy:        initiatedBy,
		}
		if err := s.dhisRepo.CreateSyncLog(&log); err != nil {
			return out, err
		}
		out.Logs = append(out.Logs, log)
		out.PostedBatches++

		now := time.Now()
		for _, dv := range filtered {
			cs := models.ReportCellSyncStatus{
				Period:            dv.Period,
				LocalIndicatorKey: "",
				OrgUnitUID:        dv.OrgUnit,
				Value:             dv.Value,
				LastSyncedAt:      &now,
				LastSyncLogID:     &log.ID,
				SyncStatus:        map[bool]string{true: "synced", false: "failed"}[success],
				Checksum:          FPReportChecksum(dv.Period, dv.OrgUnit, "", dv.Value),
			}
			if err := s.dhisRepo.UpsertCellStatus(&cs); err != nil {
				return out, err
			}
		}
	}

	if out.PostedBatches == 0 && len(out.DetailMessages) == 0 {
		out.DetailMessages = append(out.DetailMessages, "Nothing was posted to DHIS2.")
	}
	return out, nil
}
