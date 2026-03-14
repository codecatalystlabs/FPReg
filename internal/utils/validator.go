package utils

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var dateRegex = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

func IsValidDate(date string) bool {
	return dateRegex.MatchString(date)
}

func IsValidSex(sex string) bool {
	s := strings.ToUpper(sex)
	return s == "M" || s == "F"
}

func SanitizeString(s string) string {
	return strings.TrimSpace(s)
}

func IsInSet(value string, valid []string) bool {
	for _, v := range valid {
		if v == value {
			return true
		}
	}
	return false
}

type RegistrationValidationResult struct {
	Errors []ErrorDetail
}

func ValidateRegistration(data map[string]interface{}) RegistrationValidationResult {
	var result RegistrationValidationResult

	requireString := func(field, label string) string {
		val, ok := data[field]
		if !ok {
			result.Errors = append(result.Errors, ErrorDetail{Field: field, Message: label + " is required"})
			return ""
		}
		s, ok := val.(string)
		if !ok || strings.TrimSpace(s) == "" {
			result.Errors = append(result.Errors, ErrorDetail{Field: field, Message: label + " is required"})
			return ""
		}
		return strings.TrimSpace(s)
	}

	visitDate := requireString("visit_date", "Visit date")
	if visitDate != "" && !IsValidDate(visitDate) {
		result.Errors = append(result.Errors, ErrorDetail{Field: "visit_date", Message: "Visit date must be YYYY-MM-DD"})
	}

	requireString("surname", "Surname")
	requireString("given_name", "Given name")

	sex := requireString("sex", "Sex")
	if sex != "" && !IsValidSex(sex) {
		result.Errors = append(result.Errors, ErrorDetail{Field: "sex", Message: "Sex must be M or F"})
	}

	if age, ok := data["age"]; ok {
		switch v := age.(type) {
		case float64:
			if v < 0 || v > 120 {
				result.Errors = append(result.Errors, ErrorDetail{Field: "age", Message: "Age must be between 0 and 120"})
			}
		default:
			result.Errors = append(result.Errors, ErrorDetail{Field: "age", Message: "Age must be a number"})
		}
	} else {
		result.Errors = append(result.Errors, ErrorDetail{Field: "age", Message: "Age is required"})
	}

	isNew, _ := data["is_new_user"].(bool)
	isRevisit, _ := data["is_revisit"].(bool)
	if isNew && isRevisit {
		result.Errors = append(result.Errors, ErrorDetail{Field: "is_new_user", Message: "Client cannot be both new user and revisit"})
	}
	if !isNew && !isRevisit {
		result.Errors = append(result.Errors, ErrorDetail{Field: "is_new_user", Message: "Client must be either new user or revisit"})
	}

	if isRevisit {
		prev, _ := data["previous_method"].(string)
		if strings.TrimSpace(prev) == "" {
			result.Errors = append(result.Errors, ErrorDetail{Field: "previous_method", Message: "Previous method is required for revisit clients"})
		}
	}

	isSwitching, _ := data["is_switching"].(bool)
	if isSwitching {
		reason, _ := data["switching_reason"].(string)
		if strings.TrimSpace(reason) == "" {
			result.Errors = append(result.Errors, ErrorDetail{Field: "switching_reason", Message: "Switching reason is required when switching method"})
		}
	}

	return result
}
