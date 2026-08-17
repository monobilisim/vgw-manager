package models

// User represents a VersityGW user account
type User struct {
	Access    string `json:"access"`
	Secret    string `json:"secret"`
	Role      string `json:"role"`
	UserID    int    `json:"userID"`
	GroupID   int    `json:"groupID"`
	ProjectID int    `json:"projectID,omitempty"`
}

// UsersJSON represents the structure of users.json
type UsersJSON struct {
	AccessAccounts map[string]User `json:"accessAccounts"`
}

// Bucket represents a ZFS bucket with quota and owner information
type Bucket struct {
	Name       string `json:"name"`
	Mountpoint string `json:"mountpoint"`
	Quota      string `json:"quota"`
	Used       string `json:"used"`
	Available  string `json:"available"`
	Owner      string `json:"owner"`
	Public     bool   `json:"public"`
}

// BucketCreateRequest represents the data needed to create a new bucket
type BucketCreateRequest struct {
	Name       string
	Quota      string
	Owner      string
	Mountpoint string
}

// UserCreateRequest represents the data needed to create a new user
type UserCreateRequest struct {
	Access    string
	Secret    string
	Role      string
	UserID    int
	GroupID   int
	ProjectID int
}

// UserUpdateRequest represents the data needed to update an existing user
type UserUpdateRequest struct {
	Access    string
	Secret    string
	Role      string
	UserID    int
	GroupID   int
	ProjectID int
}
