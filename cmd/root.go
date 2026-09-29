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
	Long: `CLI for Stable Diffusion WebUI (AUTOMATIC1111)

Environment variables (override the same keys in the config file):
  SDCTL_URL           WebUI URL (config: url)
  SDCTL_PARAMS        default --params file for txt2img/img2img/hires (config: params)
  SDCTL_OUTPUT_DIR    default -o directory for txt2img/img2img/hires, created if missing (config: output_dir)
  SDCTL_FORMAT        default image format, png or jpeg, used when neither --format nor
                      the -o file extension decides it; built-in default png (config: format)
  SDCTL_JPEG_QUALITY  default --quality for JPEG output, 1-100; built-in default 90 (config: jpeg_quality)

Set any of these (except SDCTL_URL) to an empty string to ignore the config file value.
JPEG output is re-encoded from the PNG returned by the WebUI, so it does not keep the
generation parameters (infotext) embedded in the PNG.
Command-line flags always take precedence. Saved image paths are printed to stdout,
one per line; the progress bar goes to stderr and only when stderr is a terminal.`,
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
