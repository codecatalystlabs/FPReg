package service

import (
	"strings"

	"fpreg/internal/models"
)

// fpMethodVariants lists method codes used by AggregateForPeriod / determineMethods (FP05 split into PA/SI).
func fpMethodVariants() []struct{ Base, Sub string } {
	return []struct{ Base, Sub string }{
		{"FP01", ""}, {"FP02", ""}, {"FP03", ""}, {"FP04", ""},
		{"FP05", "PA"}, {"FP05", "SI"},
		{"FP06", ""}, {"FP07", ""}, {"FP08", ""}, {"FP09", ""},
		{"FP10", ""}, {"FP11", ""}, {"FP12", ""}, {"FP13", ""}, {"FP14", ""},
	}
}

// EnumerateAllLocalIndicatorKeys returns every local_indicator_key the monthly FP aggregate can emit.
func EnumerateAllLocalIndicatorKeys() []string {
	visits := []string{"NEW", "REVISIT"}
	ages := []string{"BELOW_15", "16_19", "20_24", "25_49", "50_PLUS"}
	var keys []string
	for _, mv := range fpMethodVariants() {
		for _, vt := range visits {
			for _, ag := range ages {
				keys = append(keys, buildLocalIndicatorKey(mv.Base, mv.Sub, vt, ag))
			}
		}
	}
	return keys
}

var fpMethodDisplayNames = map[string]string{
	"FP01": "CoC pills",
	"FP02": "POP",
	"FP03": "ECP",
	"FP04": "Injectable DMPA-IM",
	"FP05": "Injectable DMPA-SC",
	"FP06": "Implant 3y",
	"FP07": "Implant 5y",
	"FP08": "IUD Copper-T",
	"FP09": "IUD Hormonal",
	"FP10": "FAM SDM",
	"FP11": "FAM LAM",
	"FP12": "FAM Two-day",
	"FP13": "Female condoms",
	"FP14": "Male condoms",
}

// StubDHIS2MappingItems returns inactive template rows (no DHIS2 UIDs) for every possible indicator key.
func StubDHIS2MappingItems() []models.DHIS2MappingItem {
	var out []models.DHIS2MappingItem
	for _, key := range EnumerateAllLocalIndicatorKeys() {
		out = append(out, stubDHIS2MappingItemForKey(key))
	}
	return out
}

func stubDHIS2MappingItemForKey(key string) models.DHIS2MappingItem {
	methodCombo, visitType, ageGroup := ParseLocalIndicatorKey(key)
	base := methodCombo
	sub := ""
	if strings.HasPrefix(methodCombo, "FP05_") {
		base = "FP05"
		sub = strings.TrimPrefix(methodCombo, "FP05_")
	}
	name := fpMethodDisplayNames[base]
	if name == "" {
		name = base
	}
	if sub != "" {
		name = name + " (" + sub + ")"
	}
	return models.DHIS2MappingItem{
		LocalIndicatorKey:      key,
		MethodCode:             base,
		Subgroup:               sub,
		MethodName:             name,
		VisitType:              visitType,
		AgeGroup:               ageGroup,
		Active:                 false,
		DHIS2DataElementUID:    "",
		DHIS2CatOptionComboUID: "",
		Notes:                  "Set dhis2_data_element_uid and dhis2_cat_option_combo_uid from DHIS2 Maintenance; sync runs when both are non-empty.",
	}
}
