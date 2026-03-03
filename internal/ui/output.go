// Package ui provides terminal output helpers and interactive prompts.
package ui

import (
	"io"
	"os"

	lipgloss "charm.land/lipgloss/v2"
)

var Output io.Writer = os.Stderr

func Success(msg string) {
	lipgloss.Fprintln(Output, StyleSuccess.Render("✓ "+msg))
}

func Error(msg string) {
	lipgloss.Fprintln(Output, StyleError.Render("✗ "+msg))
}

func Warning(msg string) {
	lipgloss.Fprintln(Output, StyleWarning.Render("⚠ "+msg))
}

func Info(msg string) {
	lipgloss.Fprintln(Output, StyleInfo.Render(msg))
}

func Step(msg string) {
	lipgloss.Fprintln(Output, StyleInfo.Render("→ "+msg))
}

func DryRunNotice(action string) {
	lipgloss.Fprintln(Output, StyleMuted.Render("[dry-run] "+action))
}

func Command(cmd string) {
	lipgloss.Fprintln(Output, StyleCommand.Render("  $ "+cmd))
}

func Heading(msg string) {
	lipgloss.Fprintln(Output, StyleHeading.Render(msg))
}
