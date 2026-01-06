package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/omertahaoztop/vgw-manager/models"
)

// updateUserForm handles key events for the user creation form
func (m Model) updateUserForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "q":
		m.currentView = m.returnView
		m.cursor = 0
		return m, nil

	case "tab":
		m.focusIndex++
		if m.focusIndex > len(m.userFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateUserFormFocus()
		return m, nil

	case "shift+tab", "up":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.userFormInputs) + 1
		}
		m.updateUserFormFocus()
		return m, nil

	case "down":
		m.focusIndex++
		if m.focusIndex > len(m.userFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateUserFormFocus()
		return m, nil

	case "enter":
		if m.focusIndex == len(m.userFormInputs) {
			// Create button pressed
			return m.handleCreateUser()
		} else if m.focusIndex == len(m.userFormInputs)+1 {
			// Cancel button pressed
			m.currentView = m.returnView
			m.cursor = 0
			return m, nil
		}
	}

	// Update the focused input
	if m.focusIndex < len(m.userFormInputs) {
		m.userFormInputs[m.focusIndex], cmd = m.userFormInputs[m.focusIndex].Update(msg)
	}

	return m, cmd
}

// updateBucketForm handles key events for the bucket creation form
func (m Model) updateBucketForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "q":
		m.currentView = m.returnView
		m.cursor = 0
		return m, nil

	case "tab":
		m.focusIndex++
		if m.focusIndex > len(m.bucketFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateBucketFormFocus()
		return m, nil

	case "shift+tab", "up":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.bucketFormInputs) + 1
		}
		m.updateBucketFormFocus()
		return m, nil

	case "down":
		m.focusIndex++
		if m.focusIndex > len(m.bucketFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateBucketFormFocus()
		return m, nil

	case "enter":
		if m.focusIndex == len(m.bucketFormInputs) {
			// Create button pressed
			return m.handleCreateBucket()
		} else if m.focusIndex == len(m.bucketFormInputs)+1 {
			// Cancel button pressed
			m.currentView = m.returnView
			m.cursor = 0
			return m, nil
		}
	}

	// Update the focused input
	if m.focusIndex < len(m.bucketFormInputs) {
		m.bucketFormInputs[m.focusIndex], cmd = m.bucketFormInputs[m.focusIndex].Update(msg)
	}

	return m, cmd
}

// updateUserFormFocus updates focus state for user form inputs
func (m *Model) updateUserFormFocus() {
	for i := range m.userFormInputs {
		if i == m.focusIndex {
			m.userFormInputs[i].Focus()
		} else {
			m.userFormInputs[i].Blur()
		}
	}
}

// handleCreateUser handles user creation or update
func (m Model) handleCreateUser() (tea.Model, tea.Cmd) {
	req := models.UserCreateRequest{
		Access:    m.userFormInputs[0].Value(),
		Secret:    m.userFormInputs[1].Value(),
		Role:      m.userFormInputs[2].Value(),
		UserID:    0,
		GroupID:   0,
		ProjectID: 0,
	}

	// Validate
	if req.Access == "" {
		m.errorMessage = "Access key is required"
		return m, nil
	}
	if req.Secret == "" {
		m.errorMessage = "Secret key is required"
		return m, nil
	}
	if req.Role != "admin" && req.Role != "user" && req.Role != "userplus" {
		m.errorMessage = "Role must be admin, user, or userplus"
		return m, nil
	}

	if m.currentView == UpdateUserView {
		updateReq := models.UserUpdateRequest{
			Access:    req.Access,
			Secret:    req.Secret,
			Role:      req.Role,
			UserID:    req.UserID,
			GroupID:   req.GroupID,
			ProjectID: req.ProjectID,
		}
		if err := m.versitygwService.UpdateUser(updateReq); err != nil {
			m.errorMessage = fmt.Sprintf("Failed to update user: %v", err)
			return m, nil
		}
		m.successMessage = fmt.Sprintf("User '%s' updated successfully!", req.Access)
	} else {
		// Create user
		if err := m.versitygwService.CreateUser(req); err != nil {
			m.errorMessage = fmt.Sprintf("Failed to create user: %v", err)
			return m, nil
		}
		m.successMessage = fmt.Sprintf("User '%s' created successfully!", req.Access)
	}

	m.currentView = m.returnView
	m.cursor = 0

	// Reload users
	users, err := m.userService.ListUsers()
	if err == nil {
		m.users = users
	}

	return m, nil
}

// updateBucketFormFocus updates focus state for bucket form inputs
func (m *Model) updateBucketFormFocus() {
	for i := range m.bucketFormInputs {
		if i == m.focusIndex {
			m.bucketFormInputs[i].Focus()
		} else {
			m.bucketFormInputs[i].Blur()
		}
	}
}

// updateChangeOwnerForm handles key events for the change owner form
func (m Model) updateChangeOwnerForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "q":
		m.currentView = m.returnView
		m.cursor = 0
		return m, nil

	case "tab":
		m.focusIndex++
		if m.focusIndex > len(m.changeOwnerFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateChangeOwnerFormFocus()
		return m, nil

	case "shift+tab", "up":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.changeOwnerFormInputs) + 1
		}
		m.updateChangeOwnerFormFocus()
		return m, nil

	case "down":
		m.focusIndex++
		if m.focusIndex > len(m.changeOwnerFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateChangeOwnerFormFocus()
		return m, nil

	case "enter":
		if m.focusIndex == len(m.changeOwnerFormInputs) {
			return m.handleChangeOwner()
		} else if m.focusIndex == len(m.changeOwnerFormInputs)+1 {
			m.currentView = m.returnView
			// m.cursor = 0 // Keep cursor position
			return m, nil
		}
	}

	if m.focusIndex < len(m.changeOwnerFormInputs) {
		m.changeOwnerFormInputs[m.focusIndex], cmd = m.changeOwnerFormInputs[m.focusIndex].Update(msg)
	}

	return m, cmd
}

// updateChangeOwnerFormFocus updates focus state for change owner form inputs
func (m *Model) updateChangeOwnerFormFocus() {
	for i := range m.changeOwnerFormInputs {
		if i == m.focusIndex {
			m.changeOwnerFormInputs[i].Focus()
		} else {
			m.changeOwnerFormInputs[i].Blur()
		}
	}
}

// updateProvisionForm handles key events for the provision form
func (m Model) updateProvisionForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "q":
		m.currentView = m.returnView
		m.cursor = 0
		return m, nil

	case "tab":
		m.focusIndex++
		if m.focusIndex > len(m.provisionFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateProvisionFormFocus()
		m = m.ensureProvisionVisible(m.provisionLines())
		return m, nil

	case "shift+tab", "up":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.provisionFormInputs) + 1
		}
		m.updateProvisionFormFocus()
		m = m.ensureProvisionVisible(m.provisionLines())
		return m, nil

	case "down":
		m.focusIndex++
		if m.focusIndex > len(m.provisionFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateProvisionFormFocus()
		m = m.ensureProvisionVisible(m.provisionLines())
		return m, nil

	case "pgup":
		m.provisionScroll -= max(1, m.viewportHeight()/2)
		m = m.clampProvisionScroll(m.provisionLines())
		return m, nil
	case "pgdown":
		m.provisionScroll += max(1, m.viewportHeight()/2)
		m = m.clampProvisionScroll(m.provisionLines())
		return m, nil
	case "home":
		m.provisionScroll = 0
		return m, nil
	case "end":
		m.provisionScroll = len(m.provisionLines())
		m = m.clampProvisionScroll(m.provisionLines())
		return m, nil

	case "enter":
		if m.focusIndex == len(m.provisionFormInputs) {
			return m.handleProvision()
		} else if m.focusIndex == len(m.provisionFormInputs)+1 {
			m.currentView = m.returnView
			m.cursor = 0
			return m, nil
		}
	}

	if m.focusIndex < len(m.provisionFormInputs) {
		m.provisionFormInputs[m.focusIndex], cmd = m.provisionFormInputs[m.focusIndex].Update(msg)
	}

	m = m.ensureProvisionVisible(m.provisionLines())

	return m, cmd
}

// updateProvisionFormFocus updates focus state for provision form inputs
func (m *Model) updateProvisionFormFocus() {
	for i := range m.provisionFormInputs {
		if i == m.focusIndex {
			m.provisionFormInputs[i].Focus()
		} else {
			m.provisionFormInputs[i].Blur()
		}
	}
}

// updateMakePublicForm handles key events for the make public form
func (m Model) updateMakePublicForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc", "q":
		m.currentView = m.returnView
		m.cursor = 0
		return m, nil

	case "tab", "down":
		m.focusIndex++
		if m.focusIndex > len(m.bucketFormInputs)+1 {
			m.focusIndex = 0
		}
		m.updateBucketFormFocus() // Re-use bucket form focus logic since we used bucketFormInputs
		return m, nil

	case "shift+tab", "up":
		m.focusIndex--
		if m.focusIndex < 0 {
			m.focusIndex = len(m.bucketFormInputs) + 1
		}
		m.updateBucketFormFocus()
		return m, nil

	case "enter":
		if m.focusIndex == len(m.bucketFormInputs) {
			return m.handleMakePublic()
		} else if m.focusIndex == len(m.bucketFormInputs)+1 {
			m.currentView = m.returnView
			return m, nil
		}
	}

	if m.focusIndex < len(m.bucketFormInputs) {
		m.bucketFormInputs[m.focusIndex], cmd = m.bucketFormInputs[m.focusIndex].Update(msg)
	}

	return m, cmd
}
