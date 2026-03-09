package domain_test

import (
	"testing"

	"auth-service/internal/domain"
)

func TestPasswordPolicy_Validations(t *testing.T) {
	policy := domain.NewDefaultPasswordPolicy()

	tests := []struct {
		name    string
		pwd     string
		wantErr bool
	}{
		{"Valid password", "Str0ng!Pass1", false},
		{"Too short", "S!p1", true},
		{"No uppercase", "str0ng!pass1", true},
		{"No lowercase", "STR0NG!PASS1", true},
		{"No digits", "StrOng!Pass", true},
		{"No special chars", "Str0ngPass123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := policy.Validate(tt.pwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%q) error = %v, wantErr %v", tt.pwd, err, tt.wantErr)
			}
		})
	}
}

func TestPasswordPolicy_HashAndVerify(t *testing.T) {
	policy := domain.NewDefaultPasswordPolicy()
	password := "Str0ng!Pass1"

	hash, err := policy.Hash(password)
	if err != nil {
		t.Fatalf("Hash() unexpected error: %v", err)
	}

	if !policy.Verify(password, hash) {
		t.Errorf("Verify() failed for valid password and hash")
	}

	if policy.Verify("WrongPass1!", hash) {
		t.Errorf("Verify() succeeded for invalid password")
	}
}
