package service

import (
	_ "embed"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"fpreg/internal/models"
)

//go:embed dhis2_mapping_official.csv
var dhis2MappingOfficialCSV string

var deNameFPCode = regexp.MustCompile(`FP(\d{2})\.`)

// OfficialDHIS2MappingItemsFromEmbeddedCSV parses the bundled DHIS2 metadata export into
// mapping rows keyed like AggregateForPeriod (FP01_NEW_25_49, FP05_PA_NEW_BELOW_15, …).
// Skips FP21 (totals), FP22 (no age grid), FP23 / FP24 (duplicate condom DEs — app uses FP13 / FP14).
func OfficialDHIS2MappingItemsFromEmbeddedCSV() ([]models.DHIS2MappingItem, error) {
	r := csv.NewReader(strings.NewReader(dhis2MappingOfficialCSV))
	r.TrimLeadingSpace = true
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("dhis2 csv header: %w", err)
	}
	if len(header) < 4 {
		return nil, fmt.Errorf("dhis2 csv: expected at least 4 columns")
	}

	var out []models.DHIS2MappingItem
	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(rec) < 4 {
			continue
		}
		deName := strings.TrimSpace(rec[0])
		combo := strings.TrimSpace(rec[1])
		deUID := strings.TrimSpace(rec[2])
		cocUID := strings.TrimSpace(rec[3])
		if deUID == "" || cocUID == "" {
			continue
		}

		fpNum, ok := fpNumberFromDEName(deName)
		if !ok {
			continue
		}
		switch fpNum {
		case 21, 22, 23, 24:
			continue
		}

		methodCode := fmt.Sprintf("FP%02d", fpNum)
		subgroup, visit, age, ok := parseCategoryCombo(methodCode, combo)
		if !ok {
			return nil, fmt.Errorf("dhis2 csv: unhandled combo for %q: %q", deName, combo)
		}

		localKey := buildLocalIndicatorKey(methodCode, subgroup, visit, age)
		item := stubDHIS2MappingItemForKey(localKey)
		item.DHIS2DataElementUID = deUID
		item.DHIS2CatOptionComboUID = cocUID
		item.Active = true
		item.Notes = "DHIS2 UIDs from embedded dhis2_mapping_official.csv (export)."
		out = append(out, item)
	}
	return out, nil
}

func fpNumberFromDEName(deName string) (int, bool) {
	m := deNameFPCode.FindStringSubmatch(deName)
	if len(m) < 2 {
		return 0, false
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, false
	}
	return n, true
}

func parseCategoryCombo(methodCode, combo string) (subgroup, visit, age string, ok bool) {
	combo = strings.TrimSpace(combo)
	if methodCode == "FP05" {
		return parseFP05Combo(combo)
	}
	if strings.HasPrefix(combo, "New, ") {
		age, ok := normalizeCategoryAge(strings.TrimPrefix(combo, "New, "))
		if !ok {
			return "", "", "", false
		}
		return "", "NEW", age, true
	}
	if strings.HasPrefix(combo, "Revisits, ") {
		age, ok := normalizeCategoryAge(strings.TrimPrefix(combo, "Revisits, "))
		if !ok {
			return "", "", "", false
		}
		return "", "REVISIT", age, true
	}
	return "", "", "", false
}

func parseFP05Combo(combo string) (subgroup, visit, age string, ok bool) {
	var rest string
	switch {
	case strings.HasPrefix(combo, "Self Injected (SI), "):
		subgroup = "SI"
		rest = strings.TrimPrefix(combo, "Self Injected (SI), ")
	case strings.HasPrefix(combo, "Provider Administered (PA), "):
		subgroup = "PA"
		rest = strings.TrimPrefix(combo, "Provider Administered (PA), ")
	default:
		return "", "", "", false
	}
	rest = strings.TrimSpace(rest)
	if strings.HasPrefix(rest, "New, ") {
		a, ok := normalizeCategoryAge(strings.TrimPrefix(rest, "New, "))
		if !ok {
			return "", "", "", false
		}
		return subgroup, "NEW", a, true
	}
	if strings.HasPrefix(rest, "Revisits, ") {
		a, ok := normalizeCategoryAge(strings.TrimPrefix(rest, "Revisits, "))
		if !ok {
			return "", "", "", false
		}
		return subgroup, "REVISIT", a, true
	}
	return "", "", "", false
}

func normalizeCategoryAge(s string) (string, bool) {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, "<15") || strings.Contains(strings.ToLower(s), "below 15"):
		return "BELOW_15", true
	case strings.Contains(s, "15-19"):
		return "15_19", true
	case strings.Contains(s, "20-24"):
		return "20_24", true
	case strings.Contains(s, "25-49"):
		return "25_49", true
	case strings.Contains(s, "50+"):
		return "50_PLUS", true
	default:
		return "", false
	}
}
