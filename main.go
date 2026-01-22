package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/omertahaoztop/vgw-manager/config"
	"github.com/omertahaoztop/vgw-manager/models"
	"github.com/omertahaoztop/vgw-manager/services"
	"github.com/omertahaoztop/vgw-manager/ui"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && strings.EqualFold(os.Args[1], "update") {
		latestVersion, updated, err := services.SelfUpdate(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
			os.Exit(1)
		}
		if updated {
			fmt.Printf("Updated to %s. Please re-run the command.\n", latestVersion)
		} else {
			fmt.Printf("Already up to date (%s).\n", latestVersion)
		}
		return
	}

	exe := filepath.Base(os.Args[0])
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags]\n\n", exe)
		fmt.Fprintln(flag.CommandLine.Output(), "Operations:")
		fmt.Fprintln(flag.CommandLine.Output(), "  --list-users          List all users and exit")
		fmt.Fprintln(flag.CommandLine.Output(), "  --list-buckets        List all buckets and exit")
		fmt.Fprintln(flag.CommandLine.Output(), "  update               Update the binary to the latest release and exit")
		fmt.Fprintln(flag.CommandLine.Output(), "  --update              Update the binary to the latest release and exit")
		fmt.Fprintln(flag.CommandLine.Output(), "  --version             Print version and exit")
		fmt.Fprintln(flag.CommandLine.Output(), "  --provision           Create user + bucket + set owner without launching the TUI")
		fmt.Fprintln(flag.CommandLine.Output(), "                         (use with --access, --role, --bucket, --quota, optional --secret/--owner/--uid/--gid/--project-id)")
		fmt.Fprintln(flag.CommandLine.Output(), "  --config <path>       Path to YAML config file (default: /etc/vgw-manager.yaml)")
		fmt.Fprintln(flag.CommandLine.Output(), "  (no flags)            Launch the interactive TUI")
		fmt.Fprintln(flag.CommandLine.Output(), "\nFlags:")
		flag.PrintDefaults()
	}

	// Define CLI flags
	configPath := flag.String("config", config.DefaultConfigPath, "Path to the YAML config file")

	// Operations
	listUsers := flag.Bool("list-users", false, "List all users and exit")
	listBuckets := flag.Bool("list-buckets", false, "List all buckets and exit")
	selfUpdate := flag.Bool("update", false, "Update the binary to the latest release and exit")
	showVersion := flag.Bool("version", false, "Print version and exit")
	provisionAll := flag.Bool("provision", false, "Create a user, create a bucket, and set the bucket owner")
	createUser := flag.Bool("create-user", false, "Create a new user")
	createBucket := flag.Bool("create-bucket", false, "Create a new bucket")
	changeOwner := flag.Bool("change-owner", false, "Change bucket owner")
	makePublic := flag.Bool("make-public", false, "Make bucket public")
	makePrivate := flag.Bool("make-private", false, "Make bucket private")
	deleteUser := flag.Bool("delete-user", false, "Delete a user")
	deleteBucket := flag.Bool("delete-bucket", false, "Delete a bucket")

	// Arguments
	accessKey := flag.String("access", "", "Access key (User)")
	secretKey := flag.String("secret", "", "Secret key (User) (auto-generated if empty)")
	role := flag.String("role", "user", "Role (User) (admin, user, or userplus)")
	userID := flag.Int("uid", 0, "User ID (User)")
	groupID := flag.Int("gid", 0, "Group ID (User)")
	projectID := flag.Int("project-id", 0, "Project ID (User)")
	bucketName := flag.String("bucket", "", "Bucket name")
	bucketQuota := flag.String("quota", "", "Quota for the bucket (e.g., 2T, 500G)")
	bucketOwner := flag.String("owner", "", "Bucket owner access key")

	jsonOutput := flag.Bool("json", false, "Output in JSON format")
	flag.Parse()

	if *showVersion {
		fmt.Printf("vgw-manager %s\n", version)
		return
	}

	if *selfUpdate {
		latestVersion, updated, err := services.SelfUpdate(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Self-update failed: %v\n", err)
			os.Exit(1)
		}
		if updated {
			fmt.Printf("Updated to %s. Please re-run the command.\n", latestVersion)
		} else {
			fmt.Printf("Already up to date (%s).\n", latestVersion)
		}
		return
	}

	if err := config.Load(*configPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize Services
	vgwService := services.NewVersityGWService()
	bucketService := services.NewBucketService()

	// Handle Operations

	if *createUser {
		if *accessKey == "" || *secretKey == "" {
			fmt.Fprintln(os.Stderr, "Error: --access and --secret are required for create-user")
			os.Exit(1)
		}
		req := models.UserCreateRequest{
			Access:    *accessKey,
			Secret:    *secretKey,
			Role:      *role,
			UserID:    *userID,
			GroupID:   *groupID,
			ProjectID: *projectID,
		}
		if err := vgwService.CreateUser(req); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User '%s' created successfully.\n", *accessKey)
		return
	}

	if *deleteUser {
		if *accessKey == "" {
			fmt.Fprintln(os.Stderr, "Error: --access is required for delete-user")
			os.Exit(1)
		}
		if err := vgwService.DeleteUser(*accessKey); err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting user: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("User '%s' deleted successfully.\n", *accessKey)
		return
	}

	if *createBucket {
		if *bucketName == "" || *bucketQuota == "" {
			fmt.Fprintln(os.Stderr, "Error: --bucket and --quota are required for create-bucket")
			os.Exit(1)
		}

		owner := *bucketOwner
		if owner == "" {
			fmt.Fprintln(os.Stderr, "Warning: No owner specified for bucket, using 'root' or creating without explicit owner change.")
		}

		req := models.BucketCreateRequest{
			Name:  *bucketName,
			Quota: *bucketQuota,
			Owner: owner,
		}

		// Create ZFS dataset
		if err := bucketService.CreateBucket(req); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating ZFS bucket: %v\n", err)
			os.Exit(1)
		}

		// Set Owner if specified
		if owner != "" {
			if err := vgwService.ChangeBucketOwner(*bucketName, owner); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Bucket created but failed to set owner: %v\n", err)
			} else {
				fmt.Printf("Bucket '%s' created with owner '%s'.\n", *bucketName, owner)
				return
			}
		}
		fmt.Printf("Bucket '%s' created.\n", *bucketName)
		return
	}

	if *deleteBucket {
		if *bucketName == "" {
			fmt.Fprintln(os.Stderr, "Error: --bucket is required for delete-bucket")
			os.Exit(1)
		}

		// Try ZFS delete first
		err := bucketService.DeleteBucket(*bucketName)
		if err != nil {
			// Check if it's an API-only bucket (ZFS dataset missing)
			// A simple string check or error type check would be ideal, but for now we fallback
			fmt.Fprintf(os.Stderr, "ZFS delete failed (%v), attempting API delete...\n", err)

			if apiErr := vgwService.DeleteBucket(*bucketName); apiErr != nil {
				fmt.Fprintf(os.Stderr, "Error deleting bucket (API): %v\n", apiErr)
				os.Exit(1)
			}
			fmt.Printf("Bucket '%s' deleted (via API).\n", *bucketName)
			return
		}

		fmt.Printf("Bucket '%s' deleted (via ZFS).\n", *bucketName)
		return
	}

	if *changeOwner {
		if *bucketName == "" || *bucketOwner == "" {
			fmt.Fprintln(os.Stderr, "Error: --bucket and --owner are required for change-owner")
			os.Exit(1)
		}
		if err := vgwService.ChangeBucketOwner(*bucketName, *bucketOwner); err != nil {
			fmt.Fprintf(os.Stderr, "Error changing owner: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Owner of bucket '%s' changed to '%s'.\n", *bucketName, *bucketOwner)
		return
	}

	if *makePublic {
		if *bucketName == "" {
			fmt.Fprintln(os.Stderr, "Error: --bucket is required for make-public")
			os.Exit(1)
		}

		// We need an owner for the policy. Try to fetch it.
		owner := *bucketOwner
		if owner == "" {
			var err error
			owner, err = vgwService.GetBucketOwner(*bucketName)
			if err != nil || owner == "" {
				fmt.Fprintf(os.Stderr, "Error resolving owner for policy generation. Please specify --owner explicitly.\n")
				os.Exit(1)
			}
		}

		policy := services.GeneratePublicPolicy(*bucketName, owner)
		if err := vgwService.SetBucketPolicy(*bucketName, policy); err != nil {
			fmt.Fprintf(os.Stderr, "Error making bucket public: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Bucket '%s' is now PUBLIC.\n", *bucketName)
		return
	}

	if *makePrivate {
		if *bucketName == "" {
			fmt.Fprintln(os.Stderr, "Error: --bucket is required for make-private")
			os.Exit(1)
		}
		if err := vgwService.DeleteBucketPolicy(*bucketName); err != nil {
			fmt.Fprintf(os.Stderr, "Error removing policy: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Bucket '%s' is now PRIVATE (policy removed).\n", *bucketName)
		return
	}

	if *provisionAll {
		cfg := provisionConfig{
			Access:    *accessKey,
			Secret:    *secretKey,
			Role:      *role,
			UserID:    *userID,
			GroupID:   *groupID,
			ProjectID: *projectID,
			Bucket:    *bucketName,
			Quota:     *bucketQuota,
			Owner:     *bucketOwner,
		}

		summary, err := runProvision(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error provisioning user/bucket: %v\n", err)
			os.Exit(1)
		}

		if *jsonOutput {
			data, _ := json.MarshalIndent(summary, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("User '%s' created with role '%s'\n", summary.Access, summary.Role)
			fmt.Printf("Secret key: %s\n", summary.Secret)
			fmt.Printf("Bucket '%s' created with quota %s and owner '%s'\n", summary.Bucket, summary.Quota, summary.Owner)
			if summary.SecretGenerated {
				fmt.Println("(Secret key was auto-generated)")
			}
		}
		return
	}

	// Handle CLI flags
	if *listUsers {
		userService := services.NewUserService()
		users, err := userService.ListUsers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing users: %v\n", err)
			os.Exit(1)
		}

		if *jsonOutput {
			data, _ := json.MarshalIndent(users, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("%-30s %-45s %-15s\n", "ACCESS KEY", "SECRET KEY", "ROLE")
			fmt.Println("────────────────────────────────────────────────────────────────────────────────────────────")
			for _, user := range users {
				fmt.Printf("%-30s %-45s %-15s\n", user.Access, user.Secret, user.Role)
			}
		}
		return
	}

	if *listBuckets {
		bucketService := services.NewBucketService()
		buckets, err := bucketService.ListBuckets()
		// If ZFS list fails, we still might be able to get API buckets
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Error listing ZFS buckets: %v\n", err)
			buckets = []models.Bucket{}
		}

		// Try to get API buckets
		versityGWService := services.NewVersityGWService()
		apiBuckets, apiErr := versityGWService.ListBuckets()
		if apiErr == nil {
			// Create map for fast lookup of ZFS buckets
			zfsMap := make(map[string]*models.Bucket)
			for i := range buckets {
				zfsMap[buckets[i].Name] = &buckets[i]
			}

			// Create map for API info
			apiMap := make(map[string]services.BucketInfo)
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
				trueOwner, err := versityGWService.GetBucketOwner(buckets[i].Name)
				if err == nil && trueOwner != "" {
					buckets[i].Owner = trueOwner
				}
			}

			// Add buckets that exist in API but NOT in ZFS
			for _, apiBucket := range apiBuckets {
				if _, exists := zfsMap[apiBucket.Name]; !exists {
					// Create placeholder bucket
					newBucket := models.Bucket{
						Name:       apiBucket.Name,
						Mountpoint: "-",
						Quota:      "-",
						Used:       "-",
						Available:  "-",
						Owner:      apiBucket.Owner,
					}

					// Try to fetch true owner
					trueOwner, err := versityGWService.GetBucketOwner(apiBucket.Name)
					if err == nil && trueOwner != "" {
						newBucket.Owner = trueOwner
					}

					buckets = append(buckets, newBucket)
				}
			}
		} else if err != nil {
			// If both failed, then we have a real error
			fmt.Fprintf(os.Stderr, "Error listing buckets: ZFS(%v) API(%v)\n", err, apiErr)
			os.Exit(1)
		}

		// Sort merged list
		sort.Slice(buckets, func(i, j int) bool {
			return buckets[i].Name < buckets[j].Name
		})

		if *jsonOutput {
			data, _ := json.MarshalIndent(buckets, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("%-30s %-20s %-15s %-15s %-15s\n", "NAME", "OWNER", "QUOTA", "USED", "AVAILABLE")
			fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────────────")
			for _, bucket := range buckets {
				fmt.Printf("%-30s %-20s %-15s %-15s %-15s\n",
					bucket.Name, bucket.Owner, bucket.Quota, bucket.Used, bucket.Available)
			}
		}
		return
	}

	// Run TUI if no flags specified
	p := tea.NewProgram(ui.NewModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running application: %v\n", err)
		os.Exit(1)
	}
}

type provisionConfig struct {
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

type provisionSummary struct {
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

func runProvision(cfg provisionConfig) (provisionSummary, error) {
	summary := provisionSummary{}

	if cfg.Access == "" {
		return summary, fmt.Errorf("access key is required (use --access)")
	}
	if cfg.Role != "admin" && cfg.Role != "user" && cfg.Role != "userplus" {
		return summary, fmt.Errorf("role must be admin, user, or userplus")
	}
	if cfg.Bucket == "" {
		return summary, fmt.Errorf("bucket name is required (use --bucket)")
	}
	if cfg.Quota == "" {
		return summary, fmt.Errorf("quota is required (use --quota, e.g. 2T)")
	}
	if cfg.Owner == "" {
		cfg.Owner = cfg.Access
	}

	secretGenerated := false
	if cfg.Secret == "" {
		cfg.Secret = generateSecretKey()
		secretGenerated = true
	}

	vgwService := services.NewVersityGWService()
	bucketService := services.NewBucketService()

	userReq := models.UserCreateRequest{
		Access:    cfg.Access,
		Secret:    cfg.Secret,
		Role:      cfg.Role,
		UserID:    cfg.UserID,
		GroupID:   cfg.GroupID,
		ProjectID: cfg.ProjectID,
	}

	if err := vgwService.CreateUser(userReq); err != nil {
		return summary, fmt.Errorf("failed to create user: %w", err)
	}

	bucketReq := models.BucketCreateRequest{
		Name:  cfg.Bucket,
		Quota: cfg.Quota,
		Owner: cfg.Owner,
	}

	if err := bucketService.CreateBucket(bucketReq); err != nil {
		return summary, fmt.Errorf("failed to create bucket: %w", err)
	}

	if err := vgwService.ChangeBucketOwner(cfg.Bucket, cfg.Owner); err != nil {
		return summary, fmt.Errorf("failed to set bucket owner: %w", err)
	}

	summary = provisionSummary{
		Access:          cfg.Access,
		Secret:          cfg.Secret,
		Role:            cfg.Role,
		UserID:          cfg.UserID,
		GroupID:         cfg.GroupID,
		ProjectID:       cfg.ProjectID,
		Bucket:          cfg.Bucket,
		Quota:           cfg.Quota,
		Owner:           cfg.Owner,
		SecretGenerated: secretGenerated,
	}

	return summary, nil
}

func generateSecretKey() string {
	b := make([]byte, 48)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
