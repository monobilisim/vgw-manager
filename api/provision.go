package api

import (
	"net/http"

	"github.com/monobilisim/vgw-manager/services"
)

// handleProvision creates a user, a bucket, and sets the bucket owner in one call.
func handleProvision(w http.ResponseWriter, r *http.Request) {
	var req services.ProvisionRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	summary, err := services.Provision(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, summary)
}
