package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName = ".worktree.yml"
	DefaultGitDir  = ".bare"
)

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrInvalidConfig  = errors.New("invalid config")
)

type Config struct {
	Version  int      `yaml:"version"`
	GitDir   string   `yaml:"git_dir"`
	Setup    []string `yaml:"setup,omitempty"`
	Teardown []string `yaml:"teardown,omitempty"`
	Editor   string   `yaml:"editor,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		Version: 1,
		GitDir:  DefaultGitDir,
	}
}

func Load(projectRoot string) (*Config, error) {
	path := filepath.Join(projectRoot, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Join(ErrInvalidConfig, err)
	}

	return &cfg, nil
}

func (c *Config) Save(projectRoot string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	path := filepath.Join(projectRoot, ConfigFileName)
	return os.WriteFile(path, data, 0o644)
}

func Exists(projectRoot string) bool {
	path := filepath.Join(projectRoot, ConfigFileName)
	_, err := os.Stat(path)
	return err == nil
}

// renderAnnotatedConfig builds YAML with documentation comments.
// If cfg is nil, defaults are used with optional fields commented out.
// If cfg is non-nil, existing values are rendered uncommented.
func renderAnnotatedConfig(cfg *Config) string {
	var b strings.Builder

	b.WriteString("# wt - worktree project configuration\n")
	b.WriteString("# https://github.com/bkildow/wt-cli\n\n")

	b.WriteString("# Config schema version (do not change)\n")
	if cfg != nil {
		fmt.Fprintf(&b, "version: %d\n", cfg.Version)
	} else {
		b.WriteString("version: 1\n")
	}

	b.WriteString("\n# Path to the bare git repository\n")
	if cfg != nil {
		fmt.Fprintf(&b, "git_dir: %s\n", cfg.GitDir)
	} else {
		fmt.Fprintf(&b, "git_dir: %s\n", DefaultGitDir)
	}

	b.WriteString("\n# Editor for 'wt open' (e.g. cursor, code, zed)\n")
	b.WriteString("# Falls back to $EDITOR, then auto-detects\n")
	if cfg != nil && cfg.Editor != "" {
		fmt.Fprintf(&b, "editor: %s\n", cfg.Editor)
	} else {
		b.WriteString("# editor: cursor\n")
	}

	b.WriteString("\n# Commands to run after creating a new worktree\n")
	if cfg != nil && len(cfg.Setup) > 0 {
		b.WriteString("setup:\n")
		for _, s := range cfg.Setup {
			fmt.Fprintf(&b, "  - %s\n", yamlQuote(s))
		}
	} else {
		b.WriteString("# setup:\n")
		b.WriteString("#   - npm install\n")
		b.WriteString("#   - cp .env.example .env\n")
	}

	b.WriteString("\n# Commands to run before removing a worktree\n")
	if cfg != nil && len(cfg.Teardown) > 0 {
		b.WriteString("teardown:\n")
		for _, t := range cfg.Teardown {
			fmt.Fprintf(&b, "  - %s\n", yamlQuote(t))
		}
	} else {
		b.WriteString("# teardown:\n")
		b.WriteString("#   - docker compose down\n")
	}

	return b.String()
}

// yamlQuote wraps a string in double quotes if it contains characters
// that need quoting in YAML, otherwise returns it bare.
func yamlQuote(s string) string {
	if strings.ContainsAny(s, ":{}[]&*?|>!%#`@,\"'\\$\n") || s == "" {
		return fmt.Sprintf("%q", s)
	}
	return s
}

// WriteAnnotated writes a fresh .worktree.yml with default values and
// documentation comments for every field.
func WriteAnnotated(projectRoot string) error {
	content := renderAnnotatedConfig(nil)
	path := filepath.Join(projectRoot, ConfigFileName)
	return os.WriteFile(path, []byte(content), 0o644)
}

// WriteAnnotatedWithValues writes .worktree.yml using the given config's
// values, with documentation comments for every field.
func WriteAnnotatedWithValues(projectRoot string, cfg *Config) error {
	content := renderAnnotatedConfig(cfg)
	path := filepath.Join(projectRoot, ConfigFileName)
	return os.WriteFile(path, []byte(content), 0o644)
}
