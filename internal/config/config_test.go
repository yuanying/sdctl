package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuanying/sdctl/internal/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.Default()
	if cfg.URL != "http://localhost:7860" {
		t.Errorf("unexpected default URL: %s", cfg.URL)
	}
}

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	os.WriteFile(configFile, []byte("url: http://myserver:7860\n"), 0644)

	cfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.URL != "http://myserver:7860" {
		t.Errorf("unexpected URL: %s", cfg.URL)
	}
}

func TestEnvVarOverride(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	os.WriteFile(configFile, []byte("url: http://myserver:7860\n"), 0644)

	t.Setenv("SDCTL_URL", "http://envserver:7860")

	cfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.URL != "http://envserver:7860" {
		t.Errorf("env var should override file, got: %s", cfg.URL)
	}
}

func TestMissingFileUsesDefault(t *testing.T) {
	cfg, err := config.Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.URL != "http://localhost:7860" {
		t.Errorf("unexpected URL: %s", cfg.URL)
	}
}
