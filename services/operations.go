package services

import (
	"fmt"
	"sort"
	"strings"

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

// bucketIsPublic reports whether a bucket has a policy. A missing policy is
// the expected private-bucket state; other errors must not be hidden.
func bucketIsPublic(vgwService *VersityGWService, name string) (bool, error) {
	_, err := vgwService.GetBucketPolicy(name)
	if err == nil {
		return true, nil
	}
	if strings.Contains(err.Error(), "API error (status 404)") {
		return false, nil
	}
	return false, err
}

// DeleteBucketWithFallback tries ZFS deletion first, then falls back to API deletion.
// Returns the method used ("zfs" or "api") and any error.
func DeleteBucketWithFallback(name string) (string, error) {
	bucketService := NewBucketService()
	err := bucketService.DeleteBucket(name)
	if err == nil {
		return "zfs", nil
	}

	// ZFS failed, try API
	vgwService := NewVersityGWService()
	if apiErr := vgwService.DeleteBucket(name); apiErr != nil {
		return "", fmt.Errorf("ZFS: %v; API: %v", err, apiErr)
	}
	return "api", nil
}

// MakeBucketPublic resolves the bucket owner (if empty), generates a public
// read policy, and applies it via the VersityGW API.
func MakeBucketPublic(name, owner string) error {
	vgwService := NewVersityGWService()

	if owner == "" {
		var err error
		owner, err = vgwService.GetBucketOwner(name)
		if err != nil || owner == "" {
			return fmt.Errorf("failed to resolve owner for policy generation")
		}
	}

	policy := GeneratePublicPolicy(name, owner)
	if err := vgwService.SetBucketPolicy(name, policy); err != nil {
		return fmt.Errorf("failed to make bucket public: %w", err)
	}
	return nil
}
