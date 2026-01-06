package services

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/omertahaoztop/vgw-manager/config"
	"github.com/omertahaoztop/vgw-manager/models"
)

// VersityGWService handles VersityGW admin API operations
type VersityGWService struct {
	client *http.Client
}

// NewVersityGWService creates a new VersityGWService instance
func NewVersityGWService() *VersityGWService {
	return &VersityGWService{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Account represents a VersityGW account (matching the versitygw auth.Account structure)
type Account struct {
	Access    string `xml:"Access"`
	Secret    string `xml:"Secret"`
	Role      string `xml:"Role"`
	UserID    int    `xml:"UserID"`
	GroupID   int    `xml:"GroupID"`
	ProjectID int    `xml:"ProjectID,omitempty"`
}

// CreateUser creates a new user via VersityGW admin API
func (s *VersityGWService) CreateUser(req models.UserCreateRequest) error {
	acc := Account{
		Access:    req.Access,
		Secret:    req.Secret,
		Role:      req.Role,
		UserID:    req.UserID,
		GroupID:   req.GroupID,
		ProjectID: req.ProjectID,
	}

	accxml, err := xml.Marshal(acc)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	url := fmt.Sprintf("%s/create-user", config.EndpointURL)
	httpReq, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(accxml))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if _, err := s.signAndSend(httpReq, accxml); err != nil {
		return err
	}

	return nil
}

// UpdateUser updates an existing user via VersityGW admin API
func (s *VersityGWService) UpdateUser(req models.UserUpdateRequest) error {
	acc := Account{
		Access:    req.Access,
		Secret:    req.Secret,
		Role:      req.Role,
		UserID:    req.UserID,
		GroupID:   req.GroupID,
		ProjectID: req.ProjectID,
	}

	accxml, err := xml.Marshal(acc)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	url := fmt.Sprintf("%s/update-user?access=%s", config.EndpointURL, req.Access)
	httpReq, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(accxml))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if _, err := s.signAndSend(httpReq, accxml); err != nil {
		return err
	}

	return nil
}

// DeleteUser deletes a user via VersityGW admin API
func (s *VersityGWService) DeleteUser(access string) error {
	url := fmt.Sprintf("%s/delete-user?access=%s", config.EndpointURL, access)
	httpReq, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Delete usually doesn't need body, but VGW might check signature. Empty body for signature.
	if _, err := s.signAndSend(httpReq, []byte{}); err != nil {
		return err
	}

	return nil
}

// signAndSend signs the request with AWS V4 signature, sends it, and returns the response body
func (s *VersityGWService) signAndSend(httpReq *http.Request, payload []byte) ([]byte, error) {
	signer := v4.NewSigner()
	hashedPayload := sha256.Sum256(payload)
	hexPayload := hex.EncodeToString(hashedPayload[:])

	httpReq.Header.Set("X-Amz-Content-Sha256", hexPayload)

	err := signer.SignHTTP(httpReq.Context(), aws.Credentials{
		AccessKeyID:     config.AdminAccess,
		SecretAccessKey: config.AdminSecret,
	}, httpReq, hexPayload, "s3", config.Region, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return body, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// ChangeBucketOwner changes the owner of a bucket
func (s *VersityGWService) ChangeBucketOwner(bucket, owner string) error {
	url := fmt.Sprintf("%s/change-bucket-owner/?bucket=%s&owner=%s", config.EndpointURL, bucket, owner)
	httpReq, err := http.NewRequest(http.MethodPatch, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	signer := v4.NewSigner()
	hashedPayload := sha256.Sum256([]byte{})
	hexPayload := hex.EncodeToString(hashedPayload[:])

	httpReq.Header.Set("X-Amz-Content-Sha256", hexPayload)

	err = signer.SignHTTP(httpReq.Context(), aws.Credentials{
		AccessKeyID:     config.AdminAccess,
		SecretAccessKey: config.AdminSecret,
	}, httpReq, hexPayload, "s3", config.Region, time.Now())
	if err != nil {
		return fmt.Errorf("failed to sign request: %w", err)
	}

	resp, err := s.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// SetBucketPolicy sets the S3 bucket policy
func (s *VersityGWService) SetBucketPolicy(bucket string, policy string) error {
	// S3 PutBucketPolicy: PUT /<bucket>?policy
	// Body: policy JSON
	url := fmt.Sprintf("%s/%s?policy", config.EndpointURL, bucket)
	httpReq, err := http.NewRequest(http.MethodPut, url, bytes.NewBufferString(policy))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if _, err := s.signAndSend(httpReq, []byte(policy)); err != nil {
		return fmt.Errorf("failed to set bucket policy: %w", err)
	}

	return nil
}

// GeneratePublicPolicy generates a bucket policy for public read access and owner full access
func GeneratePublicPolicy(bucket, owner string) string {
	return fmt.Sprintf(`{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "PublicRead",
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::%s/*"
        },
        {
            "Sid": "UserWriteDelete",
            "Effect": "Allow",
            "Principal": {
                "AWS": "%s"
            },
            "Action": [
                "s3:ListMultipartUploadParts",
                "s3:PutObject",
                "s3:AbortMultipartUpload",
                "s3:DeleteObject",
                "s3:GetBucketLocation",
                "s3:GetObject",
                "s3:ListBucket",
                "s3:ListBucketMultipartUploads"
            ],
            "Resource": [
                "arn:aws:s3:::%s",
                "arn:aws:s3:::%s/*"
            ]
        }
    ]
}`, bucket, owner, bucket, bucket)
}

// DeleteBucketPolicy deletes the policy of a bucket (making it private)
func (s *VersityGWService) DeleteBucketPolicy(bucket string) error {
	url := fmt.Sprintf("%s/%s?policy", config.EndpointURL, bucket)
	httpReq, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if _, err := s.signAndSend(httpReq, []byte{}); err != nil {
		return fmt.Errorf("failed to delete bucket policy: %w", err)
	}

	return nil
}

// DeleteBucket deletes a bucket (must be empty)
func (s *VersityGWService) DeleteBucket(bucket string) error {
	url := fmt.Sprintf("%s/%s", config.EndpointURL, bucket)
	httpReq, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if _, err := s.signAndSend(httpReq, []byte{}); err != nil {
		return fmt.Errorf("failed to delete bucket: %w", err)
	}

	return nil
}

// GetBucketPolicy retrieves the policy of a bucket
func (s *VersityGWService) GetBucketPolicy(bucket string) (string, error) {
	url := fmt.Sprintf("%s/%s?policy", config.EndpointURL, bucket)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	body, err := s.signAndSend(httpReq, []byte{})
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// BucketInfo represents a bucket in the list response
type BucketInfo struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
	Owner        string `xml:"Owner"`
}

// ListBuckets lists all buckets via VersityGW admin API
func (s *VersityGWService) ListBuckets() ([]BucketInfo, error) {
	url := fmt.Sprintf("%s/", config.EndpointURL)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := s.signAndSend(httpReq, []byte{})
	if err != nil {
		return nil, err
	}

	// Wrapper for standard S3 ListBuckets response
	type ListAllMyBucketsResult struct {
		Buckets struct {
			Bucket []BucketInfo `xml:"Bucket"`
		} `xml:"Buckets"`
	}

	var result ListAllMyBucketsResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list buckets response: %w", err)
	}

	return result.Buckets.Bucket, nil
}

// GetBucketOwner retrieves the true bucket owner by checking the ACL
func (s *VersityGWService) GetBucketOwner(bucket string) (string, error) {
	url := fmt.Sprintf("%s/%s?acl", config.EndpointURL, bucket)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	body, err := s.signAndSend(httpReq, []byte{})
	if err != nil {
		return "", err
	}

	// Parse XML safely
	type Grantee struct {
		ID string `xml:"ID"`
	}
	type Owner struct {
		ID string `xml:"ID"`
	}
	type AccessControlPolicy struct {
		Owner Owner `xml:"Owner"`
	}

	var acl AccessControlPolicy
	if err := xml.Unmarshal(body, &acl); err != nil {
		return "", fmt.Errorf("failed to parse ACL: %w", err)
	}

	return acl.Owner.ID, nil
}
