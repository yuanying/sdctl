package cmd

import (
	"encoding/base64"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
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

// resolveModulePath returns the full path for a module name.
// If value is already an absolute path or empty, it is returned as-is.
func resolveModulePath(value string, modules []api.SDModule) string {
	if value == "" || strings.HasPrefix(value, "/") {
		return value
	}
	for _, m := range modules {
		if m.ModelName == value {
			return m.Filename
		}
	}
	return value
}

// resolveOverrideModules resolves model names in forge_additional_modules
// to full paths using the provided modules list.
func resolveOverrideModules(settings map[string]any, modules []api.SDModule) map[string]any {
	if settings == nil {
		return nil
	}
	raw, ok := settings["forge_additional_modules"]
	if !ok {
		return settings
	}

	var names []string
	switch v := raw.(type) {
	case []string:
		names = v
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				names = append(names, s)
			}
		}
	}

	resolved := make([]string, len(names))
	for i, n := range names {
		resolved[i] = resolveModulePath(n, modules)
	}

	result := make(map[string]any, len(settings))
	maps.Copy(result, settings)
	result["forge_additional_modules"] = resolved
	return result
}

func buildModelOverride(model string) map[string]any {
	if model == "" {
		return nil
	}
	return map[string]any{
		"sd_model_checkpoint": model,
	}
}

func buildAdditionalModules(vae, textEncoder string) map[string]any {
	var modules []string
	if vae != "" {
		modules = append(modules, vae)
	}
	if textEncoder != "" {
		modules = append(modules, textEncoder)
	}
	if len(modules) == 0 {
		return nil
	}
	return map[string]any{
		"forge_additional_modules": modules,
	}
}

// resolveFlag returns the flag value only when the flag was explicitly set on the command line.
func resolveFlag(cmd *cobra.Command, flagName, flagVal string) string {
	if cmd.Flags().Changed(flagName) {
		return flagVal
	}
	return ""
}

func boolPtrIfSet(m map[string]any) *bool {
	if m == nil {
		return nil
	}
	f := false
	return &f
}

func mergeMap(base, override map[string]any) map[string]any {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}
	result := make(map[string]any, len(base)+len(override))
	maps.Copy(result, base)
	maps.Copy(result, override)
	return result
}

func validateModel(name string) error {
	models, err := client.ListModels()
	if err != nil {
		return fmt.Errorf("error fetching models: %w", err)
	}
	return api.ValidateModelName(models, name)
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
	isBatch := len(images) > 1

	if outputPath != "" {
		info, err := os.Stat(outputPath)
		if err == nil && info.IsDir() {
			return saveImagesToDir(images, outputPath)
		}
		paths := make([]string, 0, len(images))
		for i, imgData := range images {
			dest := resolveUniqueFilePath(outputPath, isBatch)
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

	return saveImagesToDir(images, "")
}

// resolveUniqueFilePath returns a path that does not conflict with existing files.
// If isBatch is true, always appends an indexed suffix (output.0001.png).
// If isBatch is false, the original path is returned as-is when the file does not exist.
func resolveUniqueFilePath(outputPath string, isBatch bool) string {
	if !isBatch {
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			return outputPath
		}
	}

	ext := filepath.Ext(outputPath)
	stem := strings.TrimSuffix(outputPath, ext)
	dir := filepath.Dir(outputPath)
	base := filepath.Base(stem)

	next := findMaxIndexedSuffix(dir, base, ext) + 1
	digits := max(4, len(strconv.Itoa(next)))
	return fmt.Sprintf("%s.%0*d%s", stem, digits, next, ext)
}

// findMaxIndexedSuffix scans dir for files matching <base>.<digits><ext> and returns the max index found.
func findMaxIndexedSuffix(dir, base, ext string) int {
	pattern := regexp.MustCompile(`^` + regexp.QuoteMeta(base) + `\.(\d+)` + regexp.QuoteMeta(ext) + `$`)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	max := 0
	for _, e := range entries {
		m := pattern.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max
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
	pct := max(0, min(100, int(resp.Progress*100)))
	bar.Set(pct)
	if resp.State.SamplingSteps > 0 {
		bar.Describe(fmt.Sprintf("Generating [%d/%d steps]",
			resp.State.SamplingStep, resp.State.SamplingSteps))
	}
}
