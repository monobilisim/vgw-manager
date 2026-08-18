package services

import "testing"

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
