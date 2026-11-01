package services

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"

	"github.com/monobilisim/vgw-manager/config"
	"github.com/monobilisim/vgw-manager/models"
)

// ErrInvalidInput marks validation failures so the API layer can answer 400
// instead of 500.
var ErrInvalidInput = errors.New("invalid input")

// ErrBucketNotEmpty is returned when a delete would destroy objects.
var ErrBucketNotEmpty = errors.New("bucket is not empty")

// ErrBucketNotFound means no ZFS dataset exists for the bucket.
var ErrBucketNotFound = errors.New("bucket not found")

// metaTmpDir is VersityGW's posix scratch dir; upstream ignores it when deciding emptiness.
const metaTmpDir = ".sgwtmp"

var quotaPattern = regexp.MustCompile(`^[0-9]+(\.[0-9]+)?([KMGTPE](I?B)?)?$`)

func validateQuota(quota string) error {
	if !quotaPattern.MatchString(strings.ToUpper(quota)) {
		return fmt.Errorf("%w: quota must be a size such as 500G or 2T", ErrInvalidInput)
	}
	return nil
}

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

// validateBucketName guards every path that interpolates a name into zfs
// arguments or signed URLs: "/" would escape the pool base, "@" names a
// snapshot, and a leading "-" would be parsed as an option.
func validateBucketName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: bucket name is required", ErrInvalidInput)
	}
	if len(name) < 3 || len(name) > 63 {
		return fmt.Errorf("%w: bucket name must be 3-63 characters", ErrInvalidInput)
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-', r == '.':
		default:
			return fmt.Errorf("%w: bucket name may only contain lowercase letters, digits, dots, and hyphens", ErrInvalidInput)
		}
	}
	if strings.HasPrefix(name, "-") || strings.HasPrefix(name, ".") ||
		strings.HasSuffix(name, ".") || strings.Contains(name, "..") {
		return fmt.Errorf("%w: bucket name may only contain lowercase letters, digits, dots, and hyphens", ErrInvalidInput)
	}
	return nil
}

// validAccessKey reports whether s is safe to interpolate into bucket
// policy JSON: printable, no quotes or backslashes that could break out of
// the string literal.
func validAccessKey(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < 0x20 || r > 0x7e || r == '"' || r == '\\' {
			return false
		}
	}
	return true
}

// CreateBucket creates a new ZFS bucket with specified quota.
// Ownership is handled separately via the change-bucket-owner API.
func (s *BucketService) CreateBucket(req models.BucketCreateRequest) error {
	if err := validateBucketName(req.Name); err != nil {
		return err
	}
	if req.Mountpoint == "" {
		req.Mountpoint = fmt.Sprintf("%s/%s", config.MountBase, req.Name)
	}

	zfsPath := fmt.Sprintf("%s/%s", config.ZFSPoolBase, req.Name)

	args := []string{"create"}
	args = append(args, "-o", fmt.Sprintf("mountpoint=%s", req.Mountpoint))

	if req.Quota != "" {
		if err := validateQuota(req.Quota); err != nil {
			return err
		}
		args = append(args, "-o", fmt.Sprintf("quota=%s", req.Quota))
	}

	args = append(args, "--", zfsPath)

	cmd := exec.Command("zfs", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create ZFS bucket: %w (output: %s)", err, string(output))
	}

	return nil
}

// DeleteBucket destroys an empty, mounted ZFS bucket. Without -r, zfs refuses
// datasets that still have snapshots or children.
func (s *BucketService) DeleteBucket(name string) error {
	if err := validateBucketName(name); err != nil {
		return err
	}
	zfsPath := fmt.Sprintf("%s/%s", config.ZFSPoolBase, name)

	bucket, err := s.GetBucket(name)
	if err != nil {
		return err
	}
	out, err := exec.Command("zfs", "get", "-H", "-o", "value", "mounted", "--", zfsPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to read ZFS mount state: %w (output: %s)", err, string(out))
	}
	// An unmounted dataset shows an empty directory even when it holds data.
	if strings.TrimSpace(string(out)) != "yes" {
		return fmt.Errorf("refusing to delete %s: dataset is not mounted", zfsPath)
	}
	if err := checkBucketDirEmpty(bucket.Mountpoint); err != nil {
		return err
	}

	output, err := exec.Command("zfs", "destroy", "--", zfsPath).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to delete ZFS bucket: %w (output: %s)", err, string(output))
	}

	return nil
}

func checkBucketDirEmpty(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to inspect bucket directory: %w", err)
	}
	for _, e := range entries {
		if e.Name() != metaTmpDir {
			return fmt.Errorf("%w: %s contains %q", ErrBucketNotEmpty, dir, e.Name())
		}
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

	return nil, fmt.Errorf("%w: %s", ErrBucketNotFound, name)
}
