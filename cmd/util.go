package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/yuanying/sdctl/internal/api"
)

func validateOutputForBatch(outputPath string, batchCount, batchSize int) error {
	if batchCount*batchSize <= 1 {
		return nil
	}
	if outputPath == "" {
		return nil
	}
	info, err := os.Stat(outputPath)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("--output must be a directory when generating multiple images")
	}
	return nil
}

func saveImages(images []string, outputPath string) ([]string, error) {
	if len(images) == 1 {
		path, err := saveImage(images[0], outputPath)
		if err != nil {
			return nil, err
		}
		return []string{path}, nil
	}

	// Multiple images: output must be a directory or empty (current dir)
	if outputPath != "" {
		info, err := os.Stat(outputPath)
		if err != nil || !info.IsDir() {
			return nil, fmt.Errorf("--output must be a directory when generating multiple images")
		}
	}

	base := time.Now().Format("20060102-150405")
	paths := make([]string, 0, len(images))
	for i, imgData := range images {
		filename := fmt.Sprintf("output-%s-%d.png", base, i+1)
		var dest string
		if outputPath != "" {
			dest = filepath.Join(outputPath, filename)
		} else {
			dest = filename
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
