package services

import (
	"strings"
	"testing"
)

func TestPolicyGrantsPublicAccess(t *testing.T) {
	tests := []struct {
		name   string
		policy string
		public bool
	}{
		{
			name:   "wildcard principal",
			policy: `{"Statement":[{"Principal":"*"}]}`,
			public: true,
		},
		{
			name:   "wildcard aws principal",
			policy: `{"Statement":[{"Principal":{"AWS":"*"}}]}`,
			public: true,
		},
		{
			name:   "wildcard principal list",
			policy: `{"Statement":[{"Principal":{"AWS":["alice","*"]}}]}`,
			public: true,
		},
		{
			name:   "named principal",
			policy: `{"Statement":[{"Principal":["ilkay.atamer"]}]}`,
			public: false,
		},
		{
			name:   "invalid policy",
			policy: `{"Statement":`,
			public: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			public, err := policyGrantsPublicAccess(tt.policy)
			if tt.name == "invalid policy" {
				if err == nil {
					t.Fatal("expected invalid policy error")
				}
				return
			}
			if err != nil {
				t.Fatalf("policyGrantsPublicAccess() error = %v", err)
			}
			if public != tt.public {
				t.Fatalf("policyGrantsPublicAccess() = %v, want %v", public, tt.public)
			}
		})
	}
}

func TestValidateBucketNameAndAccessKey(t *testing.T) {
	badNames := []string{
		"",
		"../../etc",
		"foo/bar",
		"foo@snap",
		"-option",
		".hidden",
		"trailingdot.",
		"double..dot",
		"UPPER",
		"with space",
		strings.Repeat("a", 64),
	}
	for _, name := range badNames {
		if err := validateBucketName(name); err == nil {
			t.Errorf("validateBucketName(%q) accepted invalid name", name)
		}
	}
	if err := validateBucketName("my-bucket.1"); err != nil {
		t.Errorf("validateBucketName rejected valid name: %v", err)
	}

	badKeys := []string{"", `quote"key`, "back\\slash", "ctrl\x01"}
	for _, key := range badKeys {
		if validAccessKey(key) {
			t.Errorf("validAccessKey(%q) accepted invalid key", key)
		}
	}
	if !validAccessKey("AKIAEXAMPLE123") {
		t.Error("validAccessKey rejected valid key")
	}
}
