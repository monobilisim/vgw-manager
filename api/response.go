package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/monobilisim/vgw-manager/services"
)

// Sentinel errors.
var (
	errNotFound            = errors.New("not found")
	errInternalServerError = errors.New("internal server error")
)

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body: {"error": "message"}.
// Errors wrapping services.ErrInvalidInput are client errors (400). Other
// 500s are logged server-side and answered with a generic message so gateway
// responses and file paths are not leaked to clients.
func writeError(w http.ResponseWriter, status int, err error) {
	if errors.Is(err, services.ErrInvalidInput) {
		if status == http.StatusInternalServerError {
			status = http.StatusBadRequest
		}
		writeJSON(w, status, map[string]string{"error": err.Error()})
		return
	}
	if status == http.StatusInternalServerError {
		slog.Error("request failed", "error", err)
		err = errInternalServerError
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// decodeJSON reads a JSON body with a 1MB limit and disallows unknown fields.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		// Ignore EOF for optional bodies (e.g. make-public with no body).
		if err.Error() != "EOF" {
			writeError(w, http.StatusBadRequest, err)
			return false
		}
	}
	return true
}
