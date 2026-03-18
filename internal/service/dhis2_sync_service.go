package service

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"fpreg/internal/config"
	"fpreg/internal/models"
	"fpreg/internal/repository"

	"github.com/google/uuid"
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

type SyncPreview struct {
	Period          string              `json:"period"`
	OrgUnitUID      string              `json:"orgunit_uid"`
	Cells           []FPCellWithMapping `json:"cells"`
	DataValues      []dhis2DataValue    `json:"data_values"`
	MissingMappings []string            `json:"missing_mappings"`
	MissingOrgUnits []string            `json:"missing_org_units"`
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
		if err != nil || fac == nil || fac.UID == "" {
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
		orgUnit := fac.UID

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

			mapping, err := s.dhisRepo.FindMappingByLocalKey(r.LocalIndicatorKey)
			if err != nil || mapping == nil || !mapping.Active || mapping.DHIS2DataElementUID == "" || mapping.DHIS2CatOptionComboUID == "" {
				cell.ExcludedReason = "MISSING_MAPPING"
				missingMappings[r.LocalIndicatorKey] = true
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
func (s *DHIS2SyncService) Sync(period string, facilityIDs []uuid.UUID, force bool, initiatedBy string) ([]models.ReportSyncLog, error) {
	previews, err := s.BuildPreview(period, facilityIDs)
	if err != nil {
		return nil, err
	}
	settings, err := s.dhisRepo.GetActiveSettings()
	if err != nil {
		return nil, fmt.Errorf("no active DHIS2 settings: %w", err)
	}

	baseURL := settings.DHIS2BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("dhis2 base url not configured")
	}

	var logs []models.ReportSyncLog

	for _, p := range previews {
		if len(p.DataValues) == 0 {
			continue
		}

		// Filter out cells already synced (idempotency)
		filtered := []dhis2DataValue{}
		for _, dv := range p.DataValues {
			checksum := FPReportChecksum(dv.Period, dv.OrgUnit, "", dv.Value)
			if !force {
				status, err := s.dhisRepo.GetCellStatus(dv.Period, dv.OrgUnit, "") // using empty key for now; could extend to local_indicator_key in future
				if err == nil && status.SyncStatus == "synced" && status.Checksum == checksum {
					continue
				}
			}
			filtered = append(filtered, dv)
		}
		if len(filtered) == 0 {
			continue
		}

		payload := dhis2Payload{DataValues: filtered}
		bodyBytes, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/dataValueSets", baseURL), bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		// Basic auth from settings
		if settings.AuthType == "basic" && settings.Username != "" {
			user := settings.Username
			pass := settings.PasswordEncrypted // assume plaintext or pre-encrypted as needed
			auth := base64.StdEncoding.EncodeToString([]byte(user + ":" + pass))
			req.Header.Set("Authorization", "Basic "+auth)
		} else if settings.AuthType == "token" && settings.TokenEncrypted != "" {
			req.Header.Set("Authorization", "Bearer "+settings.TokenEncrypted)
		}

		resp, err := s.httpClient.Do(req)
		statusCode := 0
		respBody := ""
		success := false
		if err == nil && resp != nil {
			defer resp.Body.Close()
			statusCode = resp.StatusCode
			var buf bytes.Buffer
			_, _ = buf.ReadFrom(resp.Body)
			respBody = buf.String()
			success = statusCode >= 200 && statusCode < 300
		} else if err != nil {
			respBody = err.Error()
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
			return logs, err
		}
		logs = append(logs, log)

		// Update cell statuses
		now := time.Now()
		for _, dv := range filtered {
			cs := models.ReportCellSyncStatus{
				Period:            dv.Period,
				LocalIndicatorKey: "", // could extend to carry local key for more granular status
				OrgUnitUID:        dv.OrgUnit,
				Value:             dv.Value,
				LastSyncedAt:      &now,
				LastSyncLogID:     &log.ID,
				SyncStatus:        map[bool]string{true: "synced", false: "failed"}[success],
				Checksum:          FPReportChecksum(dv.Period, dv.OrgUnit, "", dv.Value),
			}
			if err := s.dhisRepo.UpsertCellStatus(&cs); err != nil {
				return logs, err
			}
		}
	}

	return logs, nil
}
