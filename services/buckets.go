package services

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/omertahaoztop/vgw-manager/config"
	"github.com/omertahaoztop/vgw-manager/models"
)

// BucketService handles bucket-related operations
type BucketService struct{}

// NewBucketService creates a new BucketService instance
func NewBucketService() *BucketService {
	return &BucketService{}
}

// ListBuckets returns all ZFS buckets with their properties
func (s *BucketService) ListBuckets() ([]models.Bucket, error) {
	cmd := exec.Command("zfs", "list", "-H", "-o", "name,mountpoint,quota,used,avail", "-t", "filesystem", "-r", config.ZFSPoolBase)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list ZFS filesystems: %w (output: %s)", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	buckets := make([]models.Bucket, 0)

	for _, line := range lines {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		// Skip the base pool itself
		if fields[0] == config.ZFSPoolBase {
			continue
		}

		// Check prefix
		if !strings.HasPrefix(fields[0], config.ZFSPoolBase+"/") {
			continue
		}

		// Extract bucket name from ZFS path
		name := strings.TrimPrefix(fields[0], config.ZFSPoolBase+"/")

		bucket := models.Bucket{
			Name:       name,
			Mountpoint: fields[1],
			Quota:      fields[2],
			Used:       fields[3],
			Available:  fields[4],
			Owner:      "-", // Don't use filesystem owner (usually root), rely on API
		}

		buckets = append(buckets, bucket)
	}

	sort.Slice(buckets, func(i, j int) bool {
		return buckets[i].Name < buckets[j].Name
	})

	return buckets, nil
}

// CreateBucket creates a new ZFS bucket with specified quota.
// Ownership is handled separately via the change-bucket-owner API.
func (s *BucketService) CreateBucket(req models.BucketCreateRequest) error {
	if req.Mountpoint == "" {
		req.Mountpoint = fmt.Sprintf("%s/%s", config.MountBase, req.Name)
	}

	zfsPath := fmt.Sprintf("%s/%s", config.ZFSPoolBase, req.Name)

	args := []string{"create"}
	args = append(args, "-o", fmt.Sprintf("mountpoint=%s", req.Mountpoint))

	if req.Quota != "" {
		args = append(args, "-o", fmt.Sprintf("quota=%s", req.Quota))
	}

	args = append(args, zfsPath)

	cmd := exec.Command("zfs", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create ZFS bucket: %w (output: %s)", err, string(output))
	}

	return nil
}

// DeleteBucket deletes a ZFS bucket using zfs destroy
func (s *BucketService) DeleteBucket(name string) error {
	zfsPath := fmt.Sprintf("%s/%s", config.ZFSPoolBase, name)

	// zfs destroy -r ensures snapshots/clones are also removed if standard
	cmd := exec.Command("zfs", "destroy", "-r", zfsPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete ZFS bucket: %w (output: %s)", err, string(output))
	}

	return nil
}

// GetBucket returns information about a specific bucket
func (s *BucketService) GetBucket(name string) (*models.Bucket, error) {
	buckets, err := s.ListBuckets()
	if err != nil {
		return nil, err
	}

	for _, bucket := range buckets {
		if bucket.Name == name {
			return &bucket, nil
		}
	}

	return nil, fmt.Errorf("bucket not found: %s", name)
}
