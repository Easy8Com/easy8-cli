package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"gopkg.in/yaml.v3"
)

const (
	defaultBaseURL  = "https://demo.easy8.com"
	localConfigFile = ".easy8.yaml"
)

type Defaults struct {
	ProjectID    int `yaml:"project_id"`
	TrackerID    int `yaml:"tracker_id"`
	StatusID     int `yaml:"status_id"`
	PriorityID   int `yaml:"priority_id"`
	AuthorID     int `yaml:"author_id"`
	AssignedToID int `yaml:"assigned_to_id"`
}

type Config struct {
	BaseURL       string   `yaml:"base_url"`
	APIKey        string   `yaml:"api_key"`
	AutoUpdate    bool     `yaml:"autoupdate"`
	autoUpdateSet bool     `yaml:"-"`
	Defaults      Defaults `yaml:"defaults"`
}

type fileConfig struct {
	BaseURL    string   `yaml:"base_url"`
	APIKey     string   `yaml:"api_key"`
	AutoUpdate *bool    `yaml:"autoupdate"`
	Defaults   Defaults `yaml:"defaults"`
}

func Load() (Config, error) {
	cfg := Config{
		BaseURL: defaultBaseURL,
	}

	if fileCfg, err := readGlobalConfig(); err == nil {
		cfg = mergeConfig(cfg, fileCfg)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	if fileCfg, err := readLocalConfig(); err == nil {
		cfg = mergeConfig(cfg, fileCfg)
	} else if !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}

	warnings := applyEnv(&cfg)
	for _, w := range warnings {
		fmt.Fprintf(os.Stderr, "config warning: %s\n", w)
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	return cfg, nil
}

func SaveGlobal(cfg Config) (string, error) {
	path, err := configPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", err
	}
	if err := writeConfig(path, cfg); err != nil {
		return "", err
	}
	return path, nil
}

func SaveLocal(cfg Config) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path := filepath.Join(wd, localConfigFile)
	if err := writeConfig(path, cfg); err != nil {
		return "", err
	}
	return path, nil
}

func readGlobalConfig() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	return readConfig(path)
}

func readLocalConfig() (Config, error) {
	path, err := localConfigPath()
	if err != nil {
		return Config{}, err
	}
	if path == "" {
		return Config{}, os.ErrNotExist
	}
	return readConfig(path)
}

func readConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	var raw fileConfig
	if err := decoder.Decode(&raw); err != nil {
		return Config{}, err
	}
	cfg := Config{
		BaseURL:  raw.BaseURL,
		APIKey:   raw.APIKey,
		Defaults: raw.Defaults,
	}
	if raw.AutoUpdate != nil {
		cfg.AutoUpdate = *raw.AutoUpdate
		cfg.autoUpdateSet = true
	}
	return cfg, nil
}

func writeConfig(path string, cfg Config) error {
	if cfg.BaseURL == "" {
		cfg.BaseURL = defaultBaseURL
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	if !bytes.HasSuffix(data, []byte("\n")) {
		data = append(data, '\n')
	}
	return os.WriteFile(path, data, 0o600)
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "easy8", "config.yaml"), nil
}

func localConfigPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		path := filepath.Join(dir, localConfigFile)
		_, err := os.Stat(path)
		if err == nil {
			return path, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", nil
}

func applyEnv(cfg *Config) []string {
	var warnings []string
	if base := os.Getenv("EASY8_BASE_URL"); base != "" {
		cfg.BaseURL = base
	}
	if key := os.Getenv("EASY8_API_KEY"); key != "" {
		cfg.APIKey = key
	}
	setBoolEnv(&cfg.AutoUpdate, &cfg.autoUpdateSet, "EASY8_AUTOUPDATE", &warnings)

	setIntEnv(&cfg.Defaults.ProjectID, "EASY8_DEFAULT_PROJECT_ID", &warnings)
	setIntEnv(&cfg.Defaults.TrackerID, "EASY8_DEFAULT_TRACKER_ID", &warnings)
	setIntEnv(&cfg.Defaults.StatusID, "EASY8_DEFAULT_STATUS_ID", &warnings)
	setIntEnv(&cfg.Defaults.PriorityID, "EASY8_DEFAULT_PRIORITY_ID", &warnings)
	setIntEnv(&cfg.Defaults.AuthorID, "EASY8_DEFAULT_AUTHOR_ID", &warnings)
	setIntEnv(&cfg.Defaults.AssignedToID, "EASY8_DEFAULT_ASSIGNED_TO_ID", &warnings)
	return warnings
}

func setBoolEnv(target *bool, set *bool, key string, warnings *[]string) {
	value := os.Getenv(key)
	if value == "" {
		return
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		*warnings = append(*warnings, fmt.Sprintf("invalid boolean for %s: %q", key, value))
		return
	}
	*target = parsed
	*set = true
}

func setIntEnv(target *int, key string, warnings *[]string) {
	value := os.Getenv(key)
	if value == "" {
		return
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		*warnings = append(*warnings, fmt.Sprintf("invalid integer for %s: %q", key, value))
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
	if overlay.autoUpdateSet {
		base.AutoUpdate = overlay.AutoUpdate
		base.autoUpdateSet = true
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
