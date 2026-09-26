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

// unsetEnv unsets key for the duration of the test.
func unsetEnv(t *testing.T, key string) {
	t.Helper()
	t.Setenv(key, "")
	os.Unsetenv(key)
}

func TestLoadFromFile(t *testing.T) {
	unsetEnv(t, "SDCTL_URL")
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
	unsetEnv(t, "SDCTL_URL")
	cfg, err := config.Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.URL != "http://localhost:7860" {
		t.Errorf("unexpected URL: %s", cfg.URL)
	}
}

func TestDefaultConfigHasNoParamsOrOutputDir(t *testing.T) {
	unsetEnv(t, "SDCTL_PARAMS")
	unsetEnv(t, "SDCTL_OUTPUT_DIR")

	cfg, err := config.Load("/nonexistent/config.yaml")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Params != "" {
		t.Errorf("expected empty params, got: %s", cfg.Params)
	}
	if cfg.OutputDir != "" {
		t.Errorf("expected empty output dir, got: %s", cfg.OutputDir)
	}
}

func TestLoadParamsAndOutputDirFromFile(t *testing.T) {
	unsetEnv(t, "SDCTL_PARAMS")
	unsetEnv(t, "SDCTL_OUTPUT_DIR")

	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	os.WriteFile(configFile, []byte("params: /etc/sdctl/file.yaml\noutput_dir: /tmp/file-out\n"), 0644)

	cfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Params != "/etc/sdctl/file.yaml" {
		t.Errorf("unexpected params: %s", cfg.Params)
	}
	if cfg.OutputDir != "/tmp/file-out" {
		t.Errorf("unexpected output dir: %s", cfg.OutputDir)
	}
}

func TestParamsAndOutputDirEnvOverrideFile(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	os.WriteFile(configFile, []byte("params: /etc/sdctl/file.yaml\noutput_dir: /tmp/file-out\n"), 0644)

	t.Setenv("SDCTL_PARAMS", "/etc/sdctl/env.yaml")
	t.Setenv("SDCTL_OUTPUT_DIR", "/tmp/env-out")

	cfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Params != "/etc/sdctl/env.yaml" {
		t.Errorf("env var should override file params, got: %s", cfg.Params)
	}
	if cfg.OutputDir != "/tmp/env-out" {
		t.Errorf("env var should override file output dir, got: %s", cfg.OutputDir)
	}
}

func TestEmptyEnvDisablesFileDefaults(t *testing.T) {
	dir := t.TempDir()
	configFile := filepath.Join(dir, "config.yaml")
	os.WriteFile(configFile, []byte("params: /etc/sdctl/file.yaml\noutput_dir: /tmp/file-out\n"), 0644)

	t.Setenv("SDCTL_PARAMS", "")
	t.Setenv("SDCTL_OUTPUT_DIR", "")

	cfg, err := config.Load(configFile)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if cfg.Params != "" {
		t.Errorf("empty SDCTL_PARAMS should disable file params, got: %s", cfg.Params)
	}
	if cfg.OutputDir != "" {
		t.Errorf("empty SDCTL_OUTPUT_DIR should disable file output dir, got: %s", cfg.OutputDir)
	}
}
