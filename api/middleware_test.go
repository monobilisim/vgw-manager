package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/monobilisim/vgw-manager/config"
)

func TestAuthMiddlewareRequiresBearerScheme(t *testing.T) {
	token := strings.Repeat("t", 32)
	config.APIToken = token
	t.Cleanup(func() { config.APIToken = "" })

	h := authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		header string
		want   int
	}{
		{"Bearer " + config.APIToken, http.StatusNoContent},
		{config.APIToken, http.StatusUnauthorized},
		{"Basic " + config.APIToken, http.StatusUnauthorized},
		{"Bearer wrong", http.StatusUnauthorized},
		{"", http.StatusUnauthorized},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, "/v1/buckets", nil)
		if tt.header != "" {
			r.Header.Set("Authorization", tt.header)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tt.want {
			t.Errorf("Authorization %q: status %d, want %d", tt.header, w.Code, tt.want)
		}
	}
}

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		body string
		ok   bool
	}{
		{"", true},
		{`{"owner":"alice"}`, true},
		{`{"owner":`, false},
		{`not json`, false},
		{`{"unknown":1}`, false},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
		w := httptest.NewRecorder()
		var v makePublicRequest
		if got := decodeJSON(w, r, &v); got != tt.ok {
			t.Errorf("decodeJSON(%q) = %v, want %v", tt.body, got, tt.ok)
		}
		if !tt.ok && w.Code != http.StatusBadRequest {
			t.Errorf("decodeJSON(%q) status %d, want 400", tt.body, w.Code)
		}
	}
}
