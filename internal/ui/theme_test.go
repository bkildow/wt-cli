package ui

import (
	"bytes"
	"testing"

	lipgloss "charm.land/lipgloss/v2"
)

// defaultInfo and catppuccinInfo are derived from the registry so tests
// stay in sync without duplicating hex literals.
var (
	defaultInfo    = lipgloss.Color(themes["default"].Info)
	catppuccinInfo = lipgloss.Color(themes["catppuccin-mocha"].Info)
)

func TestApplyTheme(t *testing.T) {
	t.Cleanup(func() { ApplyTheme("default") })

	ApplyTheme("catppuccin-mocha")

	if ColorInfo == defaultInfo {
		t.Error("expected ColorInfo to change from default after applying catppuccin-mocha")
	}
	if ColorInfo != catppuccinInfo {
		t.Errorf("expected ColorInfo = %v, got %v", catppuccinInfo, ColorInfo)
	}

	// Restore default and verify it reverts.
	ApplyTheme("default")
	if ColorInfo != defaultInfo {
		t.Errorf("expected ColorInfo = %v after restoring default, got %v", defaultInfo, ColorInfo)
	}
}

func TestApplyThemeUnknown(t *testing.T) {
	origOutput := Output
	var buf bytes.Buffer
	Output = &buf
	t.Cleanup(func() {
		Output = origOutput
		ApplyTheme("default")
	})

	ApplyTheme("bogus")

	out := buf.String()
	if out == "" {
		t.Fatal("expected warning output for unknown theme")
	}
	if !bytes.Contains([]byte(out), []byte("bogus")) {
		t.Errorf("warning should mention the unknown theme name, got: %s", out)
	}
	if !bytes.Contains([]byte(out), []byte("catppuccin-mocha")) {
		t.Errorf("warning should list available themes, got: %s", out)
	}

	// Should fall back to default.
	if ColorInfo != defaultInfo {
		t.Errorf("expected fallback to default ColorInfo, got %v", ColorInfo)
	}
}

func TestThemeNames(t *testing.T) {
	names := ThemeNames()
	if len(names) < 9 {
		t.Fatalf("expected at least 9 themes, got %d", len(names))
	}

	// Verify sorted order.
	for i := 1; i < len(names); i++ {
		if names[i] < names[i-1] {
			t.Errorf("theme names not sorted: %v", names)
			break
		}
	}

	// Verify both built-in themes are present.
	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"default", "catppuccin-mocha", "snazzy", "dracula", "nord", "gruvbox"} {
		if !found[want] {
			t.Errorf("expected theme %q in list %v", want, names)
		}
	}
}
