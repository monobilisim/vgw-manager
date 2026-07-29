package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/monobilisim/vgw-manager/models"
)

// ProvisionRequest holds the parameters for provisioning a user and bucket.
type ProvisionRequest struct {
	Access string
	Secret string
	Role   string

	UserID    int
	GroupID   int
	ProjectID int

	Bucket string
	Quota  string
	Owner  string
}

// ProvisionSummary holds the result of a provisioning operation.
type ProvisionSummary struct {
	Access string `json:"access"`
	Secret string `json:"secret"`
	Role   string `json:"role"`

	UserID    int `json:"userID"`
	GroupID   int `json:"groupID"`
	ProjectID int `json:"projectID"`

	Bucket string `json:"bucket"`
	Quota  string `json:"quota"`
	Owner  string `json:"owner"`

	SecretGenerated bool `json:"secretGenerated"`
}

// Provision creates a user, creates a bucket, and sets the bucket owner.
func Provision(req ProvisionRequest) (ProvisionSummary, error) {
	summary := ProvisionSummary{}

	if req.Access == "" {
		return summary, fmt.Errorf("access key is required (use --access)")
	}
	if req.Role != "admin" && req.Role != "user" && req.Role != "userplus" {
		return summary, fmt.Errorf("role must be admin, user, or userplus")
	}
	if req.Bucket == "" {
		return summary, fmt.Errorf("bucket name is required (use --bucket)")
	}
	if req.Quota == "" {
		return summary, fmt.Errorf("quota is required (use --quota, e.g. 2T)")
	}
	if req.Owner == "" {
		req.Owner = req.Access
	}

	secretGenerated := false
	if req.Secret == "" {
		req.Secret = GenerateSecretKey()
		secretGenerated = true
	}

	vgwService := NewVersityGWService()
	bucketService := NewBucketService()

	userReq := models.UserCreateRequest{
		Access:    req.Access,
		Secret:    req.Secret,
		Role:      req.Role,
		UserID:    req.UserID,
		GroupID:   req.GroupID,
		ProjectID: req.ProjectID,
	}

	if err := vgwService.CreateUser(userReq); err != nil {
		return summary, fmt.Errorf("failed to create user: %w", err)
	}

	bucketReq := models.BucketCreateRequest{
		Name:  req.Bucket,
		Quota: req.Quota,
		Owner: req.Owner,
	}

	if err := bucketService.CreateBucket(bucketReq); err != nil {
		return summary, fmt.Errorf("failed to create bucket: %w", err)
	}

	if err := vgwService.ChangeBucketOwner(req.Bucket, req.Owner); err != nil {
		return summary, fmt.Errorf("failed to set bucket owner: %w", err)
	}

	summary = ProvisionSummary{
		Access:          req.Access,
		Secret:          req.Secret,
		Role:            req.Role,
		UserID:          req.UserID,
		GroupID:         req.GroupID,
		ProjectID:       req.ProjectID,
		Bucket:          req.Bucket,
		Quota:           req.Quota,
		Owner:           req.Owner,
		SecretGenerated: secretGenerated,
	}

	return summary, nil
}

// GenerateSecretKey generates a random base64-encoded secret key.
func GenerateSecretKey() string {
	b := make([]byte, 48)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
