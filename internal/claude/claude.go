// Package claude manages Claude Code hook configuration for wt projects.
package claude

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const settingsFile = ".claude/settings.local.json"

// Hook event names used by Claude Code.
const (
	HookWorktreeCreate = "WorktreeCreate"
	HookWorktreeRemove = "WorktreeRemove"
)

// ConfigureHooks writes WorktreeCreate and WorktreeRemove hooks into
// .claude/settings.local.json, deep-merging with any existing settings.
// wtBinary is the command to invoke (e.g. "wt" or "/opt/bin/wt").
func ConfigureHooks(projectRoot, wtBinary string) error {
	path := filepath.Join(projectRoot, settingsFile)

	existing, err := readSettings(path)
	if err != nil {
		return err
	}

	hooks := buildHooksConfig(wtBinary)
	existing["hooks"] = hooks

	return writeSettings(path, existing)
}

// IsHooksConfigured checks if Claude Code hooks are already present.
func IsHooksConfigured(projectRoot string) bool {
	path := filepath.Join(projectRoot, settingsFile)

	settings, err := readSettings(path)
	if err != nil {
		return false
	}

	hooks, ok := settings["hooks"]
	if !ok {
		return false
	}

	hooksMap, ok := hooks.(map[string]any)
	if !ok {
		return false
	}

	_, hasCreate := hooksMap[HookWorktreeCreate]
	_, hasRemove := hooksMap[HookWorktreeRemove]
	return hasCreate && hasRemove
}

// RemoveHooks removes the WorktreeCreate and WorktreeRemove hooks from
// .claude/settings.local.json, preserving other settings.
func RemoveHooks(projectRoot string) error {
	path := filepath.Join(projectRoot, settingsFile)

	settings, err := readSettings(path)
	if err != nil {
		return err
	}

	hooks, ok := settings["hooks"]
	if !ok {
		return nil
	}

	hooksMap, ok := hooks.(map[string]any)
	if !ok {
		return nil
	}

	delete(hooksMap, HookWorktreeCreate)
	delete(hooksMap, HookWorktreeRemove)

	if len(hooksMap) == 0 {
		delete(settings, "hooks")
	} else {
		settings["hooks"] = hooksMap
	}

	return writeSettings(path, settings)
}

func buildHooksConfig(wtBinary string) map[string]any {
	return map[string]any{
		HookWorktreeCreate: []any{
			map[string]any{
				"matcher": "",
				"hooks": []any{
					map[string]any{
						"type":    "command",
						"command": wtBinary + " claude hook-worktree-create",
					},
				},
			},
		},
		HookWorktreeRemove: []any{
			map[string]any{
				"matcher": "",
				"hooks": []any{
					map[string]any{
						"type":    "command",
						"command": wtBinary + " claude hook-worktree-remove",
					},
				},
			},
		},
	}
}

func readSettings(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, err
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return settings, nil
}

func writeSettings(path string, settings map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}
