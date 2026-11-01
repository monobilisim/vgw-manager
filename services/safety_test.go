package services

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateQuota(t *testing.T) {
	for _, q := range []string{"2T", "500G", "1.5T", "100GB", "10GiB", "1024", "2t"} {
		if err := validateQuota(q); err != nil {
			t.Errorf("validateQuota(%q) = %v, want nil", q, err)
		}
	}
	for _, q := range []string{"", "none", "-1", "1T,refquota=1", "1T -o x", "T", "1.T", "1X"} {
		if err := validateQuota(q); !errors.Is(err, ErrInvalidInput) {
			t.Errorf("validateQuota(%q) = %v, want ErrInvalidInput", q, err)
		}
	}
}

func TestCheckBucketDirEmpty(t *testing.T) {
	dir := t.TempDir()
	if err := checkBucketDirEmpty(dir); err != nil {
		t.Fatalf("empty dir: %v", err)
	}

	if err := os.Mkdir(filepath.Join(dir, metaTmpDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := checkBucketDirEmpty(dir); err != nil {
		t.Fatalf("dir with only %s: %v", metaTmpDir, err)
	}

	if err := os.WriteFile(filepath.Join(dir, ".hidden"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := checkBucketDirEmpty(dir); !errors.Is(err, ErrBucketNotEmpty) {
		t.Fatalf("dir with object: %v, want ErrBucketNotEmpty", err)
	}

	if err := checkBucketDirEmpty(filepath.Join(dir, "missing")); err == nil || errors.Is(err, ErrBucketNotEmpty) {
		t.Fatalf("missing dir: %v, want a non-nil inspection error", err)
	}
}

func TestChecksumFor(t *testing.T) {
	const sum = "4f0a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8"
	list := []byte(sum + "  vgw-manager_1.2.0_linux_amd64.tar.gz\n" +
		"0000000000000000000000000000000000000000000000000000000000000000  vgw-manager_1.2.0_linux_arm64.tar.gz\n" +
		"abc  short.tar.gz\n")

	got, err := checksumFor(list, "vgw-manager_1.2.0_linux_amd64.tar.gz")
	if err != nil || got != sum {
		t.Fatalf("checksumFor = %q, %v; want %q", got, err, sum)
	}
	if _, err := checksumFor(list, "vgw-manager_1.2.0_darwin_amd64.tar.gz"); err == nil {
		t.Fatal("missing asset: want error")
	}
	if _, err := checksumFor(list, "short.tar.gz"); err == nil {
		t.Fatal("malformed sum: want error")
	}
}
