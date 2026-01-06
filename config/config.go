package config

import (
	"fmt"
	"net/url"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds runtime configuration values.
type Config struct {
	AdminAccess   string `json:"adminAccess" yaml:"adminAccess"`
	AdminSecret   string `json:"adminSecret" yaml:"adminSecret"`
	EndpointURL   string `json:"endpointURL" yaml:"endpointURL"`
	Region        string `json:"region" yaml:"region"`
	UsersJSONPath string `json:"usersJSONPath" yaml:"usersJSONPath"`
	ZFSPoolBase   string `json:"zfsPoolBase" yaml:"zfsPoolBase"`
	MountBase     string `json:"mountBase" yaml:"mountBase"`
}

var (
	// DefaultConfigPath is the default file path used when no override is provided.
	DefaultConfigPath = "/etc/vgw-manager.yaml"

	// Defaults used when no file/env is provided.
	defaultConfig = Config{
		AdminAccess:   "changeme-access",
		AdminSecret:   "changeme-secret",
		EndpointURL:   "http://localhost:7070",
		Region:        "local",
		UsersJSONPath: "/tank/s3/accounts/users.json",
		ZFSPoolBase:   "tank/s3/buckets",
		MountBase:     "/tank/s3/buckets",
	}

	// Exported values used across the app (populated in init).
	AdminAccess   string
	AdminSecret   string
	EndpointURL   string
	Region        string
	UsersJSONPath string
	ZFSPoolBase   string
	MountBase     string
)

func init() {
	// Load defaults without failing hard if config is absent.
	_ = Load("")
}

// Load applies configuration from the provided path, environment variables, and defaults.
// The search order for the file is:
//  1. configPath argument (if non-empty)
//  2. VGW_CONFIG_PATH environment variable
//  3. DefaultConfigPath ("/etc/vgw-manager.yaml")
//
// Environment variables still override values loaded from the file.
// If the file cannot be read and the path came from the flag/env, the error is returned.
func Load(configPath string) error {
	cfg := defaultConfig

	resolvedPath := resolvePath(configPath)

	loadedCfg, err := loadFromFile(resolvedPath, cfg)
	if err == nil {
		cfg = loadedCfg
	} else if configPath != "" || os.Getenv("VGW_CONFIG_PATH") != "" {
		return fmt.Errorf("failed to load config file %s: %w", resolvedPath, err)
	}

	cfg = applyEnv(cfg)

	// Validate the final configuration.
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Expose package-level vars for callers.
	AdminAccess = cfg.AdminAccess
	AdminSecret = cfg.AdminSecret
	EndpointURL = cfg.EndpointURL
	Region = cfg.Region
	UsersJSONPath = cfg.UsersJSONPath
	ZFSPoolBase = cfg.ZFSPoolBase
	MountBase = cfg.MountBase

	return nil
}

// Validate checks if the configuration values are valid.
func (c *Config) Validate() error {
	if c.EndpointURL == "" {
		return fmt.Errorf("endpointURL is required")
	}
	if _, err := url.ParseRequestURI(c.EndpointURL); err != nil {
		return fmt.Errorf("invalid endpointURL: %w", err)
	}
	if c.Region == "" {
		return fmt.Errorf("region is required")
	}
	if c.UsersJSONPath == "" {
		return fmt.Errorf("usersJSONPath is required")
	}
	if c.ZFSPoolBase == "" {
		return fmt.Errorf("zfsPoolBase is required")
	}
	if c.MountBase == "" {
		return fmt.Errorf("mountBase is required")
	}
	return nil
}

func resolvePath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}

	if envPath := os.Getenv("VGW_CONFIG_PATH"); envPath != "" {
		return envPath
	}

	return DefaultConfigPath
}

func loadFromFile(path string, base Config) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return base, err
	}

	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return base, err
	}

	// Merge: non-empty fields override base.
	if fileCfg.AdminAccess != "" {
		base.AdminAccess = fileCfg.AdminAccess
	}
	if fileCfg.AdminSecret != "" {
		base.AdminSecret = fileCfg.AdminSecret
	}
	if fileCfg.EndpointURL != "" {
		base.EndpointURL = fileCfg.EndpointURL
	}
	if fileCfg.Region != "" {
		base.Region = fileCfg.Region
	}
	if fileCfg.UsersJSONPath != "" {
		base.UsersJSONPath = fileCfg.UsersJSONPath
	}
	if fileCfg.ZFSPoolBase != "" {
		base.ZFSPoolBase = fileCfg.ZFSPoolBase
	}
	if fileCfg.MountBase != "" {
		base.MountBase = fileCfg.MountBase
	}

	return base, nil
}

func applyEnv(base Config) Config {
	if v := os.Getenv("VGW_ADMIN_ACCESS"); v != "" {
		base.AdminAccess = v
	}
	if v := os.Getenv("VGW_ADMIN_SECRET"); v != "" {
		base.AdminSecret = v
	}
	if v := os.Getenv("VGW_ENDPOINT_URL"); v != "" {
		base.EndpointURL = v
	}
	if v := os.Getenv("VGW_REGION"); v != "" {
		base.Region = v
	}
	if v := os.Getenv("VGW_USERS_JSON_PATH"); v != "" {
		base.UsersJSONPath = v
	}
	if v := os.Getenv("VGW_ZFS_POOL_BASE"); v != "" {
		base.ZFSPoolBase = v
	}
	if v := os.Getenv("VGW_MOUNT_BASE"); v != "" {
		base.MountBase = v
	}
	return base
}
