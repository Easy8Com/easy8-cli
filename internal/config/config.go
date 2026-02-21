package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
)

type Defaults struct {
	ProjectID    int `json:"project_id"`
	TrackerID    int `json:"tracker_id"`
	StatusID     int `json:"status_id"`
	PriorityID   int `json:"priority_id"`
	AuthorID     int `json:"author_id"`
	AssignedToID int `json:"assigned_to_id"`
}

type Config struct {
	BaseURL  string   `json:"base_url"`
	APIKey   string   `json:"api_key"`
	Defaults Defaults `json:"defaults"`
}

func Load() (Config, error) {
	cfg := Config{
		BaseURL: "https://demo.easysoftware.com",
	}

	if fileCfg, err := readFileConfig(); err == nil {
		cfg = mergeConfig(cfg, fileCfg)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	applyEnv(&cfg)
	return cfg, nil
}

func readFileConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "easy8", "config.json"), nil
}

func applyEnv(cfg *Config) {
	if base := os.Getenv("EASY8_BASE_URL"); base != "" {
		cfg.BaseURL = base
	}
	if key := os.Getenv("EASY8_API_KEY"); key != "" {
		cfg.APIKey = key
	}

	setIntEnv(&cfg.Defaults.ProjectID, "EASY8_DEFAULT_PROJECT_ID")
	setIntEnv(&cfg.Defaults.TrackerID, "EASY8_DEFAULT_TRACKER_ID")
	setIntEnv(&cfg.Defaults.StatusID, "EASY8_DEFAULT_STATUS_ID")
	setIntEnv(&cfg.Defaults.PriorityID, "EASY8_DEFAULT_PRIORITY_ID")
	setIntEnv(&cfg.Defaults.AuthorID, "EASY8_DEFAULT_AUTHOR_ID")
	setIntEnv(&cfg.Defaults.AssignedToID, "EASY8_DEFAULT_ASSIGNED_TO_ID")
}

func setIntEnv(target *int, key string) {
	value := os.Getenv(key)
	if value == "" {
		return
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return
	}
	*target = parsed
}

func mergeConfig(base Config, overlay Config) Config {
	if overlay.BaseURL != "" {
		base.BaseURL = overlay.BaseURL
	}
	if overlay.APIKey != "" {
		base.APIKey = overlay.APIKey
	}

	if overlay.Defaults.ProjectID != 0 {
		base.Defaults.ProjectID = overlay.Defaults.ProjectID
	}
	if overlay.Defaults.TrackerID != 0 {
		base.Defaults.TrackerID = overlay.Defaults.TrackerID
	}
	if overlay.Defaults.StatusID != 0 {
		base.Defaults.StatusID = overlay.Defaults.StatusID
	}
	if overlay.Defaults.PriorityID != 0 {
		base.Defaults.PriorityID = overlay.Defaults.PriorityID
	}
	if overlay.Defaults.AuthorID != 0 {
		base.Defaults.AuthorID = overlay.Defaults.AuthorID
	}
	if overlay.Defaults.AssignedToID != 0 {
		base.Defaults.AssignedToID = overlay.Defaults.AssignedToID
	}

	return base
}
