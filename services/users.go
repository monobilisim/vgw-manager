package services

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/omertahaoztop/vgw-manager/config"
	"github.com/omertahaoztop/vgw-manager/models"
)

// UserService handles user-related operations
type UserService struct{}

// NewUserService creates a new UserService instance
func NewUserService() *UserService {
	return &UserService{}
}

// ListUsers reads and returns all users from users.json
func (s *UserService) ListUsers() ([]models.User, error) {
	data, err := os.ReadFile(config.UsersJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read users.json: %w", err)
	}

	var usersJSON models.UsersJSON
	if err := json.Unmarshal(data, &usersJSON); err != nil {
		return nil, fmt.Errorf("failed to parse users.json: %w", err)
	}

	users := make([]models.User, 0, len(usersJSON.AccessAccounts))
	for _, user := range usersJSON.AccessAccounts {
		users = append(users, user)
	}

	sort.Slice(users, func(i, j int) bool {
		return users[i].Access < users[j].Access
	})

	return users, nil
}

// GetUser returns a specific user by access key
func (s *UserService) GetUser(access string) (*models.User, error) {
	data, err := os.ReadFile(config.UsersJSONPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read users.json: %w", err)
	}

	var usersJSON models.UsersJSON
	if err := json.Unmarshal(data, &usersJSON); err != nil {
		return nil, fmt.Errorf("failed to parse users.json: %w", err)
	}

	user, ok := usersJSON.AccessAccounts[access]
	if !ok {
		return nil, fmt.Errorf("user not found: %s", access)
	}

	return &user, nil
}
