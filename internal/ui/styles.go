package ui

import "github.com/charmbracelet/lipgloss"

var (
	// Primary colors matching Singapore GP UI
	brightGreen = lipgloss.Color("#00FF87")  // Bright green for highlights
	darkBg      = lipgloss.Color("#0A0E1A")  // Very dark background
	mediumGray  = lipgloss.Color("#8B95A5")  // Muted gray for secondary text
	white       = lipgloss.Color("#FFFFFF")  // White for primary text
	darkGreen   = lipgloss.Color("#1A3A2E")  // Dark green for cards/borders
)

var (
	TitleStyle    = lipgloss.NewStyle().Bold(true).Foreground(brightGreen).MarginBottom(1)
	GPStyle       = lipgloss.NewStyle().Bold(true).Foreground(white).MarginBottom(1)
	CardStyle     = lipgloss.NewStyle().Padding(1, 2).Border(lipgloss.RoundedBorder()).BorderForeground(brightGreen).Background(darkGreen)
	LabelStyle    = lipgloss.NewStyle().Foreground(mediumGray)
	LiveBadge     = lipgloss.NewStyle().Bold(true).Foreground(darkBg).Background(brightGreen).Padding(0, 1)
	ErrorStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	RoundBadge    = lipgloss.NewStyle().Bold(true).Foreground(darkBg).Background(brightGreen).Padding(0, 1)

	// Style for circuit ASCII art
	CircuitStyle  = lipgloss.NewStyle().Foreground(brightGreen)
)
