package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetConfigByMenuID_NormalizesComboSwitcherKeys(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))

	configDir := filepath.Join(tempDir, ".config", "gmenu", comboSwitcherMenuID)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	yaml := strings.Join([]string{
		"MinWidth: 900",
		"NoNumericSelection: true",
	}, "\n")

	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := GetConfigByMenuID(comboSwitcherMenuID)
	if err != nil {
		t.Fatalf("GetConfigByMenuID returned error: %v", err)
	}

	if cfg.MinWidth != 900 {
		t.Fatalf("expected MinWidth 900, got %v", cfg.MinWidth)
	}
	if !cfg.NoNumericSelection {
		t.Fatalf("expected NoNumericSelection true")
	}
}

func TestGetConfigByMenuID_DetectsDuplicateComboKeys(t *testing.T) {
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, ".config"))

	configDir := filepath.Join(tempDir, ".config", "gmenu", comboSwitcherMenuID)
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("create config dir: %v", err)
	}

	yaml := strings.Join([]string{
		"min_width: 600",
		"MinWidth: 800",
	}, "\n")

	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	if _, err := GetConfigByMenuID(comboSwitcherMenuID); err == nil {
		t.Fatal("expected error due to duplicate key variants, got nil")
	}
}
