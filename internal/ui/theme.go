package ui

import (
	"fmt"
	"sort"
	"strings"

	lipgloss "charm.land/lipgloss/v2"
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/huh"
	lipglossv1 "github.com/charmbracelet/lipgloss"
)

// theme holds the five semantic colors used throughout wt output.
type theme struct {
	Success string
	Error   string
	Warning string
	Info    string
	Muted   string
}

// themes is the built-in theme registry.
var themes = map[string]theme{
	"default": {
		Success: "#22c55e",
		Error:   "#ef4444",
		Warning: "#eab308",
		Info:    "#3b82f6",
		Muted:   "#6b7280",
	},
	"catppuccin": {
		Success: catppuccin.Mocha.Green().Hex,
		Error:   catppuccin.Mocha.Red().Hex,
		Warning: catppuccin.Mocha.Yellow().Hex,
		Info:    catppuccin.Mocha.Sapphire().Hex,
		Muted:   catppuccin.Mocha.Overlay0().Hex,
	},
}

// activeTheme tracks the currently applied theme name for WtTheme.
var activeTheme = "default"

// ApplyTheme looks up a theme by name and reassigns all Color*/Style* package
// vars. Unknown names print a warning and fall back to "default".
func ApplyTheme(name string) {
	t, ok := themes[name]
	if !ok {
		fmt.Fprintf(Output, "wt: unknown theme %q (available: %s)\n",
			name, strings.Join(ThemeNames(), ", "))
		name = "default"
		t = themes[name]
	}

	activeTheme = name

	ColorSuccess = lipgloss.Color(t.Success)
	ColorError = lipgloss.Color(t.Error)
	ColorWarning = lipgloss.Color(t.Warning)
	ColorInfo = lipgloss.Color(t.Info)
	ColorMuted = lipgloss.Color(t.Muted)

	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleError = lipgloss.NewStyle().Foreground(ColorError)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleInfo = lipgloss.NewStyle().Foreground(ColorInfo)
	StyleMuted = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleCommand = lipgloss.NewStyle().Foreground(ColorMuted).Italic(true)
	StylePath = lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)
	StyleHeading = lipgloss.NewStyle().Foreground(ColorInfo).Bold(true)
}

// ThemeNames returns a sorted list of available theme names.
func ThemeNames() []string {
	names := make([]string, 0, len(themes))
	for n := range themes {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// WtTheme returns a huh theme customised to match the active wt palette.
func WtTheme() *huh.Theme {
	t := huh.ThemeCharm()

	info := lipglossv1.Color(themes[activeTheme].Info)
	muted := lipglossv1.Color(themes[activeTheme].Muted)

	t.Focused.Title = t.Focused.Title.Foreground(info)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(info)
	t.Focused.Directory = t.Focused.Directory.Foreground(info)
	t.Focused.Description = t.Focused.Description.Foreground(muted)
	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}
