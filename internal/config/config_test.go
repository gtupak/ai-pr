package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMissingFileReturnsEmptyConfig(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	cfg, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	if cfg.Base != "" {
		t.Fatalf("expected empty base, got %q", cfg.Base)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	want := Config{Base: "develop"}

	if err := Save(repoRoot, want); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	got, err := Load(repoRoot)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got.Base != want.Base {
		t.Fatalf("expected base %q, got %q", want.Base, got.Base)
	}
}

func TestLoadMalformedConfig(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	cfgPath := filepath.Join(repoRoot, ".aipr.json")
	if err := os.WriteFile(cfgPath, []byte("{invalid"), 0o644); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	_, err := Load(repoRoot)
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}
