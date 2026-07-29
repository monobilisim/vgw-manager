package api

import (
	"encoding/json"
	"errors"
	"net/http"
)

// Sentinel errors.
var errNotFound = errors.New("not found")

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes a JSON error body: {"error": "message"}.
func writeError(w http.ResponseWriter, status int, err error) {
	msg := err.Error()
	writeJSON(w, status, map[string]string{"error": msg})
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

