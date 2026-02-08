package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectType(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  string
	}{
		{
			name:  "drupal project",
			files: map[string]string{"composer.json": `{"require": {"drupal/core": "^10"}}`},
			want:  "drupal",
		},
		{
			name:  "composer without drupal",
			files: map[string]string{"composer.json": `{"require": {"laravel/framework": "^10"}}`},
			want:  "generic",
		},
		{
			name:  "node project",
			files: map[string]string{"package.json": `{}`},
			want:  "node",
		},
		{
			name:  "go project",
			files: map[string]string{"go.mod": "module example.com/foo"},
			want:  "go",
		},
		{
			name:  "rust project",
			files: map[string]string{"Cargo.toml": "[package]"},
			want:  "rust",
		},
		{
			name:  "python pyproject",
			files: map[string]string{"pyproject.toml": "[project]"},
			want:  "python",
		},
		{
			name:  "python requirements",
			files: map[string]string{"requirements.txt": "flask"},
			want:  "python",
		},
		{
			name:  "empty directory",
			files: map[string]string{},
			want:  "generic",
		},
		{
			name:  "node takes priority over go",
			files: map[string]string{"package.json": `{}`, "go.mod": "module foo"},
			want:  "node",
		},
		{
			name:  "drupal takes priority over node",
			files: map[string]string{"composer.json": `{"require": {"drupal/core": "^10"}}`, "package.json": `{}`},
			want:  "drupal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			got := DetectProjectType(dir)
			if got != tt.want {
				t.Errorf("DetectProjectType() = %q, want %q", got, tt.want)
			}
		})
	}
}
