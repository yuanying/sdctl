package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"github.com/yuanying/sdctl/internal/api"
	"github.com/yuanying/sdctl/internal/genconfig"
)

func resolvePrompt(args []string, promptCfg *genconfig.PromptConfig) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	if promptCfg != nil && promptCfg.Prompt != "" {
		return promptCfg.Prompt, nil
	}
	return "", fmt.Errorf("prompt is required: provide as argument or via --prompt file")
}

func resolveNegativePrompt(cmd *cobra.Command, flagVal string, promptCfg *genconfig.PromptConfig, paramCfg *genconfig.ParamConfig) string {
	if cmd.Flags().Changed("negative") {
		return flagVal
	}
	if promptCfg != nil && promptCfg.NegativePrompt != "" {
		return promptCfg.NegativePrompt
	}
	if v := paramCfg.NegativePromptValue(); v != nil {
		return *v
	}
	return flagVal
}

func resolveInt(cmd *cobra.Command, flagName string, flagVal int, cfgVal *int) int {
	if cmd.Flags().Changed(flagName) || cfgVal == nil {
		return flagVal
	}
	return *cfgVal
}

func resolveInt64(cmd *cobra.Command, flagName string, flagVal int64, cfgVal *int64) int64 {
	if cmd.Flags().Changed(flagName) || cfgVal == nil {
		return flagVal
	}
	return *cfgVal
}

func resolveFloat64(cmd *cobra.Command, flagName string, flagVal float64, cfgVal *float64) float64 {
	if cmd.Flags().Changed(flagName) || cfgVal == nil {
		return flagVal
	}
	return *cfgVal
}

func resolveString(cmd *cobra.Command, flagName string, flagVal string, cfgVal *string) string {
	if cmd.Flags().Changed(flagName) || cfgVal == nil {
		return flagVal
	}
	return *cfgVal
}

func validateSampler(name string) error {
	samplers, err := client.ListSamplers()
	if err != nil {
		return fmt.Errorf("error fetching samplers: %w", err)
	}
	return api.ValidateSamplerName(samplers, name)
}

func validateScheduler(name string) error {
	schedulers, err := client.ListSchedulers()
	if err != nil {
		return fmt.Errorf("error fetching schedulers: %w", err)
	}
	return api.ValidateSchedulerName(schedulers, name)
}

func saveImages(images []string, outputPath string) ([]string, error) {
	if len(images) == 1 {
		path, err := saveImage(images[0], outputPath)
		if err != nil {
			return nil, err
		}
		return []string{path}, nil
	}

	// Multiple images: directory or default → timestamp-based names
	if outputPath == "" {
		return saveImagesToDir(images, "")
	}
	info, err := os.Stat(outputPath)
	if err == nil && info.IsDir() {
		return saveImagesToDir(images, outputPath)
	}

	// File path specified: generate <stem>.<N><ext> with zero-padding
	ext := filepath.Ext(outputPath)
	stem := strings.TrimSuffix(outputPath, ext)
	width := len(strconv.Itoa(len(images) - 1))

	paths := make([]string, 0, len(images))
	for i, imgData := range images {
		dest := fmt.Sprintf("%s.%0*d%s", stem, width, i, ext)
		data, err := base64.StdEncoding.DecodeString(imgData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image %d: %w", i+1, err)
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write image %d: %w", i+1, err)
		}
		paths = append(paths, dest)
	}
	return paths, nil
}

func saveImagesToDir(images []string, dir string) ([]string, error) {
	base := time.Now().Format("20060102-150405")
	paths := make([]string, 0, len(images))
	for i, imgData := range images {
		filename := fmt.Sprintf("output-%s-%d.png", base, i+1)
		dest := filename
		if dir != "" {
			dest = filepath.Join(dir, filename)
		}
		data, err := base64.StdEncoding.DecodeString(imgData)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image %d: %w", i+1, err)
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to write image %d: %w", i+1, err)
		}
		paths = append(paths, dest)
	}
	return paths, nil
}

func saveImage(b64data, outputPath string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(b64data)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	dest := resolveOutputPath(outputPath)
	if err := os.WriteFile(dest, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write image: %w", err)
	}
	return dest, nil
}

func resolveOutputPath(outputPath string) string {
	if outputPath != "" {
		info, err := os.Stat(outputPath)
		if err == nil && info.IsDir() {
			return filepath.Join(outputPath, defaultFilename())
		}
		return outputPath
	}
	return defaultFilename()
}

func defaultFilename() string {
	return fmt.Sprintf("output-%s.png", time.Now().Format("20060102-150405"))
}

func watchProgress(stop <-chan struct{}) {
	bar := progressbar.NewOptions(100,
		progressbar.OptionSetDescription("Generating"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowBytes(false),
		progressbar.OptionClearOnFinish(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			bar.Set(100)
			bar.Finish()
			return
		case <-ticker.C:
			resp, err := client.GetProgress()
			if err != nil {
				continue
			}
			updateProgress(bar, resp)
		}
	}
}

func updateProgress(bar *progressbar.ProgressBar, resp *api.ProgressResponse) {
	pct := int(resp.Progress * 100)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	bar.Set(pct)
	if resp.State.SamplingSteps > 0 {
		bar.Describe(fmt.Sprintf("Generating [%d/%d steps]",
			resp.State.SamplingStep, resp.State.SamplingSteps))
	}
}
