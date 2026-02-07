package ui

import (
	"fmt"
	"io"
	"os"
)

var Output io.Writer = os.Stderr

func Success(msg string) {
	fmt.Fprintln(Output, StyleSuccess.Render("✓ "+msg))
}

func Error(msg string) {
	fmt.Fprintln(Output, StyleError.Render("✗ "+msg))
}

func Warning(msg string) {
	fmt.Fprintln(Output, StyleWarning.Render("⚠ "+msg))
}

func Info(msg string) {
	fmt.Fprintln(Output, StyleInfo.Render(msg))
}

func Step(msg string) {
	fmt.Fprintln(Output, StyleInfo.Render("→ "+msg))
}

func DryRunNotice(action string) {
	fmt.Fprintln(Output, StyleMuted.Render("[dry-run] "+action))
}

func Command(cmd string) {
	fmt.Fprintln(Output, StyleCommand.Render("  $ "+cmd))
}
