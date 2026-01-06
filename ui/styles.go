package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Color scheme
	primaryColor   = lipgloss.Color("#00D9FF")
	secondaryColor = lipgloss.Color("#7D56F4")
	accentColor    = lipgloss.Color("#FF6B9D")
	successColor   = lipgloss.Color("#50FA7B")
	errorColor     = lipgloss.Color("#FF5555")
	warningColor   = lipgloss.Color("#FFB86C")
	textColor      = lipgloss.Color("#F8F8F2")
	dimColorVal    = lipgloss.Color("#6272A4")

	// Base styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1).
			Padding(0, 1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Italic(true).
			MarginBottom(1)

	menuItemStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 2)

	selectedMenuItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true).
				Padding(0, 2).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)

	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(accentColor).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(dimColorVal)

	tableCellStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Padding(0, 1)

	selectedTableRowStyle = lipgloss.NewStyle().
				Background(secondaryColor).
				Foreground(textColor).
				Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(dimColorVal)

	helpStyle = lipgloss.NewStyle().
			Foreground(dimColorVal).
			MarginTop(1).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor)

	successStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Padding(1, 2).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(successColor)

	inputLabelStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			MarginRight(1)

	inputStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(dimColorVal).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	buttonStyle = lipgloss.NewStyle().
			Foreground(textColor).
			Background(secondaryColor).
			Padding(0, 3).
			MarginTop(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor)

	focusedButtonStyle = lipgloss.NewStyle().
				Foreground(textColor).
				Background(primaryColor).
				Bold(true).
				Padding(0, 3).
				MarginTop(1).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor)
)
