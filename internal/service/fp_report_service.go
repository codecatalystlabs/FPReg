package service

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"fpreg/internal/models"
	"fpreg/internal/repository"

	"github.com/google/uuid"
)

type FPReportAggregationRow struct {
	Period           string
	FacilityID       uuid.UUID
	LocalIndicatorKey string
	Value            int
}

type FPReportService struct {
	regRepo      *repository.RegistrationRepository
	facilityRepo *repository.FacilityRepository
	reportRepo   *repository.FPReportRepository
}

func NewFPReportService(
	regRepo *repository.RegistrationRepository,
	facilityRepo *repository.FacilityRepository,
) *FPReportService {
	return &FPReportService{
		regRepo:      regRepo,
		facilityRepo: facilityRepo,
		reportRepo:   repository.NewFPReportRepository(regRepo.DB()),
	}
}

// AggregateForPeriod aggregates registrations for given period (YYYYMM) and optional facility IDs.
func (s *FPReportService) AggregateForPeriod(period string, facilityIDs []uuid.UUID) ([]FPReportAggregationRow, error) {
	start, end, err := periodBounds(period)
	if err != nil {
		return nil, err
	}
	regs, err := s.reportRepo.ListForRange(start, end, facilityIDs)
	if err != nil {
		return nil, err
	}

	type key struct {
		FacilityID       uuid.UUID
		LocalIndicatorKey string
	}
	agg := map[key]int{}

	for _, r := range regs {
		ageGroup := classifyAge(r.Age)
		visitType := "REVISIT"
		if r.IsNewUser && !r.IsRevisit {
			visitType = "NEW"
		}

		methodKeys := determineMethods(&r)
		for _, mk := range methodKeys {
			if mk == "" {
				continue
			}
			methodCode, subgroup := splitMethodSubgroup(mk)
			localKey := buildLocalIndicatorKey(methodCode, subgroup, visitType, ageGroup)
			k := key{FacilityID: r.FacilityID, LocalIndicatorKey: localKey}
			agg[k]++
		}
	}

	out := make([]FPReportAggregationRow, 0, len(agg))
	for k, v := range agg {
		out = append(out, FPReportAggregationRow{
			Period:           period,
			FacilityID:       k.FacilityID,
			LocalIndicatorKey: k.LocalIndicatorKey,
			Value:            v,
		})
	}
	return out, nil
}

func periodBounds(period string) (time.Time, time.Time, error) {
	// period format YYYYMM
	t, err := time.Parse("200601", period)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	start := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, 0)
	return start, end, nil
}

func classifyAge(age int) string {
	switch {
	case age < 15:
		return "BELOW_15"
	case age >= 15 && age <= 19:
		return "16_19"
	case age >= 20 && age <= 24:
		return "20_24"
	case age >= 25 && age <= 49:
		return "25_49"
	default:
		return "50_PLUS"
	}
}

// determineMethods returns internal method keys like FP01, FP05_PA, FP05_SI, FP14, etc.
func determineMethods(r *models.FPRegistration) []string {
	var out []string
	if r.PillsCOCCycles > 0 {
		out = append(out, "FP01")
	}
	if r.PillsPOPCycles > 0 {
		out = append(out, "FP02")
	}
	if r.PillsECPPieces > 0 {
		out = append(out, "FP03")
	}
	if r.InjectableDMPAIMDoses > 0 {
		out = append(out, "FP04")
	}
	if r.InjectableDMPASCPADoses > 0 {
		out = append(out, "FP05_PA")
	}
	if r.InjectableDMPASCSIDoses > 0 {
		out = append(out, "FP05_SI")
	}
	if r.Implant3Years {
		out = append(out, "FP06")
	}
	if r.Implant5Years {
		out = append(out, "FP07")
	}
	if r.IUDCopperT {
		out = append(out, "FP08")
	}
	if r.IUDHormonal3Years || r.IUDHormonal5Years {
		out = append(out, "FP09")
	}
	if r.FAMStandardDays {
		out = append(out, "FP10")
	}
	if r.FAMLAM {
		out = append(out, "FP11")
	}
	if r.FAMTwoDay {
		out = append(out, "FP12")
	}
	if r.CondomsFemaleUnits > 0 {
		out = append(out, "FP13")
	}
	if r.CondomsMaleUnits > 0 {
		out = append(out, "FP14")
	}
	return out
}

func splitMethodSubgroup(m string) (methodCode, subgroup string) {
	// FP05_PA => FP05, PA
	if len(m) > 5 && m[:4] == "FP05" {
		methodCode = "FP05"
		if len(m) > 6 {
			subgroup = m[5:]
		}
		return
	}
	return m, ""
}

func buildLocalIndicatorKey(methodCode, subgroup, visitType, ageGroup string) string {
	if subgroup != "" {
		return methodCode + "_" + subgroup + "_" + visitType + "_" + ageGroup
	}
	return methodCode + "_" + visitType + "_" + ageGroup
}

// checksum helper for cell status
func FPReportChecksum(period, orgUnit, localKey string, value int) string {
	h := sha1.New()
	_ = json.NewEncoder(h).Encode(struct {
		Period string
		Org    string
		Key    string
		Value  int
	}{period, orgUnit, localKey, value})
	return hex.EncodeToString(h.Sum(nil))
}

// ParseLocalIndicatorKey splits keys like FP01_NEW_BELOW_15 or FP05_PA_REVISIT_20_24.
func ParseLocalIndicatorKey(key string) (methodCode, visitType, ageGroup string) {
	parts := strings.Split(key, "_")
	if len(parts) < 3 {
		return key, "", ""
	}
	// FP05_PA / FP05_SI subgroup: FP05_PA_NEW_25_49 → 5+ segments
	if len(parts) >= 5 && parts[0] == "FP05" && (parts[1] == "PA" || parts[1] == "SI") {
		methodCode = parts[0] + "_" + parts[1]
		visitType = parts[2]
		ageGroup = strings.Join(parts[3:], "_")
		return
	}
	methodCode = parts[0]
	visitType = parts[1]
	ageGroup = parts[2]
	if len(parts) > 3 {
		ageGroup = strings.Join(parts[2:], "_")
	}
	return
}

