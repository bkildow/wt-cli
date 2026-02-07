package ui

import "github.com/charmbracelet/lipgloss"

var (
	ColorSuccess = lipgloss.Color("#22c55e")
	ColorError   = lipgloss.Color("#ef4444")
	ColorWarning = lipgloss.Color("#eab308")
	ColorInfo    = lipgloss.Color("#3b82f6")
	ColorMuted   = lipgloss.Color("#6b7280")

	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleError   = lipgloss.NewStyle().Foreground(ColorError)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleInfo    = lipgloss.NewStyle().Foreground(ColorInfo)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleCommand = lipgloss.NewStyle().Foreground(ColorMuted).Italic(true)
	StylePath    = lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)
)
