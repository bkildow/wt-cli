package ui

import "charm.land/lipgloss/v2"

// defaultTheme is a package-level reference to avoid map lookups at init time.
// These vars are reassigned at runtime by ApplyTheme().
var dt = themes["default"]

var (
	ColorSuccess = lipgloss.Color(dt.Success)
	ColorError   = lipgloss.Color(dt.Error)
	ColorWarning = lipgloss.Color(dt.Warning)
	ColorInfo    = lipgloss.Color(dt.Info)
	ColorMuted   = lipgloss.Color(dt.Muted)

	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleError   = lipgloss.NewStyle().Foreground(ColorError)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleInfo    = lipgloss.NewStyle().Foreground(ColorInfo)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleCommand = lipgloss.NewStyle().Foreground(ColorMuted).Italic(true)
	StylePath    = lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)
	StyleHeading = lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)
)
