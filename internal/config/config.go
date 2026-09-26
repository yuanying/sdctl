package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	URL string `yaml:"url"`
	// Params is the default generation parameter file used when --params is not given.
	Params string `yaml:"params"`
	// OutputDir is the default output directory used when -o is not given.
	OutputDir string `yaml:"output_dir"`
}

func Default() *Config {
	return &Config{
		URL: "http://localhost:7860",
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, err
		}
	}

	if url := os.Getenv("SDCTL_URL"); url != "" {
		cfg.URL = url
	}
	// An environment variable set to an empty string disables the config file default.
	if params, ok := os.LookupEnv("SDCTL_PARAMS"); ok {
		cfg.Params = params
	}
	if outputDir, ok := os.LookupEnv("SDCTL_OUTPUT_DIR"); ok {
		cfg.OutputDir = outputDir
	}

	return cfg, nil
}
