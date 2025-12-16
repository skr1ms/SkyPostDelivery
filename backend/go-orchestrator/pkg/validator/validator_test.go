package validator

import (
	"testing"
)

func TestValidateRussianPhone(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		valid bool
	}{
		{"Valid +7", "+79991234567", true},
		{"Valid 8", "89991234567", true},
		{"Valid 7", "79991234567", true},
		{"Valid with spaces", "+7 999 123 45 67", true},
		{"Valid with dashes", "+7-999-123-45-67", true},
		{"Valid with brackets", "+7 (999) 123-45-67", true},
		{"Valid 8 with spaces", "8 999 123 45 67", true},
		{"Invalid short", "+799912345", false},
		{"Invalid long", "+799912345678", false},
		{"Invalid empty", "", false},
		{"Invalid letters", "+7999abc4567", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateRussianPhone(tt.phone)
			if result != tt.valid {
				t.Errorf("ValidateRussianPhone(%s) = %v, want %v", tt.phone, result, tt.valid)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"Valid simple", "test@example.com", true},
		{"Valid with dots", "test.user@example.com", true},
		{"Valid with plus", "test+tag@example.com", true},
		{"Valid subdomain", "test@mail.example.com", true},
		{"Invalid no @", "testexample.com", false},
		{"Invalid no domain", "test@", false},
		{"Invalid empty", "", false},
		{"Invalid spaces", "test @example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateEmail(tt.email)
			if result != tt.valid {
				t.Errorf("ValidateEmail(%s) = %v, want %v", tt.email, result, tt.valid)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"Valid simple", "password123", true},
		{"Valid min length", "pass12", true},
		{"Valid with special", "Pass@123!", true},
		{"Invalid too short", "pass1", false},
		{"Invalid empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePassword(tt.password)
			if result != tt.valid {
				t.Errorf("ValidatePassword(%s) = %v, want %v", tt.password, result, tt.valid)
			}
		})
	}
}

func TestNormalizeRussianPhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{"From 8", "89991234567", "+79991234567"},
		{"From 7", "79991234567", "+79991234567"},
		{"With spaces", "+7 999 123 45 67", "+79991234567"},
		{"With dashes", "+7-999-123-45-67", "+79991234567"},
		{"With brackets", "+7 (999) 123-45-67", "+79991234567"},
		{"Already normalized", "+79991234567", "+79991234567"},
		{"From 10 digits", "9991234567", "+79991234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeRussianPhone(tt.phone)
			if result != tt.expected {
				t.Errorf("NormalizeRussianPhone(%s) = %s, want %s", tt.phone, result, tt.expected)
			}
		})
	}
}
