// Package ui provides terminal output helpers and interactive prompts.
package ui

import (
	"io"
	"os"

	lipgloss "charm.land/lipgloss/v2"
)

var (
	Output  io.Writer = os.Stderr
	Verbose bool
)

func Success(msg string) {
	_, _ = lipgloss.Fprintln(Output, StyleSuccess.Render("✓ "+msg))
}

func Error(msg string) {
	_, _ = lipgloss.Fprintln(Output, StyleError.Render("✗ "+msg))
}

func Warning(msg string) {
	_, _ = lipgloss.Fprintln(Output, StyleWarning.Render("⚠ "+msg))
}

func Info(msg string) {
	_, _ = lipgloss.Fprintln(Output, StyleInfo.Render(msg))
}

func Step(msg string) {
	_, _ = lipgloss.Fprintln(Output, StyleInfo.Render("→ "+msg))
}

func DryRunNotice(action string) {
	_, _ = lipgloss.Fprintln(Output, StyleMuted.Render("[dry-run] "+action))
}

func Command(cmd string) {
	if !Verbose {
		return
	}
	_, _ = lipgloss.Fprintln(Output, StyleCommand.Render("  $ "+cmd))
}

func Heading(msg string) {
	_, _ = lipgloss.Fprintln(Output, StyleHeading.Render(msg))
}
