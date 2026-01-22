package ui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/omertahaoztop/vgw-manager/models"
	"github.com/omertahaoztop/vgw-manager/services"
)

// View represents different screens in the TUI
type View int

const (
	MainMenuView View = iota
	UsersListView
	BucketsListView
	OperationsView
	CreateUserView
	CreateBucketView
	ChangeOwnerView
	UserDetailView
	BucketDetailView
	ProvisionView
	UpdateUserView
	MakeBucketPublicView
	ConfirmView
)

// Model represents the main application state
type Model struct {
	currentView View
	cursor      int
	width       int
	height      int

	// Services
	userService      *services.UserService
	bucketService    *services.BucketService
	versitygwService *services.VersityGWService

	// Data
	users               []models.User
	buckets             []models.Bucket
	selectedUserIndex   int
	selectedBucketIndex int
	returnView          View

	// Cache for session-persistent data

	// UI components
	userFormInputs        []textinput.Model
	bucketFormInputs      []textinput.Model
	changeOwnerFormInputs []textinput.Model
	provisionFormInputs   []textinput.Model
	focusIndex            int

	// Scroll state
	provisionScroll int
	page            int // Current page index (0-based)
	pageSize        int // Items per page

	// Messages
	errorMessage   string
	successMessage string

	// Confirmation
	pendingAction string // e.g., "delete_user", "delete_bucket", "make_public", "make_private"
	pendingTarget string // The name/access key of the item
}

// NewModel creates a new application model
func NewModel() Model {
	return Model{
		currentView:      MainMenuView,
		userService:      services.NewUserService(),
		bucketService:    services.NewBucketService(),
		versitygwService: services.NewVersityGWService(),
		cursor:           0,
		page:             0,
		pageSize:         20,
		returnView:       MainMenuView,
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.currentView == ProvisionView {
			m = m.ensureProvisionVisible(m.provisionLines())
		}
		return m, nil

	case tea.KeyMsg:
		// Handle form views differently
		if m.currentView == CreateUserView {
			return m.updateUserForm(msg)
		}
		if m.currentView == CreateBucketView {
			return m.updateBucketForm(msg)
		}
		if m.currentView == ChangeOwnerView {
			return m.updateChangeOwnerForm(msg)
		}
		if m.currentView == ProvisionView {
			return m.updateProvisionForm(msg)
		}
		if m.currentView == UpdateUserView {
			return m.updateUserForm(msg) // Reuse user form handler
		}
		if m.currentView == MakeBucketPublicView {
			return m.updateMakePublicForm(msg)
		}

		// Clear messages on any key press
		m.errorMessage = ""
		m.successMessage = ""

		switch msg.String() {
		case "ctrl+c", "q":
			if m.currentView == MainMenuView {
				return m, tea.Quit
			}
			// Return to main menu from any other view
			m.currentView = MainMenuView
			m.cursor = 0
			return m, nil

		case "esc":
			if m.currentView == UserDetailView {
				m.currentView = UsersListView
				return m, nil
			}
			if m.currentView == BucketDetailView {
				m.currentView = BucketsListView
				return m, nil
			}
			if m.currentView == ProvisionView {
				m.currentView = OperationsView
				m.cursor = 0
				return m, nil
			}
			if m.currentView != MainMenuView {
				m.currentView = MainMenuView
				m.cursor = 0
			}
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			maxItems := m.getMaxCursorForView() + 1 // Convert max index to count
			// Limit cursor to pageSize or remaining items on current page
			limit := m.pageSize
			remaining := maxItems - (m.page * m.pageSize)
			if remaining < limit {
				limit = remaining
			}

			if m.cursor < limit-1 {
				m.cursor++
			}

		case "left", "h":
			if m.page > 0 {
				m.page--
				m.cursor = 0
			}

		case "right", "l":
			maxItems := m.getMaxCursorForView() + 1
			totalPages := (maxItems + m.pageSize - 1) / m.pageSize
			if m.page < totalPages-1 {
				m.page++
				m.cursor = 0
			}

		case "enter":
			return m.handleEnter()

		case "c":
			// Handle copy in UsersListView and UserDetailView
			if m.currentView == UsersListView && len(m.users) > 0 {
				idx := m.page*m.pageSize + m.cursor
				if idx < len(m.users) {
					// We need to access item at idx.
					// But copyUserCredentials uses m.cursor assumes absolute?
					// Or copyUserCredentials uses m.users[m.cursor]...
					// I need to update copyUserCredentials or update state to set SelectedUserIndex beforehand
					// Or just pass User to copy helper.
					// copyUserCredentials currently likely uses m.users[m.cursor]
					// I should update it to use resolved index.
					// For now, let's update call site if possible or method logic.
					// Let's assume I fix copyUserCredentials later or it needs update.
					// Actually, simplest is to set m.selectedUserIndex = idx then call.
					m.selectedUserIndex = idx
					if err := m.copyUserCredentials(); err != nil {
						m.errorMessage = fmt.Sprintf("Copy failed: %v", err)
					} else {
						m.successMessage = "✓ Credentials copied to clipboard!"
					}
				}
			} else if m.currentView == UserDetailView {
				if err := m.copyUserCredentialsFromDetail(); err != nil {
					m.errorMessage = fmt.Sprintf("Copy failed: %v", err)
				} else {
					m.successMessage = "✓ Credentials copied to clipboard!"
				}
			}

		case "d":
			// Handle Delete (User or Bucket)
			if m.currentView == UsersListView && len(m.users) > 0 {
				idx := m.page*m.pageSize + m.cursor
				if idx < len(m.users) {
					user := m.users[idx]
					m.pendingAction = "delete_user"
					m.pendingTarget = user.Access
					m.returnView = UsersListView
					m.currentView = ConfirmView
				}
			} else if m.currentView == BucketsListView && len(m.buckets) > 0 {
				idx := m.page*m.pageSize + m.cursor
				if idx < len(m.buckets) {
					bucket := m.buckets[idx]
					m.pendingAction = "delete_bucket"
					m.pendingTarget = bucket.Name
					m.returnView = BucketsListView
					m.currentView = ConfirmView
				}
			}

		case "e":
			// Handle Edit User
			if m.currentView == UsersListView && len(m.users) > 0 {
				idx := m.page*m.pageSize + m.cursor
				if idx < len(m.users) {
					user := m.users[idx]
					m.initUpdateUserForm(user)
					m.currentView = UpdateUserView
					m.returnView = UsersListView
				}
			}

		case "p":
			// Handle Make Public (Buckets)
			if m.currentView == BucketsListView && len(m.buckets) > 0 {
				idx := m.page*m.pageSize + m.cursor
				if idx < len(m.buckets) {
					bucket := m.buckets[idx]
					m.pendingAction = "make_public"
					m.pendingTarget = bucket.Name
					m.returnView = BucketsListView
					m.currentView = ConfirmView
				}
			}

		case "P":
			// Handle Make Private (Delete Policy)
			if m.currentView == BucketsListView && len(m.buckets) > 0 {
				idx := m.page*m.pageSize + m.cursor
				if idx < len(m.buckets) {
					bucket := m.buckets[idx]
					m.pendingAction = "make_private"
					m.pendingTarget = bucket.Name
					m.returnView = BucketsListView
					m.currentView = ConfirmView
				}
			}

		case "y", "Y":
			if m.currentView == ConfirmView {
				return m.executeConfirmedAction()
			}

		case "n", "N":
			if m.currentView == ConfirmView {
				m.pendingAction = ""
				m.pendingTarget = ""
				m.currentView = m.returnView
				m.successMessage = "Operation cancelled."
			}

		}
	}

	return m, nil
}

// executeConfirmedAction performs the pending action after confirmation
func (m Model) executeConfirmedAction() (tea.Model, tea.Cmd) {
	switch m.pendingAction {
	case "delete_user":
		if err := m.versitygwService.DeleteUser(m.pendingTarget); err != nil {
			m.errorMessage = fmt.Sprintf("Failed to delete user: %v", err)
		} else {
			m.successMessage = fmt.Sprintf("User '%s' deleted", m.pendingTarget)
			users, err := m.userService.ListUsers()
			if err == nil {
				m.users = users
				if len(m.users) == 0 {
					m.page = 0
					m.cursor = 0
				} else {
					newMaxPage := (len(m.users) - 1) / m.pageSize
					if m.page > newMaxPage {
						m.page = newMaxPage
						m.cursor = 0
					}
					itemsOnPage := m.pageSize
					if m.page == newMaxPage {
						itemsOnPage = len(m.users) - (m.page * m.pageSize)
					}
					if itemsOnPage <= 0 {
						if m.page > 0 {
							m.page--
							m.cursor = 0
						} else {
							m.cursor = 0
						}
					} else if m.cursor >= itemsOnPage {
						m.cursor = itemsOnPage - 1
					}
					if m.cursor < 0 {
						m.cursor = 0
					}
				}
			}
		}

	case "delete_bucket":
		if err := m.bucketService.DeleteBucket(m.pendingTarget); err != nil {
			if strings.Contains(err.Error(), "dataset is busy") || strings.Contains(err.Error(), "not empty") {
				m.errorMessage = fmt.Sprintf("Cannot delete bucket '%s': Bucket is not empty or busy.", m.pendingTarget)
			} else {
				m.errorMessage = fmt.Sprintf("Failed to delete bucket: %v", err)
			}
		} else {
			m.successMessage = fmt.Sprintf("Bucket '%s' deleted.", m.pendingTarget)
			// Reload buckets
			m.currentView = MainMenuView
			m.cursor = 1
			newM, cmd := m.handleEnter()
			model, ok := newM.(Model)
			if ok {
				m = model
				m.successMessage = fmt.Sprintf("Bucket '%s' deleted.", m.pendingTarget)
			}
			m.pendingAction = ""
			m.pendingTarget = ""
			return m, cmd
		}

	case "make_public":
		// Find the bucket to get owner
		var bucket models.Bucket
		found := false
		for _, b := range m.buckets {
			if b.Name == m.pendingTarget {
				bucket = b
				found = true
				break
			}
		}
		if !found {
			m.errorMessage = fmt.Sprintf("Bucket '%s' not found.", m.pendingTarget)
			break
		}

		// Check if policy already exists
		_, err := m.versitygwService.GetBucketPolicy(bucket.Name)
		if err == nil {
			m.errorMessage = fmt.Sprintf("Bucket '%s' is already PUBLIC (or has policy). Use 'P' to make private.", bucket.Name)
		} else if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "NoSuchBucketPolicy") {
			m.errorMessage = fmt.Sprintf("Failed to check policy status: %v", err)
		} else {
			owner := bucket.Owner
			if owner == "" || owner == "unknown" || owner == "root" {
				owner = bucket.Name
			}
			policy := services.GeneratePublicPolicy(bucket.Name, owner)
			if err := m.versitygwService.SetBucketPolicy(bucket.Name, policy); err != nil {
				m.initMakePublicForm()
				m.bucketFormInputs[0].SetValue(bucket.Name)
				m.bucketFormInputs[1].SetValue(owner)
				m.bucketFormInputs[1].Focus()
				m.focusIndex = 1
				m.currentView = MakeBucketPublicView
				m.returnView = BucketsListView
				errMsg := fmt.Sprintf("%v", err)
				if len(errMsg) > 50 {
					errMsg = errMsg[:47] + "..."
				}
				m.errorMessage = "Auto-public failed: " + errMsg
				m.pendingAction = ""
				m.pendingTarget = ""
				return m, nil
			} else {
				m.successMessage = fmt.Sprintf("Bucket '%s' is now PUBLIC!", bucket.Name)
			}
		}

	case "make_private":
		_, err := m.versitygwService.GetBucketPolicy(m.pendingTarget)
		if err != nil {
			if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "NoSuchBucketPolicy") {
				m.successMessage = fmt.Sprintf("Bucket '%s' is already PRIVATE (no policy).", m.pendingTarget)
			} else {
				m.errorMessage = fmt.Sprintf("Failed to check policy: %v", err)
			}
		} else {
			if err := m.versitygwService.DeleteBucketPolicy(m.pendingTarget); err != nil {
				m.errorMessage = fmt.Sprintf("Failed to make private: %v", err)
			} else {
				m.successMessage = fmt.Sprintf("Bucket '%s' is now PRIVATE (policy removed).", m.pendingTarget)
			}
		}
	}

	m.pendingAction = ""
	m.pendingTarget = ""
	m.currentView = m.returnView
	return m, nil
}

// handleEnter handles enter key press based on current view
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.currentView {
	case MainMenuView:
		switch m.cursor {
		case 0: // List Users
			users, err := m.userService.ListUsers()
			if err != nil {
				m.errorMessage = fmt.Sprintf("Error loading users: %v", err)
			} else {
				m.users = users
				m.currentView = UsersListView
				m.cursor = 0
				m.page = 0
			}

		case 1: // List Buckets
			// 1. Get ZFS buckets (source of truth for quotas/usage)
			buckets, err := m.bucketService.ListBuckets()
			if err != nil {
				m.errorMessage = fmt.Sprintf("Error loading buckets from ZFS: %v", err)
				// Even if ZFS fails, we want to try showing API buckets
				buckets = []models.Bucket{}
			}

			// 2. Try to get API buckets (source of truth for owners and missing ZFS buckets)
			apiBuckets, apiErr := m.versitygwService.ListBuckets()
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

					// Fetch TRUE owner from ACL separately if needed (preserving existing logic)
					trueOwner, err := m.versitygwService.GetBucketOwner(buckets[i].Name)
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

						// Try to fetch true owner for these as well
						trueOwner, err := m.versitygwService.GetBucketOwner(apiBucket.Name)
						if err == nil && trueOwner != "" {
							newBucket.Owner = trueOwner
						}

						buckets = append(buckets, newBucket)
					}
				}
			} else if err != nil {
				// If both failed, show error
				m.errorMessage = fmt.Sprintf("Error loading buckets: ZFS(%v) API(%v)", err, apiErr)
			}

			// Sort merged list
			sort.Slice(buckets, func(i, j int) bool {
				return buckets[i].Name < buckets[j].Name
			})

			m.buckets = buckets
			m.currentView = BucketsListView
			m.cursor = 0
			m.page = 0

		case 2: // Operations
			m.currentView = OperationsView
			m.cursor = 0

		case 3: // Quit
			return m, tea.Quit
		}
	case OperationsView:
		switch m.cursor {
		case 0: // Create User
			m.initUserForm()
			m.returnView = OperationsView
			m.currentView = CreateUserView
			m.cursor = 0
		case 1: // Create Bucket
			m.initBucketForm()
			m.returnView = OperationsView
			m.currentView = CreateBucketView
			m.cursor = 0
		case 2: // Change Bucket Owner
			m.initChangeOwnerForm()
			m.returnView = OperationsView
			m.currentView = ChangeOwnerView
			m.cursor = 0
		case 3: // Make Bucket Public
			m.initMakePublicForm()
			m.returnView = OperationsView
			m.currentView = MakeBucketPublicView
			m.cursor = 0
		case 4: // Provision (user + bucket)
			m.initProvisionForm()
			m.returnView = OperationsView
			m.currentView = ProvisionView
			m.cursor = 0
		}

	case UsersListView:
		// Navigate to user detail view
		idx := m.page*m.pageSize + m.cursor
		if len(m.users) > 0 && idx >= 0 && idx < len(m.users) {
			m.selectedUserIndex = idx
			m.currentView = UserDetailView
		}

	case BucketsListView:
		// Navigate to bucket detail view
		idx := m.page*m.pageSize + m.cursor
		if len(m.buckets) > 0 && idx >= 0 && idx < len(m.buckets) {
			m.selectedBucketIndex = idx
			m.currentView = BucketDetailView
		}

	case CreateUserView, UpdateUserView:
		return m.handleCreateUser()

	case CreateBucketView:
		return m.handleCreateBucket()

	case MakeBucketPublicView:
		return m.handleMakePublic()
	}

	return m, nil
}

// View renders the current view
func (m Model) View() string {
	switch m.currentView {
	case MainMenuView:
		return m.renderMainMenu()
	case UsersListView:
		return m.renderUsersList()
	case BucketsListView:
		return m.renderBucketsList()
	case OperationsView:
		return m.renderOperationsMenu()
	case CreateUserView:
		return m.renderCreateUserForm()
	case CreateBucketView:
		return m.renderCreateBucketForm()
	case ChangeOwnerView:
		return m.renderChangeOwnerForm()
	case UserDetailView:
		return m.renderUserDetail()
	case BucketDetailView:
		return m.renderBucketDetail()
	case ProvisionView:
		return m.renderProvisionForm()
	case UpdateUserView:
		return m.renderCreateUserForm() // Reuse create form style for now
	case MakeBucketPublicView:
		return m.renderMakePublicForm()
	case ConfirmView:
		return m.renderConfirmView()
	default:
		return "Unknown view"
	}
}
