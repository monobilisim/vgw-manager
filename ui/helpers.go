package ui

import (
	"fmt"

	"github.com/atotto/clipboard"
)

// getMaxCursorForView returns the maximum cursor index for the current view
func (m Model) getMaxCursorForView() int {
	switch m.currentView {
	case MainMenuView:
		return 3 // 4 menu items (0-3)
	case OperationsView:
		return 4 // 5 operations (0-4)
	case UsersListView:
		return len(m.users) - 1
	case BucketsListView:
		return len(m.buckets) - 1
	default:
		return 0
	}
}

// copyUserCredentials copies the selected user's credentials to clipboard
func (m *Model) copyUserCredentials() error {
	if m.cursor < 0 || m.cursor >= len(m.users) {
		return fmt.Errorf("invalid selection")
	}

	user := m.users[m.cursor]
	credentials := fmt.Sprintf("Access Key: %s\nSecret Key: %s", user.Access, user.Secret)

	err := clipboard.WriteAll(credentials)
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

// copyUserCredentialsFromDetail copies user credentials from detail view
func (m *Model) copyUserCredentialsFromDetail() error {
	if m.selectedUserIndex < 0 || m.selectedUserIndex >= len(m.users) {
		return fmt.Errorf("invalid selection")
	}

	user := m.users[m.selectedUserIndex]
	credentials := fmt.Sprintf("Access Key: %s\nSecret Key: %s", user.Access, user.Secret)

	err := clipboard.WriteAll(credentials)
	if err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}

func (m Model) viewportHeight() int {
	if m.height > 0 {
		return m.height
	}
	return 24
}
