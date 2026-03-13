package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const fileName = ".aipr.json"

type Config struct {
	Base string `json:"base"`
}

func Path(repoRoot string) string {
	return filepath.Join(repoRoot, fileName)
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
