package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type Config struct {
	URL string `yaml:"url"`
	// Params is the default generation parameter file used when --params is not given.
	Params string `yaml:"params"`
	// OutputDir is the default output directory used when -o is not given.
	OutputDir string `yaml:"output_dir"`
	// Format is the default image format (png or jpeg) used when neither --format
	// nor the -o file extension decides it.
	Format string `yaml:"format"`
	// JPEGQuality is the default JPEG quality (1-100) used when --quality is not given.
	// 0 means unset.
	JPEGQuality int `yaml:"jpeg_quality"`
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
	if format, ok := os.LookupEnv("SDCTL_FORMAT"); ok {
		cfg.Format = format
	}
	if quality, ok := os.LookupEnv("SDCTL_JPEG_QUALITY"); ok {
		cfg.JPEGQuality = 0
		if quality != "" {
			q, err := strconv.Atoi(quality)
			if err != nil {
				return nil, fmt.Errorf("SDCTL_JPEG_QUALITY must be an integer: %q", quality)
			}
			cfg.JPEGQuality = q
		}
	}

	return cfg, nil
}
