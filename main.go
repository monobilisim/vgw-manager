package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/omertahaoztop/vgw-manager/api"
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
		fmt.Fprintln(flag.CommandLine.Output(), "  --serve                Start HTTP API server instead of TUI")
		fmt.Fprintln(flag.CommandLine.Output(), "  --listen <addr>        Listen address for API server (default: 127.0.0.1:8080)")
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
	serve := flag.Bool("serve", false, "Start HTTP API server instead of TUI")
	listenAddr := flag.String("listen", "", "Listen address for API server")
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

	if *serve {
		if config.APIToken == "" {
			fmt.Fprintln(os.Stderr, "Error: apiToken is required when serving the API (set apiToken in config or VGW_API_TOKEN env). This service runs as root and writes to ZFS.")
			os.Exit(1)
		}

		addr := *listenAddr
		if addr == "" {
			addr = config.APIListen
		}
		if addr == "" {
			addr = "127.0.0.1:8080"
		}

		srv := api.NewServer(version)
		srv.Addr = addr

		go func() {
			slog.Info("API server listening", "addr", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				slog.Error("API server failed", "error", err)
				os.Exit(1)
			}
		}()

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer stop()
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("API server shutdown failed", "error", err)
			os.Exit(1)
		}
		slog.Info("API server stopped")
		return
	}

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

		via, err := services.DeleteBucketWithFallback(*bucketName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error deleting bucket: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Bucket '%s' deleted (via %s).\n", *bucketName, via)
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

		if err := services.MakeBucketPublic(*bucketName, *bucketOwner); err != nil {
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
		req := services.ProvisionRequest{
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

		summary, err := services.Provision(req)
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
		buckets, err := services.ListMergedBuckets()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing buckets: %v\n", err)
			os.Exit(1)
		}

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
