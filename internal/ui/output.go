// Package ui provides terminal output helpers and interactive prompts.
package ui

import (
	"fmt"
	"io"
	"os"
	"time"

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

// FormatDuration formats a duration into a human-friendly string like
// "45 seconds" or "2 minutes 30 seconds".
func FormatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	s := int(d.Seconds())
	if s < 60 {
		if s == 1 {
			return "1 second"
		}
		return fmt.Sprintf("%d seconds", s)
	}
	m := s / 60
	rem := s % 60
	mUnit := "minutes"
	if m == 1 {
		mUnit = "minute"
	}
	if rem == 0 {
		return fmt.Sprintf("%d %s", m, mUnit)
	}
	sUnit := "seconds"
	if rem == 1 {
		sUnit = "second"
	}
	return fmt.Sprintf("%d %s %d %s", m, mUnit, rem, sUnit)
}
