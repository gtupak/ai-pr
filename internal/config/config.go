package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	localFileName  = ".aipr.json"
	globalDirName  = ".aipr"
	globalFileName = "config.json"
)

type Config struct {
	Base string `json:"base"`
}

func Path(repoRoot string) string {
	return filepath.Join(repoRoot, localFileName)
}

func Load(repoRoot string) (Config, error) {
	cfgPath := Path(repoRoot)
	b, err := os.ReadFile(cfgPath)
	if errors.Is(err, os.ErrNotExist) {
		return Config{}, nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", cfgPath, err)
	}

	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", cfgPath, err)
	}
	cfg.Base = strings.TrimSpace(cfg.Base)
	return cfg, nil
}

func Save(repoRoot string, cfg Config) error {
	cfg.Base = strings.TrimSpace(cfg.Base)

	cfgPath := Path(repoRoot)
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	b = append(b, '\n')

	if err := os.WriteFile(cfgPath, b, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", cfgPath, err)
	}
	return nil
}

type GlobalConfig struct {
	OpenRouterAPIKey string `json:"openrouter_api_key"`
}

func GlobalPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, globalDirName, globalFileName), nil
}

func LoadGlobal() (GlobalConfig, error) {
	cfgPath, err := GlobalPath()
	if err != nil {
		return GlobalConfig{}, err
	}

	b, err := os.ReadFile(cfgPath)
	if errors.Is(err, os.ErrNotExist) {
		return GlobalConfig{}, nil
	}
	if err != nil {
		return GlobalConfig{}, fmt.Errorf("read %s: %w", cfgPath, err)
	}

	var cfg GlobalConfig
	if err := json.Unmarshal(b, &cfg); err != nil {
		return GlobalConfig{}, fmt.Errorf("parse %s: %w", cfgPath, err)
	}
	cfg.OpenRouterAPIKey = strings.TrimSpace(cfg.OpenRouterAPIKey)
	return cfg, nil
}

func SaveGlobal(cfg GlobalConfig) error {
	cfgPath, err := GlobalPath()
	if err != nil {
		return err
	}

	cfg.OpenRouterAPIKey = strings.TrimSpace(cfg.OpenRouterAPIKey)
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal global config: %w", err)
	}
	b = append(b, '\n')

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return fmt.Errorf("create config directory %s: %w", filepath.Dir(cfgPath), err)
	}
	if err := os.WriteFile(cfgPath, b, 0o600); err != nil {
		return fmt.Errorf("write %s: %w", cfgPath, err)
	}
	return nil
}
