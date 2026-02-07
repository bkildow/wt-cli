package config

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = ".worktree.yml"
const DefaultGitDir = ".bare"

var (
	ErrConfigNotFound = errors.New("config file not found")
	ErrInvalidConfig  = errors.New("invalid config")
)

type Config struct {
	Version     int      `yaml:"version"`
	GitDir      string   `yaml:"git_dir"`
	PostCreate  []string `yaml:"post_create,omitempty"`
	ProjectType string   `yaml:"project_type,omitempty"`
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
	return os.WriteFile(path, data, 0644)
}

func Exists(projectRoot string) bool {
	path := filepath.Join(projectRoot, ConfigFileName)
	_, err := os.Stat(path)
	return err == nil
}
