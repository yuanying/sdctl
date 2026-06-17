package genconfig

import (
	"os"

	"gopkg.in/yaml.v3"
)

type ParamConfig struct {
	NegativePrompt    *string        `yaml:"negative_prompt"`
	Steps             *int           `yaml:"steps"`
	Width             *int           `yaml:"width"`
	Height            *int           `yaml:"height"`
	CFGScale          *float64       `yaml:"cfg_scale"`
	Sampler           *string        `yaml:"sampler"`
	Scheduler         *string        `yaml:"scheduler"`
	Seed              *int64         `yaml:"seed"`
	BatchCount        *int           `yaml:"batch_count"`
	BatchSize         *int           `yaml:"batch_size"`
	DenoisingStrength *float64       `yaml:"denoising_strength"`
	OverrideSettings  map[string]any `yaml:"override_settings"`
	EnableHR          *bool          `yaml:"enable_hr"`
	HRScale           *float64       `yaml:"hr_scale"`
	HRUpscaler        *string        `yaml:"hr_upscaler"`
	HRSecondPassSteps *int           `yaml:"hr_second_pass_steps"`
	HRDenoise         *float64       `yaml:"hr_denoise"`
}

type PromptConfig struct {
	Prompt         string `yaml:"prompt"`
	NegativePrompt string `yaml:"negative_prompt"`
}

func LoadParamConfig(path string) (*ParamConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ParamConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *ParamConfig) StepsValue() *int {
	if c == nil {
		return nil
	}
	return c.Steps
}

func (c *ParamConfig) WidthValue() *int {
	if c == nil {
		return nil
	}
	return c.Width
}

func (c *ParamConfig) HeightValue() *int {
	if c == nil {
		return nil
	}
	return c.Height
}

func (c *ParamConfig) CFGScaleValue() *float64 {
	if c == nil {
		return nil
	}
	return c.CFGScale
}

func (c *ParamConfig) SamplerValue() *string {
	if c == nil {
		return nil
	}
	return c.Sampler
}

func (c *ParamConfig) SchedulerValue() *string {
	if c == nil {
		return nil
	}
	return c.Scheduler
}

func (c *ParamConfig) SeedValue() *int64 {
	if c == nil {
		return nil
	}
	return c.Seed
}

func (c *ParamConfig) BatchCountValue() *int {
	if c == nil {
		return nil
	}
	return c.BatchCount
}

func (c *ParamConfig) BatchSizeValue() *int {
	if c == nil {
		return nil
	}
	return c.BatchSize
}

func (c *ParamConfig) DenoisingStrengthValue() *float64 {
	if c == nil {
		return nil
	}
	return c.DenoisingStrength
}

func (c *ParamConfig) NegativePromptValue() *string {
	if c == nil {
		return nil
	}
	return c.NegativePrompt
}

func (c *ParamConfig) OverrideSettingsValue() map[string]any {
	if c == nil {
		return nil
	}
	return c.OverrideSettings
}

func (c *ParamConfig) EnableHRValue() *bool {
	if c == nil {
		return nil
	}
	return c.EnableHR
}

func (c *ParamConfig) HRScaleValue() *float64 {
	if c == nil {
		return nil
	}
	return c.HRScale
}

func (c *ParamConfig) HRUpscalerValue() *string {
	if c == nil {
		return nil
	}
	return c.HRUpscaler
}

func (c *ParamConfig) HRSecondPassStepsValue() *int {
	if c == nil {
		return nil
	}
	return c.HRSecondPassSteps
}

func (c *ParamConfig) HRDenoiseValue() *float64 {
	if c == nil {
		return nil
	}
	return c.HRDenoise
}

func LoadPromptConfig(path string) (*PromptConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg PromptConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
