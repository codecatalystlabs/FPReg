package service

import "testing"

func TestParseLocalIndicatorKey(t *testing.T) {
	tests := []struct {
		key, method, visit, age string
	}{
		{"FP01_NEW_BELOW_15", "FP01", "NEW", "BELOW_15"},
		{"FP01_REVISIT_25_49", "FP01", "REVISIT", "25_49"},
		{"FP05_PA_NEW_20_24", "FP05_PA", "NEW", "20_24"},
		{"FP05_SI_REVISIT_50_PLUS", "FP05_SI", "REVISIT", "50_PLUS"},
	}
	for _, tc := range tests {
		m, v, a := ParseLocalIndicatorKey(tc.key)
		if m != tc.method || v != tc.visit || a != tc.age {
			t.Fatalf("%q => got (%q,%q,%q) want (%q,%q,%q)", tc.key, m, v, a, tc.method, tc.visit, tc.age)
		}
	}
}
