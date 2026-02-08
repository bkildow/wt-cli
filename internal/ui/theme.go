package ui

import (
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// WtTheme returns a Catppuccin Mocha themed huh theme.
func WtTheme() *huh.Theme {
	t := huh.ThemeBase()
	m := catppuccin.Mocha

	t.Focused.Title = t.Focused.Title.Foreground(lipgloss.Color(m.Mauve().Hex)).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(lipgloss.Color(m.Subtext0().Hex))
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(lipgloss.Color(m.Sapphire().Hex))
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(lipgloss.Color(m.Peach().Hex))
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(lipgloss.Color(m.Text().Hex))
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(lipgloss.Color(m.Sapphire().Hex))
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(lipgloss.Color(m.Sapphire().Hex))
	t.Focused.Base = t.Focused.Base.BorderForeground(lipgloss.Color(m.Surface1().Hex))

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(lipgloss.Color(m.Rosewater().Hex))
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(lipgloss.Color(m.Overlay0().Hex))
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(lipgloss.Color(m.Sapphire().Hex))

	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(lipgloss.Color(m.Red().Hex))
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(lipgloss.Color(m.Red().Hex))

	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())

	return t
}
