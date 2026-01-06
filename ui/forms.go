package ui

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/omertahaoztop/vgw-manager/models"
	"github.com/omertahaoztop/vgw-manager/services"
)

// initUserForm initializes the user creation form
func (m *Model) initUserForm() {
	inputs := make([]textinput.Model, 3)

	// Access Key
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Access Key (username)"
	inputs[0].Focus()
	inputs[0].CharLimit = 64
	inputs[0].Width = 40

	// Secret Key (auto-generate)
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Secret Key (auto-generated)"
	inputs[1].CharLimit = 128
	inputs[1].Width = 60
	inputs[1].SetValue(generateSecretKey())

	// Role
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Role (admin/user/userplus)"
	inputs[2].CharLimit = 20
	inputs[2].Width = 30
	inputs[2].SetValue("user")

	m.userFormInputs = inputs
	m.focusIndex = 0
	m.userFormInputs = inputs
	m.focusIndex = 0
}

// initUpdateUserForm initializes the user update form with existing data
func (m *Model) initUpdateUserForm(user models.User) {
	inputs := make([]textinput.Model, 3)

	// Access Key (read-only)
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Access Key (username)"
	inputs[0].SetValue(user.Access)
	inputs[0].Blur() // Cannot edit access key
	inputs[0].CharLimit = 64
	inputs[0].Width = 40

	// Secret Key
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Secret Key"
	inputs[1].SetValue(user.Secret)
	inputs[1].Focus()
	inputs[1].CharLimit = 128
	inputs[1].Width = 60

	// Role
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Role (admin/user/userplus)"
	inputs[2].SetValue(user.Role)
	inputs[2].CharLimit = 20
	inputs[2].Width = 30

	m.userFormInputs = inputs
	m.focusIndex = 1 // Start focus on Secret Key
}

// initBucketForm initializes the bucket creation form
func (m *Model) initBucketForm() {
	inputs := make([]textinput.Model, 3)

	// Bucket Name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Bucket Name"
	inputs[0].Focus()
	inputs[0].CharLimit = 63
	inputs[0].Width = 40

	// Quota
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Quota (e.g., 2T, 500G, 100M)"
	inputs[1].CharLimit = 20
	inputs[1].Width = 30
	inputs[1].SetValue("1T")

	// Owner
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Owner Access Key"
	inputs[2].CharLimit = 64
	inputs[2].Width = 40

	m.bucketFormInputs = inputs
	m.focusIndex = 0
}

// initChangeOwnerForm initializes the change bucket owner form
func (m *Model) initChangeOwnerForm() {
	inputs := make([]textinput.Model, 2)

	// Bucket Name
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Bucket Name"
	inputs[0].Focus()
	inputs[0].CharLimit = 63
	inputs[0].Width = 40

	// New Owner
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "New Owner Access Key"
	inputs[1].CharLimit = 64
	inputs[1].Width = 40

	m.changeOwnerFormInputs = inputs
	m.focusIndex = 0
}

// initProvisionForm initializes the combined provision form
func (m *Model) initProvisionForm() {
	inputs := make([]textinput.Model, 9)

	// Access Key
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Access Key (username)"
	inputs[0].Focus()
	inputs[0].CharLimit = 64
	inputs[0].Width = 40

	// Secret Key
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Secret Key (auto-generated if empty)"
	inputs[1].CharLimit = 128
	inputs[1].Width = 60
	inputs[1].SetValue(generateSecretKey())

	// Role
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Role (admin/user/userplus)"
	inputs[2].CharLimit = 20
	inputs[2].Width = 30
	inputs[2].SetValue("user")

	// User ID
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "User ID (default 0)"
	inputs[3].CharLimit = 10
	inputs[3].Width = 15
	inputs[3].SetValue("0")

	// Group ID
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "Group ID (default 0)"
	inputs[4].CharLimit = 10
	inputs[4].Width = 15
	inputs[4].SetValue("0")

	// Project ID
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Project ID (optional)"
	inputs[5].CharLimit = 10
	inputs[5].Width = 15
	inputs[5].SetValue("0")

	// Bucket Name
	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Bucket Name"
	inputs[6].CharLimit = 63
	inputs[6].Width = 40

	// Quota
	inputs[7] = textinput.New()
	inputs[7].Placeholder = "Quota (e.g., 2T, 500G)"
	inputs[7].CharLimit = 20
	inputs[7].Width = 30
	inputs[7].SetValue("1T")

	// Owner
	inputs[8] = textinput.New()
	inputs[8].Placeholder = "Owner Access Key (default = access)"
	inputs[8].CharLimit = 64
	inputs[8].Width = 40

	m.provisionFormInputs = inputs
	m.focusIndex = 0
	m.provisionScroll = 0
}

// renderCreateUserForm renders the user creation/update form
func (m Model) renderCreateUserForm() string {
	var s strings.Builder

	// Title
	if m.currentView == UpdateUserView {
		s.WriteString(titleStyle.Render("Update User") + "\n\n")
	} else {
		s.WriteString(titleStyle.Render("Create New User") + "\n\n")
	}

	// Form fields
	labels := []string{"Access Key:", "Secret Key:", "Role:"}

	for i, input := range m.userFormInputs {
		label := inputLabelStyle.Render(labels[i])
		s.WriteString(label + "\n")

		if i == m.focusIndex {
			s.WriteString(focusedInputStyle.Render(input.View()) + "\n\n")
		} else {
			s.WriteString(inputStyle.Render(input.View()) + "\n\n")
		}
	}

	// Buttons
	createBtn := "[ Create User ]"
	if m.currentView == UpdateUserView {
		createBtn = "[ Update User ]"
	}
	cancelBtn := "[ Cancel ]"

	if m.focusIndex == len(m.userFormInputs) {
		s.WriteString(focusedButtonStyle.Render(createBtn) + "  ")
		s.WriteString(buttonStyle.Render(cancelBtn) + "\n")
	} else if m.focusIndex == len(m.userFormInputs)+1 {
		s.WriteString(buttonStyle.Render(createBtn) + "  ")
		s.WriteString(focusedButtonStyle.Render(cancelBtn) + "\n")
	} else {
		s.WriteString(buttonStyle.Render(createBtn) + "  ")
		s.WriteString(buttonStyle.Render(cancelBtn) + "\n")
	}

	// Help text
	help := helpStyle.Render("tab: Next field • shift+tab: Previous • enter: Submit/Select • esc: Cancel")
	s.WriteString("\n" + help)

	// Error/Success messages
	if m.errorMessage != "" {
		s.WriteString("\n" + errorStyle.Render("Error: "+m.errorMessage))
	}
	if m.successMessage != "" {
		s.WriteString("\n" + successStyle.Render(m.successMessage))
	}

	return s.String()
}

// renderProvisionForm renders the combined provision form
func (m Model) renderProvisionForm() string {
	lines := m.provisionLines()

	visible := m.viewportHeight()
	if visible < 5 {
		visible = 5
	}

	scroll := m.provisionScroll
	maxOffset := len(lines) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}
	if scroll > maxOffset {
		scroll = maxOffset
	}
	if scroll < 0 {
		scroll = 0
	}

	start := scroll
	end := scroll + visible
	if end > len(lines) {
		end = len(lines)
	}

	var s strings.Builder
	for i := start; i < end; i++ {
		s.WriteString(lines[i])
		if i != end-1 {
			s.WriteString("\n")
		}
	}

	return s.String()
}

// renderCreateBucketForm renders the bucket creation form
func (m Model) renderCreateBucketForm() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Create New Bucket") + "\n\n")

	// Form fields
	labels := []string{"Bucket Name:", "Quota:", "Owner:"}

	for i, input := range m.bucketFormInputs {
		label := inputLabelStyle.Render(labels[i])
		s.WriteString(label + "\n")

		if i == m.focusIndex {
			s.WriteString(focusedInputStyle.Render(input.View()) + "\n\n")
		} else {
			s.WriteString(inputStyle.Render(input.View()) + "\n\n")
		}
	}

	// Buttons
	createBtn := "[ Create Bucket ]"
	cancelBtn := "[ Cancel ]"

	if m.focusIndex == len(m.bucketFormInputs) {
		s.WriteString(focusedButtonStyle.Render(createBtn) + "  ")
		s.WriteString(buttonStyle.Render(cancelBtn) + "\n")
	} else if m.focusIndex == len(m.bucketFormInputs)+1 {
		s.WriteString(buttonStyle.Render(createBtn) + "  ")
		s.WriteString(focusedButtonStyle.Render(cancelBtn) + "\n")
	} else {
		s.WriteString(buttonStyle.Render(createBtn) + "  ")
		s.WriteString(buttonStyle.Render(cancelBtn) + "\n")
	}

	// Help text
	help := helpStyle.Render("tab: Next field • shift+tab: Previous • enter: Submit/Select • esc: Cancel")
	s.WriteString("\n" + help)

	// Info
	info := dimStyle.Render("Note: Bucket will be created with ZFS and mountpoint will be /tank/s3/buckets/<name>")
	s.WriteString("\n" + info)

	// Error/Success messages
	if m.errorMessage != "" {
		s.WriteString("\n" + errorStyle.Render("Error: "+m.errorMessage))
	}
	if m.successMessage != "" {
		s.WriteString("\n" + successStyle.Render(m.successMessage))
	}

	return s.String()
}

// renderChangeOwnerForm renders the change bucket owner form
func (m Model) renderChangeOwnerForm() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Change Bucket Owner") + "\n\n")

	labels := []string{"Bucket Name:", "New Owner Access Key:"}

	for i, input := range m.changeOwnerFormInputs {
		label := inputLabelStyle.Render(labels[i])
		s.WriteString(label + "\n")

		if i == m.focusIndex {
			s.WriteString(focusedInputStyle.Render(input.View()) + "\n\n")
		} else {
			s.WriteString(inputStyle.Render(input.View()) + "\n\n")
		}
	}

	updateBtn := "[ Change Owner ]"
	cancelBtn := "[ Cancel ]"

	if m.focusIndex == len(m.changeOwnerFormInputs) {
		s.WriteString(focusedButtonStyle.Render(updateBtn) + "  ")
		s.WriteString(buttonStyle.Render(cancelBtn) + "\n")
	} else if m.focusIndex == len(m.changeOwnerFormInputs)+1 {
		s.WriteString(buttonStyle.Render(updateBtn) + "  ")
		s.WriteString(focusedButtonStyle.Render(cancelBtn) + "\n")
	} else {
		s.WriteString(buttonStyle.Render(updateBtn) + "  ")
		s.WriteString(buttonStyle.Render(cancelBtn) + "\n")
	}

	help := helpStyle.Render("tab: Next field • shift+tab: Previous • enter: Submit/Select • esc: Cancel")
	s.WriteString("\n" + help)

	if m.errorMessage != "" {
		s.WriteString("\n" + errorStyle.Render("Error: "+m.errorMessage))
	}
	if m.successMessage != "" {
		s.WriteString("\n" + successStyle.Render(m.successMessage))
	}

	return s.String()
}

// handleCreateBucket handles bucket creation
func (m Model) handleCreateBucket() (tea.Model, tea.Cmd) {
	req := models.BucketCreateRequest{
		Name:  m.bucketFormInputs[0].Value(),
		Quota: m.bucketFormInputs[1].Value(),
		Owner: m.bucketFormInputs[2].Value(),
	}

	// Validate
	if req.Name == "" {
		m.errorMessage = "Bucket name is required"
		return m, nil
	}
	if req.Quota == "" {
		m.errorMessage = "Quota is required"
		return m, nil
	}

	// Create bucket
	err := m.bucketService.CreateBucket(req)
	if err != nil {
		m.errorMessage = fmt.Sprintf("Failed to create bucket: %v", err)
		return m, nil
	}

	// If owner is specified, change bucket owner
	if req.Owner != "" {
		err = m.versitygwService.ChangeBucketOwner(req.Name, req.Owner)
		if err != nil {
			m.errorMessage = fmt.Sprintf("Bucket created but failed to set owner: %v", err)
			return m, nil
		}
	}

	m.successMessage = fmt.Sprintf("Bucket '%s' created successfully!", req.Name)
	m.currentView = m.returnView
	m.cursor = 0

	return m, nil
}

// handleChangeOwner handles bucket owner change
func (m Model) handleChangeOwner() (tea.Model, tea.Cmd) {
	bucket := m.changeOwnerFormInputs[0].Value()
	owner := m.changeOwnerFormInputs[1].Value()

	if bucket == "" {
		m.errorMessage = "Bucket name is required"
		return m, nil
	}
	if owner == "" {
		m.errorMessage = "New owner is required"
		return m, nil
	}

	if err := m.versitygwService.ChangeBucketOwner(bucket, owner); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to change owner: %v", err)
		return m, nil
	}

	// Optimistically update the local bucket list to reflect the change immediately
	for i := range m.buckets {
		if m.buckets[i].Name == bucket {
			m.buckets[i].Owner = owner
			break
		}
	}

	m.successMessage = fmt.Sprintf("Owner for '%s' changed to '%s'", bucket, owner)
	m.currentView = m.returnView
	m.cursor = 0

	return m, nil
}

// handleProvision handles creating user + bucket + owner in one go
func (m Model) handleProvision() (tea.Model, tea.Cmd) {
	access := strings.TrimSpace(m.provisionFormInputs[0].Value())
	secret := strings.TrimSpace(m.provisionFormInputs[1].Value())
	role := strings.TrimSpace(m.provisionFormInputs[2].Value())
	uidStr := strings.TrimSpace(m.provisionFormInputs[3].Value())
	gidStr := strings.TrimSpace(m.provisionFormInputs[4].Value())
	pidStr := strings.TrimSpace(m.provisionFormInputs[5].Value())
	bucket := strings.TrimSpace(m.provisionFormInputs[6].Value())
	quota := strings.TrimSpace(m.provisionFormInputs[7].Value())
	owner := strings.TrimSpace(m.provisionFormInputs[8].Value())

	// Defaults and validation
	if access == "" {
		m.errorMessage = "Access key is required"
		return m, nil
	}
	if role != "admin" && role != "user" && role != "userplus" {
		m.errorMessage = "Role must be admin, user, or userplus"
		return m, nil
	}
	if bucket == "" {
		m.errorMessage = "Bucket name is required"
		return m, nil
	}
	if quota == "" {
		m.errorMessage = "Quota is required"
		return m, nil
	}
	if owner == "" {
		owner = access
	}
	if secret == "" {
		secret = generateSecretKey()
	}

	parseInt := func(val string) (int, error) {
		if val == "" {
			return 0, nil
		}
		return strconv.Atoi(val)
	}

	uid, err := parseInt(uidStr)
	if err != nil {
		m.errorMessage = "User ID must be a number"
		return m, nil
	}
	gid, err := parseInt(gidStr)
	if err != nil {
		m.errorMessage = "Group ID must be a number"
		return m, nil
	}
	pid, err := parseInt(pidStr)
	if err != nil {
		m.errorMessage = "Project ID must be a number"
		return m, nil
	}

	userReq := models.UserCreateRequest{
		Access:    access,
		Secret:    secret,
		Role:      role,
		UserID:    uid,
		GroupID:   gid,
		ProjectID: pid,
	}

	if err := m.versitygwService.CreateUser(userReq); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to create user: %v", err)
		return m, nil
	}

	bucketReq := models.BucketCreateRequest{
		Name:  bucket,
		Quota: quota,
		Owner: owner,
	}

	if err := m.bucketService.CreateBucket(bucketReq); err != nil {
		m.errorMessage = fmt.Sprintf("Failed to create bucket: %v", err)
		return m, nil
	}

	if err := m.versitygwService.ChangeBucketOwner(bucket, owner); err != nil {
		m.errorMessage = fmt.Sprintf("Bucket created but failed to set owner: %v", err)
		return m, nil
	}

	m.successMessage = fmt.Sprintf("Provisioned user '%s' and bucket '%s' (owner '%s')", access, bucket, owner)
	m.currentView = m.returnView
	m.cursor = 0

	return m, nil
}

func (m Model) provisionLines() []string {
	var lines []string

	add := func(text string) {
		parts := strings.Split(text, "\n")
		lines = append(lines, parts...)
	}

	add(titleStyle.Render("Provision: User + Bucket"))
	add("")

	labels := []string{
		"Access Key:",
		"Secret Key:",
		"Role:",
		"User ID:",
		"Group ID:",
		"Project ID:",
		"Bucket Name:",
		"Quota:",
		"Owner:",
	}

	for i, input := range m.provisionFormInputs {
		label := inputLabelStyle.Render(labels[i])

		var renderedInput string
		if i == m.focusIndex {
			renderedInput = focusedInputStyle.Render(input.View())
		} else {
			renderedInput = inputStyle.Render(input.View())
		}

		line := lipgloss.JoinHorizontal(lipgloss.Top, label, renderedInput)
		add(line)
	}

	createBtn := "[ Provision ]"
	cancelBtn := "[ Cancel ]"

	if m.focusIndex == len(m.provisionFormInputs) {
		add(focusedButtonStyle.Render(createBtn) + "  " + buttonStyle.Render(cancelBtn))
	} else if m.focusIndex == len(m.provisionFormInputs)+1 {
		add(buttonStyle.Render(createBtn) + "  " + focusedButtonStyle.Render(cancelBtn))
	} else {
		add(buttonStyle.Render(createBtn) + "  " + buttonStyle.Render(cancelBtn))
	}

	add(helpStyle.Render("tab: Next • shift+tab: Prev • enter: Submit/Select • esc: Cancel • PgUp/PgDn scroll"))
	add(dimStyle.Render("Creates user, then bucket, then sets bucket owner."))

	if m.errorMessage != "" {
		add(errorStyle.Render("Error: " + m.errorMessage))
	}
	if m.successMessage != "" {
		add(successStyle.Render(m.successMessage))
	}

	return lines
}

func (m Model) provisionTargetLine() int {
	fields := len(m.provisionFormInputs)
	if m.focusIndex < fields {
		return 2 + m.focusIndex
	}
	if m.focusIndex == fields || m.focusIndex == fields+1 {
		return 2 + fields
	}
	return 0
}

func (m Model) ensureProvisionVisible(lines []string) Model {
	visible := m.viewportHeight()
	if visible < 5 {
		visible = 5
	}

	maxOffset := len(lines) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}

	target := m.provisionTargetLine()
	scroll := m.provisionScroll

	if target < scroll {
		scroll = target
	} else if target >= scroll+visible {
		scroll = target - visible + 1
	}

	if scroll < 0 {
		scroll = 0
	}
	if scroll > maxOffset {
		scroll = maxOffset
	}

	m.provisionScroll = scroll
	return m
}

func (m Model) clampProvisionScroll(lines []string) Model {
	visible := m.viewportHeight()
	if visible < 5 {
		visible = 5
	}
	maxOffset := len(lines) - visible
	if maxOffset < 0 {
		maxOffset = 0
	}

	if m.provisionScroll < 0 {
		m.provisionScroll = 0
	}
	if m.provisionScroll > maxOffset {
		m.provisionScroll = maxOffset
	}
	return m
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// generateSecretKey generates a random secret key
func generateSecretKey() string {
	b := make([]byte, 48)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// initMakePublicForm initializes the make public form
func (m *Model) initMakePublicForm() {
	m.bucketFormInputs = make([]textinput.Model, 2)

	// Reusing styles from other forms (assuming they are package level or need to be accessed via correct names)
	// Actually styles like focusedInputStyle are in ui/styles.go but not exported or are private to views.
	// Oh wait, styles.go defines package level variables.
	// Checking previous form code:
	// t.Cursor.Style = cursorStyle  <-- this seems used in other forms.
	// t.PromptStyle = focusedStyle <-- this seems WRONG based on renderCreateUserForm using `focusedInputStyle.Render(input.View())`
	// Actually, textinput.Model has its own style fields?
	// Let's look at initUserForm:
	/*
		inputs[2].Placeholder = "Role (admin/user/userplus)"
		inputs[2].CharLimit = 20
		inputs[2].Width = 30
		inputs[2].SetValue("user")
	*/
	// It doesn't set PromptStyle usually.
	// Let's remove the style setting on init, relying on Render loop to style it.

	var t textinput.Model
	// Bucket Name
	t = textinput.New()
	t.CharLimit = 64
	t.Placeholder = "Bucket Name"
	t.Focus()
	m.bucketFormInputs[0] = t

	// Owner
	t = textinput.New()
	t.CharLimit = 64
	t.Placeholder = "Owner Access Key"
	m.bucketFormInputs[1] = t

	m.focusIndex = 0
}

// renderMakePublicForm renders the make public form
func (m Model) renderMakePublicForm() string {
	var s strings.Builder

	s.WriteString(titleStyle.Render("Make Bucket Public") + "\n\n")

	// Inputs
	labels := []string{"Bucket Name:", "Owner Access Key:"}

	for i, input := range m.bucketFormInputs {
		label := inputLabelStyle.Render(labels[i])
		s.WriteString(label + "\n")

		if i == m.focusIndex {
			s.WriteString(focusedInputStyle.Render(input.View()) + "\n\n")
		} else {
			s.WriteString(inputStyle.Render(input.View()) + "\n\n")
		}
	}

	// Buttons
	createBtn := "[ Make Public ]"
	cancelBtn := "[ Cancel ]"

	if m.focusIndex == len(m.bucketFormInputs) {
		s.WriteString(focusedButtonStyle.Render(createBtn) + "  ")
		s.WriteString(buttonStyle.Render(cancelBtn) + "\n")
	} else if m.focusIndex == len(m.bucketFormInputs)+1 {
		s.WriteString(buttonStyle.Render(createBtn) + "  ")
		s.WriteString(focusedButtonStyle.Render(cancelBtn) + "\n")
	} else {
		s.WriteString(buttonStyle.Render(createBtn) + "  ")
		s.WriteString(buttonStyle.Render(cancelBtn) + "\n")
	}

	help := helpStyle.Render("tab: Next field • enter: Submit/Select • esc: Cancel")
	s.WriteString("\n" + help)

	if m.errorMessage != "" {
		s.WriteString("\n" + errorStyle.Render("Error: "+m.errorMessage))
	} else if m.successMessage != "" {
		s.WriteString("\n" + successStyle.Render(m.successMessage))
	}

	return s.String()
}

// handleMakePublic handles making a bucket public
func (m Model) handleMakePublic() (tea.Model, tea.Cmd) {
	bucketName := m.bucketFormInputs[0].Value()
	owner := m.bucketFormInputs[1].Value()

	// Validate
	if bucketName == "" {
		m.errorMessage = "Bucket name is required"
		return m, nil
	}
	if owner == "" {
		m.errorMessage = "Owner access key is required"
		return m, nil
	}

	policy := services.GeneratePublicPolicy(bucketName, owner)
	err := m.versitygwService.SetBucketPolicy(bucketName, policy)
	if err != nil {
		m.errorMessage = fmt.Sprintf("Failed to make public: %v", err)
		return m, nil
	}

	m.successMessage = fmt.Sprintf("Bucket '%s' is now PUBLIC!", bucketName)
	m.currentView = m.returnView
	m.cursor = 0

	return m, nil
}
