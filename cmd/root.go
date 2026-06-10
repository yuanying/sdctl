package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yuanying/sdctl/internal/api"
	"github.com/yuanying/sdctl/internal/config"
)

var (
	cfgFile string
	cfg     *config.Config
	client  *api.Client
)

var rootCmd = &cobra.Command{
	Use:   "sdctl",
	Short: "CLI for Stable Diffusion WebUI (AUTOMATIC1111)",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("config: %w", err)
		}
		client = api.NewClient(cfg.URL)
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	defaultConfig := filepath.Join(os.Getenv("HOME"), ".config", "sdctl", "config.yaml")
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", defaultConfig, "config file path")
}
