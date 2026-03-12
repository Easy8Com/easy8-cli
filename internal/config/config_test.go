package config

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoadDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("EASY8_BASE_URL", "")
	t.Setenv("EASY8_API_KEY", "")

	project := t.TempDir()
	setWorkingDir(t, project)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.BaseURL != "https://demo.easy8.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("EASY8_BASE_URL", "https://example.com")
	t.Setenv("EASY8_API_KEY", "abc")
	t.Setenv("EASY8_DEFAULT_PROJECT_ID", "10")
	t.Setenv("EASY8_DEFAULT_TRACKER_ID", "11")
	t.Setenv("EASY8_DEFAULT_STATUS_ID", "12")
	t.Setenv("EASY8_DEFAULT_PRIORITY_ID", "13")
	t.Setenv("EASY8_DEFAULT_AUTHOR_ID", "14")
	t.Setenv("EASY8_DEFAULT_ASSIGNED_TO_ID", "15")

	project := t.TempDir()
	setWorkingDir(t, project)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "abc" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}
	if cfg.Defaults.ProjectID != 10 {
		t.Fatalf("ProjectID = %d", cfg.Defaults.ProjectID)
	}
	if cfg.Defaults.TrackerID != 11 {
		t.Fatalf("TrackerID = %d", cfg.Defaults.TrackerID)
	}
	if cfg.Defaults.StatusID != 12 {
		t.Fatalf("StatusID = %d", cfg.Defaults.StatusID)
	}
	if cfg.Defaults.PriorityID != 13 {
		t.Fatalf("PriorityID = %d", cfg.Defaults.PriorityID)
	}
	if cfg.Defaults.AuthorID != 14 {
		t.Fatalf("AuthorID = %d", cfg.Defaults.AuthorID)
	}
	if cfg.Defaults.AssignedToID != 15 {
		t.Fatalf("AssignedToID = %d", cfg.Defaults.AssignedToID)
	}
}

func TestInvalidIntEnvWarning(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("EASY8_BASE_URL", "")
	t.Setenv("EASY8_API_KEY", "")
	t.Setenv("EASY8_DEFAULT_PROJECT_ID", "abc")
	t.Setenv("EASY8_DEFAULT_TRACKER_ID", "")

	var cfg Config
	warnings := applyEnv(&cfg)
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if warnings[0] != `invalid integer for EASY8_DEFAULT_PROJECT_ID: "abc"` {
		t.Fatalf("unexpected warning: %s", warnings[0])
	}
	if cfg.Defaults.ProjectID != 0 {
		t.Fatalf("ProjectID should be 0, got %d", cfg.Defaults.ProjectID)
	}
}

func TestLoadYAMLMerge(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("EASY8_BASE_URL", "")
	t.Setenv("EASY8_API_KEY", "")

	global := Config{
		BaseURL: "https://global.example.com",
		APIKey:  "global-key",
		Defaults: Defaults{
			ProjectID:    1,
			TrackerID:    2,
			StatusID:     3,
			PriorityID:   4,
			AuthorID:     5,
			AssignedToID: 6,
		},
	}
	writeConfigFile(t, filepath.Join(home, ".config", "easy8", "config.yaml"), global)

	project := t.TempDir()
	writeConfigFile(t, filepath.Join(project, ".easy8.yaml"), Config{
		BaseURL: "https://local.example.com",
		Defaults: Defaults{
			ProjectID: 10,
		},
	})

	nested := filepath.Join(project, "nested", "repo")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}
	setWorkingDir(t, nested)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}

	if cfg.BaseURL != "https://local.example.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "global-key" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}
	if cfg.Defaults.ProjectID != 10 {
		t.Fatalf("ProjectID = %d", cfg.Defaults.ProjectID)
	}
	if cfg.Defaults.TrackerID != 2 {
		t.Fatalf("TrackerID = %d", cfg.Defaults.TrackerID)
	}
}

func TestEnvOverridesYAML(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("EASY8_BASE_URL", "https://env.example.com")
	t.Setenv("EASY8_API_KEY", "env-key")

	global := Config{
		BaseURL: "https://global.example.com",
		APIKey:  "global-key",
	}
	writeConfigFile(t, filepath.Join(home, ".config", "easy8", "config.yaml"), global)

	project := t.TempDir()
	writeConfigFile(t, filepath.Join(project, ".easy8.yaml"), Config{
		BaseURL: "https://local.example.com",
		APIKey:  "local-key",
	})
	setWorkingDir(t, project)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.BaseURL != "https://env.example.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "env-key" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}
}

func TestSaveGlobal(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := SaveGlobal(Config{BaseURL: "https://saved.example.com", APIKey: "saved-key"})
	if err != nil {
		t.Fatalf("SaveGlobal error: %v", err)
	}

	expected := filepath.Join(home, ".config", "easy8", "config.yaml")
	if path != expected {
		t.Fatalf("path = %q, want %q", path, expected)
	}

	loaded := readConfigFile(t, expected)
	if loaded.BaseURL != "https://saved.example.com" {
		t.Fatalf("BaseURL = %q", loaded.BaseURL)
	}
	if loaded.APIKey != "saved-key" {
		t.Fatalf("APIKey = %q", loaded.APIKey)
	}
}

func TestSaveLocal(t *testing.T) {
	project := t.TempDir()
	setWorkingDir(t, project)

	path, err := SaveLocal(Config{BaseURL: "https://local-save.example.com", APIKey: "local-key"})
	if err != nil {
		t.Fatalf("SaveLocal error: %v", err)
	}

	expected := filepath.Join(project, ".easy8.yaml")
	if path != expected {
		t.Fatalf("path = %q, want %q", path, expected)
	}

	loaded := readConfigFile(t, expected)
	if loaded.BaseURL != "https://local-save.example.com" {
		t.Fatalf("BaseURL = %q", loaded.BaseURL)
	}
	if loaded.APIKey != "local-key" {
		t.Fatalf("APIKey = %q", loaded.APIKey)
	}
}

func setWorkingDir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(old)
	})
}

func writeConfigFile(t *testing.T, path string, cfg Config) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func readConfigFile(t *testing.T, path string) Config {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return cfg
}
