package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	URL string `yaml:"url"`
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

	return cfg, nil
}
