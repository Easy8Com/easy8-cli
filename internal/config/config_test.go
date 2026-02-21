package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("EASY8_BASE_URL", "")
	t.Setenv("EASY8_API_KEY", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.BaseURL != "https://demo.easysoftware.com" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
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

func TestLoadConfigFileMerge(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path := filepath.Join(home, ".config", "easy8")
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	fileCfg := Config{
		BaseURL: "https://from-config",
		APIKey:  "config-key",
		Defaults: Defaults{
			ProjectID:    1,
			TrackerID:    2,
			StatusID:     3,
			PriorityID:   4,
			AuthorID:     5,
			AssignedToID: 6,
		},
	}

	data, err := json.Marshal(fileCfg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if err := os.WriteFile(filepath.Join(path, "config.json"), data, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("EASY8_BASE_URL", "https://from-env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load error: %v", err)
	}
	if cfg.BaseURL != "https://from-env" {
		t.Fatalf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.APIKey != "config-key" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}
	if cfg.Defaults.ProjectID != 1 {
		t.Fatalf("ProjectID = %d", cfg.Defaults.ProjectID)
	}
}
