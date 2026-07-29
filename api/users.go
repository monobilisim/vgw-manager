package api

import (
	"errors"
	"net/http"

	"github.com/monobilisim/vgw-manager/models"
	"github.com/monobilisim/vgw-manager/services"
)

// Validation errors.
var (
	errBucketNameRequired      = errors.New("bucket name is required")
	errBucketNameQuotaRequired = errors.New("bucket name and quota are required")
	errAccessSecretRequired    = errors.New("access and secret are required")
	errInvalidRole             = errors.New("role must be admin, user, or userplus")
)

// maskSecret replaces the secret field with "***" unless showSecrets is true.
func maskSecret(user models.User, showSecrets bool) models.User {
	if !showSecrets {
		user.Secret = "***"
	}
	return user
}

// handleListUsers returns all users. Secrets are masked unless ?showSecrets=true.
func handleListUsers(w http.ResponseWriter, r *http.Request) {
	userService := services.NewUserService()
	users, err := userService.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	showSecrets := r.URL.Query().Get("showSecrets") == "true"
	masked := make([]models.User, len(users))
	for i, u := range users {
		masked[i] = maskSecret(u, showSecrets)
	}
	writeJSON(w, http.StatusOK, masked)
}

// handleGetUser returns a single user by access key. Secret is masked unless ?showSecrets=true.
func handleGetUser(w http.ResponseWriter, r *http.Request) {
	access := r.PathValue("access")
	if access == "" {
		writeError(w, http.StatusBadRequest, errors.New("access key is required"))
		return
	}

	userService := services.NewUserService()
	user, err := userService.GetUser(access)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	showSecrets := r.URL.Query().Get("showSecrets") == "true"
	writeJSON(w, http.StatusOK, maskSecret(*user, showSecrets))
}

// createUserRequest is the JSON body for POST /v1/users.
type createUserRequest struct {
	Access    string `json:"access"`
	Secret    string `json:"secret"`
	Role      string `json:"role"`
	UserID    int    `json:"userID"`
	GroupID   int    `json:"groupID"`
	ProjectID int    `json:"projectID"`
}

// handleCreateUser creates a new user via the VersityGW API.
func handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if req.Access == "" || req.Secret == "" {
		writeError(w, http.StatusBadRequest, errAccessSecretRequired)
		return
	}
	if req.Role != "admin" && req.Role != "user" && req.Role != "userplus" {
		writeError(w, http.StatusBadRequest, errInvalidRole)
		return
	}

	userReq := models.UserCreateRequest{
		Access:    req.Access,
		Secret:    req.Secret,
		Role:      req.Role,
		UserID:    req.UserID,
		GroupID:   req.GroupID,
		ProjectID: req.ProjectID,
	}

	vgwService := services.NewVersityGWService()
	if err := vgwService.CreateUser(userReq); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"access": req.Access,
		"role":   req.Role,
	})
}

// handleDeleteUser deletes a user by access key.
func handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	access := r.PathValue("access")
	if access == "" {
		writeError(w, http.StatusBadRequest, errors.New("access key is required"))
		return
	}

	vgwService := services.NewVersityGWService()
	if err := vgwService.DeleteUser(access); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"access": access,
		"status": "deleted",
	})
}
