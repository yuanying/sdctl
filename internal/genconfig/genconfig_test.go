package genconfig_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yuanying/sdctl/internal/genconfig"
)

func TestLoadParamConfig_AllFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "params.yaml")
	content := `
negative_prompt: "bad quality"
steps: 30
width: 768
height: 768
cfg_scale: 8.5
sampler: "DPM++ 2M"
scheduler: "Karras"
seed: 12345
batch_count: 2
batch_size: 3
denoising_strength: 0.6
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := genconfig.LoadParamConfig(path)
	if err != nil {
		t.Fatalf("LoadParamConfig failed: %v", err)
	}

	if cfg.NegativePrompt == nil || *cfg.NegativePrompt != "bad quality" {
		t.Errorf("unexpected NegativePrompt: %v", cfg.NegativePrompt)
	}
	if cfg.Steps == nil || *cfg.Steps != 30 {
		t.Errorf("unexpected Steps: %v", cfg.Steps)
	}
	if cfg.Width == nil || *cfg.Width != 768 {
		t.Errorf("unexpected Width: %v", cfg.Width)
	}
	if cfg.Height == nil || *cfg.Height != 768 {
		t.Errorf("unexpected Height: %v", cfg.Height)
	}
	if cfg.CFGScale == nil || *cfg.CFGScale != 8.5 {
		t.Errorf("unexpected CFGScale: %v", cfg.CFGScale)
	}
	if cfg.Sampler == nil || *cfg.Sampler != "DPM++ 2M" {
		t.Errorf("unexpected Sampler: %v", cfg.Sampler)
	}
	if cfg.Scheduler == nil || *cfg.Scheduler != "Karras" {
		t.Errorf("unexpected Scheduler: %v", cfg.Scheduler)
	}
	if cfg.Seed == nil || *cfg.Seed != 12345 {
		t.Errorf("unexpected Seed: %v", cfg.Seed)
	}
	if cfg.BatchCount == nil || *cfg.BatchCount != 2 {
		t.Errorf("unexpected BatchCount: %v", cfg.BatchCount)
	}
	if cfg.BatchSize == nil || *cfg.BatchSize != 3 {
		t.Errorf("unexpected BatchSize: %v", cfg.BatchSize)
	}
	if cfg.DenoisingStrength == nil || *cfg.DenoisingStrength != 0.6 {
		t.Errorf("unexpected DenoisingStrength: %v", cfg.DenoisingStrength)
	}
}

func TestLoadParamConfig_PartialFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "params.yaml")
	os.WriteFile(path, []byte("steps: 50\n"), 0644)

	cfg, err := genconfig.LoadParamConfig(path)
	if err != nil {
		t.Fatalf("LoadParamConfig failed: %v", err)
	}

	if cfg.Steps == nil || *cfg.Steps != 50 {
		t.Errorf("unexpected Steps: %v", cfg.Steps)
	}
	if cfg.NegativePrompt != nil {
		t.Errorf("NegativePrompt should be nil when not specified, got: %v", cfg.NegativePrompt)
	}
	if cfg.Width != nil {
		t.Errorf("Width should be nil when not specified, got: %v", cfg.Width)
	}
}

func TestLoadParamConfig_MissingFile(t *testing.T) {
	_, err := genconfig.LoadParamConfig("/nonexistent/params.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadParamConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "params.yaml")
	os.WriteFile(path, []byte("invalid: yaml: :\n"), 0644)

	_, err := genconfig.LoadParamConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoadPromptConfig_Full(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.yaml")
	content := `
prompt: "beautiful landscape, sunset"
negative_prompt: "ugly, distorted"
`
	os.WriteFile(path, []byte(content), 0644)

	cfg, err := genconfig.LoadPromptConfig(path)
	if err != nil {
		t.Fatalf("LoadPromptConfig failed: %v", err)
	}

	if cfg.Prompt != "beautiful landscape, sunset" {
		t.Errorf("unexpected Prompt: %s", cfg.Prompt)
	}
	if cfg.NegativePrompt != "ugly, distorted" {
		t.Errorf("unexpected NegativePrompt: %s", cfg.NegativePrompt)
	}
}

func TestLoadPromptConfig_PromptOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.yaml")
	os.WriteFile(path, []byte("prompt: \"cats playing piano\"\n"), 0644)

	cfg, err := genconfig.LoadPromptConfig(path)
	if err != nil {
		t.Fatalf("LoadPromptConfig failed: %v", err)
	}

	if cfg.Prompt != "cats playing piano" {
		t.Errorf("unexpected Prompt: %s", cfg.Prompt)
	}
	if cfg.NegativePrompt != "" {
		t.Errorf("NegativePrompt should be empty when not specified, got: %s", cfg.NegativePrompt)
	}
}

func TestLoadPromptConfig_MissingFile(t *testing.T) {
	_, err := genconfig.LoadPromptConfig("/nonexistent/prompt.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadPromptConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.yaml")
	os.WriteFile(path, []byte("invalid: yaml: :\n"), 0644)

	_, err := genconfig.LoadPromptConfig(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}
