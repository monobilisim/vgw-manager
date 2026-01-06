package ui

import (
	"fmt"
	"strings"
)

// renderMainMenu renders the main menu
func (m Model) renderMainMenu() string {
	var s strings.Builder

	// Title
	title := titleStyle.Render("VGW Manager")
	subtitle := subtitleStyle.Render("VersityGW Management Tool")

	s.WriteString(title + "\n")
	s.WriteString(subtitle + "\n\n")

	// Menu items
	menuItems := []string{
		"List Users",
		"List Buckets",
		"Operations",
		"Quit",
	}

	for i, item := range menuItems {
		cursor := " "
		if m.cursor == i {
			cursor = "▶"
			s.WriteString(cursor + " " + selectedMenuItemStyle.Render(item) + "\n")
		} else {
			s.WriteString(cursor + " " + menuItemStyle.Render(item) + "\n")
		}
	}

	// Help text
	help := helpStyle.Render("↑/↓: Navigate • Enter: Select • q: Quit")
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

// renderOperationsMenu renders the combined operations menu
func (m Model) renderOperationsMenu() string {
	var s strings.Builder

	// Title
	title := titleStyle.Render("Operations")
	subtitle := subtitleStyle.Render("Create User • Create Bucket • Change Owner • Provision")

	s.WriteString(title + "\n")
	s.WriteString(subtitle + "\n\n")

	menuItems := []string{
		"Create User",
		"Create Bucket",
		"Change Bucket Owner",
		"Make Bucket Public",
		"Provision (User + Bucket)",
	}

	for i, item := range menuItems {
		cursor := " "
		if m.cursor == i {
			cursor = "▶"
			s.WriteString(cursor + " " + selectedMenuItemStyle.Render(item) + "\n")
		} else {
			s.WriteString(cursor + " " + menuItemStyle.Render(item) + "\n")
		}
	}

	help := helpStyle.Render("↑/↓: Navigate • Enter: Select • esc/q: Back to main menu")
	s.WriteString("\n" + help)

	if m.errorMessage != "" {
		s.WriteString("\n" + errorStyle.Render("Error: "+m.errorMessage))
	}
	if m.successMessage != "" {
		s.WriteString("\n" + successStyle.Render(m.successMessage))
	}

	return s.String()
}

// renderUsersList renders the users list view
func (m Model) renderUsersList() string {
	s := strings.Builder{}
	s.WriteString(titleStyle.Render("Users List") + "\n\n")

	// Pagination logic
	start := m.page * m.pageSize
	end := start + m.pageSize
	if end > len(m.users) {
		end = len(m.users)
	}

	// Header
	header := fmt.Sprintf("  %-20s %-20s %-10s", "ACCESS KEY", "SECRET KEY", "ROLE")
	s.WriteString(dimStyle.Render(header) + "\n")
	s.WriteString(dimStyle.Render(strings.Repeat("-", 60)) + "\n")

	for i := start; i < end; i++ {
		user := m.users[i]
		cursor := " "
		// absolute index i, cursor is relative
		if m.cursor == (i - start) {
			cursor = ">"
		}

		// Helper to safely truncate strings
		truncate := func(s string, max int) string {
			if len(s) > max {
				return s[:max-3] + "..."
			}
			return s
		}

		// Row content
		line := fmt.Sprintf("%s %-20s %-20s %-10s", cursor, truncate(user.Access, 20), truncate(user.Secret, 20), user.Role)

		if m.cursor == (i - start) {
			s.WriteString(selectedTableRowStyle.Render(line) + "\n")
		} else {
			s.WriteString(line + "\n")
		}
	}

	// Pagination Footer
	totalPages := (len(m.users) + m.pageSize - 1) / m.pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	pageInfo := fmt.Sprintf("Page %d of %d (%d items)", m.page+1, totalPages, len(m.users))
	s.WriteString("\n" + helpStyle.Render(pageInfo))

	// Help text
	help := helpStyle.Render("↑/↓: Navigate • ←/→: Page • c: Copy • e: Edit • d: Delete • Enter: View • q: Back")
	s.WriteString("\n" + help)

	// Error/Success messages
	if m.errorMessage != "" {
		s.WriteString("\n" + errorStyle.Render(m.errorMessage))
	} else if m.successMessage != "" {
		s.WriteString("\n" + successStyle.Render(m.successMessage))
	}

	return s.String()
}

// renderBucketsList renders the buckets list view
func (m Model) renderBucketsList() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Buckets List") + "\n\n")

	// Pagination logic
	start := m.page * m.pageSize
	end := start + m.pageSize
	if end > len(m.buckets) {
		end = len(m.buckets)
	}

	// Header (Matched to previous preferred layout)
	// Name (30) | Mountpoint (40) | Quota (10) | Used (10) | Avail (10) | Owner (15)
	header := fmt.Sprintf("  %-30s %-40s %-10s %-10s %-10s %-15s", "Name", "Mountpoint", "Quota", "Used", "Available", "Owner")
	s.WriteString(dimStyle.Render(header) + "\n")
	s.WriteString(dimStyle.Render(strings.Repeat("-", 125)) + "\n")

	for i := start; i < end; i++ {
		bucket := m.buckets[i]
		cursor := " "
		if m.cursor == (i - start) {
			cursor = ">"
		}

		// Owner display
		owner := bucket.Owner
		if owner == "" {
			owner = "unknown"
		}

		// Truncate helper
		trunc := func(str string, w int) string {
			if len(str) > w {
				return str[:w-3] + "..."
			}
			return str
		}

		// Row content
		line := fmt.Sprintf("%s %-30s %-40s %-10s %-10s %-10s %-15s",
			cursor,
			trunc(bucket.Name, 30),
			trunc(bucket.Mountpoint, 40),
			bucket.Quota,
			bucket.Used,      // Assuming bucket.Used is already formatted or a string
			bucket.Available, // Available is already formatted string from service? Let's check service. Assuming yes or string.
			owner,
		)

		if m.cursor == (i - start) {
			s.WriteString(selectedTableRowStyle.Render(line) + "\n")
		} else {
			s.WriteString(line + "\n")
		}
	}
	// Pagination Footer
	totalPages := (len(m.buckets) + m.pageSize - 1) / m.pageSize
	if totalPages == 0 {
		totalPages = 1
	}
	pageInfo := fmt.Sprintf("Page %d of %d (%d items)", m.page+1, totalPages, len(m.buckets))
	s.WriteString("\n" + helpStyle.Render(pageInfo))

	// Help text
	help := helpStyle.Render("↑/k: Up • ↓/j: Down • ←/h: Prev Page • →/l: Next Page • p: Public • P: Private • d: Delete • enter: Details • esc: Back")
	s.WriteString("\n" + help)

	// Error/Success messages
	if m.errorMessage != "" {
		s.WriteString("\n" + errorStyle.Render(m.errorMessage))
	} else if m.successMessage != "" {
		s.WriteString("\n" + successStyle.Render(m.successMessage))
	}

	return s.String()
}

// renderUserDetail renders detailed view of a selected user
func (m Model) renderUserDetail() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("User Details") + "\n\n")

	if m.selectedUserIndex < 0 || m.selectedUserIndex >= len(m.users) {
		s.WriteString(errorStyle.Render("Invalid user selection\n"))
		return s.String()
	}

	user := m.users[m.selectedUserIndex]

	// Display user details in a nice format
	s.WriteString(tableHeaderStyle.Render("Access Key") + "\n")
	s.WriteString(tableCellStyle.Render(user.Access) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("Secret Key") + "\n")
	s.WriteString(tableCellStyle.Render(user.Secret) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("Role") + "\n")
	s.WriteString(tableCellStyle.Render(user.Role) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("User ID") + "\n")
	s.WriteString(tableCellStyle.Render(fmt.Sprintf("%d", user.UserID)) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("Group ID") + "\n")
	s.WriteString(tableCellStyle.Render(fmt.Sprintf("%d", user.GroupID)) + "\n\n")

	if user.ProjectID > 0 {
		s.WriteString(tableHeaderStyle.Render("Project ID") + "\n")
		s.WriteString(tableCellStyle.Render(fmt.Sprintf("%d", user.ProjectID)) + "\n\n")
	}

	// Help text
	help := helpStyle.Render("c: Copy Credentials • esc/q: Back to list")
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

// renderBucketDetail renders detailed view of a selected bucket
func (m Model) renderBucketDetail() string {
	var s strings.Builder

	// Title
	s.WriteString(titleStyle.Render("Bucket Details") + "\n\n")

	if m.selectedBucketIndex < 0 || m.selectedBucketIndex >= len(m.buckets) {
		s.WriteString(errorStyle.Render("Invalid bucket selection\n"))
		return s.String()
	}

	bucket := m.buckets[m.selectedBucketIndex]

	// Display bucket details
	s.WriteString(tableHeaderStyle.Render("Bucket Name") + "\n")
	s.WriteString(tableCellStyle.Render(bucket.Name) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("Mountpoint") + "\n")
	s.WriteString(tableCellStyle.Render(bucket.Mountpoint) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("Owner") + "\n")
	s.WriteString(tableCellStyle.Render(bucket.Owner) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("Quota") + "\n")
	s.WriteString(tableCellStyle.Render(bucket.Quota) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("Used Space") + "\n")
	s.WriteString(tableCellStyle.Render(bucket.Used) + "\n\n")

	s.WriteString(tableHeaderStyle.Render("Available Space") + "\n")
	s.WriteString(tableCellStyle.Render(bucket.Available) + "\n\n")

	// Help text
	help := helpStyle.Render("esc/q: Back to list")
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

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
