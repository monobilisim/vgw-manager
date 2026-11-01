package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/monobilisim/vgw-manager/models"
)

// ListMergedBuckets returns a merged list of ZFS and API buckets, sorted by name.
// ZFS buckets are enriched with real owner info from the VersityGW API (via ACL).
// Buckets that exist only in the API are added as placeholders.
// Returns an error when both listings fail or a bucket policy cannot be read.
func ListMergedBuckets() ([]models.Bucket, error) {
	bucketService := NewBucketService()
	vgwService := NewVersityGWService()

	buckets, zfsErr := bucketService.ListBuckets()
	if zfsErr != nil {
		buckets = []models.Bucket{}
	}

	apiBuckets, apiErr := vgwService.ListBuckets()
	if apiErr == nil {
		// Create map for fast lookup of ZFS buckets
		zfsMap := make(map[string]*models.Bucket)
		for i := range buckets {
			zfsMap[buckets[i].Name] = &buckets[i]
		}

		// Create map for API info
		apiMap := make(map[string]BucketInfo)
		for _, b := range apiBuckets {
			apiMap[b.Name] = b
		}

		// Update existing ZFS buckets with owner info
		for i := range buckets {
			if info, ok := apiMap[buckets[i].Name]; ok {
				if info.Owner != "" {
					buckets[i].Owner = info.Owner
				}
			}

			// Fetch TRUE owner from ACL
			trueOwner, err := vgwService.GetBucketOwner(buckets[i].Name)
			if err == nil && trueOwner != "" {
				buckets[i].Owner = trueOwner
			}

			buckets[i].Public, err = bucketIsPublic(vgwService, buckets[i].Name)
			if err != nil {
				return nil, fmt.Errorf("checking bucket %q policy: %w", buckets[i].Name, err)
			}
		}

		// Add buckets that exist in API but NOT in ZFS
		for _, apiBucket := range apiBuckets {
			if _, exists := zfsMap[apiBucket.Name]; !exists {
				newBucket := models.Bucket{
					Name:       apiBucket.Name,
					Mountpoint: "-",
					Quota:      "-",
					Used:       "-",
					Available:  "-",
					Owner:      apiBucket.Owner,
				}

				trueOwner, err := vgwService.GetBucketOwner(apiBucket.Name)
				if err == nil && trueOwner != "" {
					newBucket.Owner = trueOwner
				}

				newBucket.Public, err = bucketIsPublic(vgwService, apiBucket.Name)
				if err != nil {
					return nil, fmt.Errorf("checking bucket %q policy: %w", apiBucket.Name, err)
				}

				buckets = append(buckets, newBucket)
			}
		}
	} else if zfsErr != nil {
		// If both failed, then we have a real error
		return nil, fmt.Errorf("listing buckets: ZFS(%v) API(%v)", zfsErr, apiErr)
	}

	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Name < buckets[j].Name
	})

	return buckets, nil
}

type bucketPolicy struct {
	Statements []bucketPolicyStatement `json:"Statement"`
}

type bucketPolicyStatement struct {
	Principal json.RawMessage `json:"Principal"`
}

// bucketIsPublic reports whether a bucket policy grants access to everyone.
// A missing policy or a policy limited to named principals is private.
func bucketIsPublic(vgwService *VersityGWService, name string) (bool, error) {
	policy, err := vgwService.GetBucketPolicy(name)
	if err == nil {
		return policyGrantsPublicAccess(policy)
	}
	if IsNotFound(err) {
		return false, nil
	}
	return false, err
}

func policyGrantsPublicAccess(policy string) (bool, error) {
	var document bucketPolicy
	if err := json.Unmarshal([]byte(policy), &document); err != nil {
		return false, fmt.Errorf("parsing bucket policy: %w", err)
	}

	for _, statement := range document.Statements {
		var principal string
		if json.Unmarshal(statement.Principal, &principal) == nil && principal == "*" {
			return true, nil
		}

		var principals map[string]json.RawMessage
		if json.Unmarshal(statement.Principal, &principals) == nil {
			for _, value := range principals {
				if json.Unmarshal(value, &principal) == nil && principal == "*" {
					return true, nil
				}

				var values []string
				if json.Unmarshal(value, &values) == nil {
					for _, value := range values {
						if value == "*" {
							return true, nil
						}
					}
				}
			}
		}

		var principalList []string
		if json.Unmarshal(statement.Principal, &principalList) == nil {
			for _, principal := range principalList {
				if principal == "*" {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// RemoveBucket deletes an empty bucket and returns where it lived ("zfs" or
// "api"). Non-empty buckets fail with ErrBucketNotEmpty and are left intact.
func RemoveBucket(name string) (string, error) {
	if err := validateBucketName(name); err != nil {
		return "", err
	}
	bucketService := NewBucketService()
	vgwService := NewVersityGWService()

	if _, err := bucketService.GetBucket(name); err != nil {
		if !errors.Is(err, ErrBucketNotFound) {
			return "", err
		}
		// Not a dataset here: VersityGW itself refuses non-empty buckets with 409.
		if err := vgwService.DeleteBucket(name); err != nil {
			if hasStatus(err, http.StatusConflict) {
				return "", fmt.Errorf("%w: %w", ErrBucketNotEmpty, err)
			}
			return "", err
		}
		return "api", nil
	}

	if err := bucketService.DeleteBucket(name); err != nil {
		return "", err
	}
	if err := vgwService.DeleteBucket(name); err != nil && !IsNotFound(err) {
		return "", fmt.Errorf("ZFS dataset destroyed but API delete failed: %w", err)
	}
	return "zfs", nil
}

// CreateBucketWithOwner creates the ZFS dataset and assigns its owner; the
// dataset is destroyed again when the owner cannot be set.
func CreateBucketWithOwner(req models.BucketCreateRequest) error {
	if req.Owner != "" && !validAccessKey(req.Owner) {
		return fmt.Errorf("%w: invalid owner access key", ErrInvalidInput)
	}
	bucketService := NewBucketService()
	if err := bucketService.CreateBucket(req); err != nil {
		return err
	}
	if req.Owner == "" {
		return nil
	}
	if err := NewVersityGWService().ChangeBucketOwner(req.Name, req.Owner); err != nil {
		if rbErr := bucketService.DeleteBucket(req.Name); rbErr != nil {
			return fmt.Errorf("failed to set bucket owner: %w (rollback failed, dataset %s left in place: %v)", err, req.Name, rbErr)
		}
		return fmt.Errorf("failed to set bucket owner, bucket creation rolled back: %w", err)
	}
	return nil
}

// MakeBucketPublic resolves the bucket owner (if empty), generates a public
// read policy, and applies it via the VersityGW API.
func MakeBucketPublic(name, owner string) error {
	if err := validateBucketName(name); err != nil {
		return err
	}

	vgwService := NewVersityGWService()

	if owner == "" {
		var err error
		owner, err = vgwService.GetBucketOwner(name)
		if err != nil || owner == "" {
			return fmt.Errorf("failed to resolve owner for policy generation")
		}
	}

	policy, err := GeneratePublicPolicy(name, owner)
	if err != nil {
		return err
	}
	if err := vgwService.SetBucketPolicy(name, policy); err != nil {
		return fmt.Errorf("failed to make bucket public: %w", err)
	}
	return nil
}
