package cmd

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // register the PNG decoder for image.Decode
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

type imageFormat string

const (
	formatPNG  imageFormat = "png"
	formatJPEG imageFormat = "jpeg"

	defaultJPEGQuality = 90
)

// ext returns the file extension used for auto-generated file names.
func (f imageFormat) ext() string {
	if f == formatJPEG {
		return ".jpg"
	}
	return ".png"
}

// saveOptions controls how images returned by the API are written.
type saveOptions struct {
	format  imageFormat
	quality int
}

// parseImageFormat parses png, jpeg or jpg (case-insensitive).
func parseImageFormat(s string) (imageFormat, error) {
	switch strings.ToLower(s) {
	case "png":
		return formatPNG, nil
	case "jpeg", "jpg":
		return formatJPEG, nil
	}
	return "", fmt.Errorf("unknown image format %q: use png or jpeg (jpg)", s)
}

// imageFormatFromPath returns the format implied by the file extension of path.
// ok is false when the extension is not .png, .jpg or .jpeg.
func imageFormatFromPath(path string) (imageFormat, bool) {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".png":
		return formatPNG, true
	case ".jpg", ".jpeg":
		return formatJPEG, true
	}
	return "", false
}

// resolveImageFormat decides the output format.
// Precedence: --format flag, -o file extension, SDCTL_FORMAT / config format, PNG.
// A --format that contradicts the -o file extension is an error.
func resolveImageFormat(cmd *cobra.Command, flagVal, output, cfgVal string) (imageFormat, error) {
	var extFormat imageFormat
	hasExt := false
	if output != "" {
		if info, err := os.Stat(output); err != nil || !info.IsDir() {
			extFormat, hasExt = imageFormatFromPath(output)
		}
	}

	if cmd.Flags().Changed("format") {
		f, err := parseImageFormat(flagVal)
		if err != nil {
			return "", fmt.Errorf("--format: %w", err)
		}
		if hasExt && extFormat != f {
			return "", fmt.Errorf("--format %s conflicts with the extension of -o %s", flagVal, output)
		}
		return f, nil
	}
	if hasExt {
		return extFormat, nil
	}
	if cfgVal != "" {
		f, err := parseImageFormat(cfgVal)
		if err != nil {
			return "", fmt.Errorf("SDCTL_FORMAT / config format: %w", err)
		}
		return f, nil
	}
	return formatPNG, nil
}

// resolveJPEGQuality returns the JPEG quality from --quality, SDCTL_JPEG_QUALITY / config
// jpeg_quality (0 means unset), or the default, and checks that it is within 1-100.
func resolveJPEGQuality(cmd *cobra.Command, flagVal, cfgVal int) (int, error) {
	if cmd.Flags().Changed("quality") {
		if flagVal < 1 || flagVal > 100 {
			return 0, fmt.Errorf("--quality must be between 1 and 100, got %d", flagVal)
		}
		return flagVal, nil
	}
	if cfgVal == 0 {
		return defaultJPEGQuality, nil
	}
	if cfgVal < 1 || cfgVal > 100 {
		return 0, fmt.Errorf("SDCTL_JPEG_QUALITY / config jpeg_quality must be between 1 and 100, got %d", cfgVal)
	}
	return cfgVal, nil
}

// resolveSaveOptions resolves the image format and JPEG quality for a generation command.
func resolveSaveOptions(cmd *cobra.Command, formatFlag string, qualityFlag int, output string) (saveOptions, error) {
	format, err := resolveImageFormat(cmd, formatFlag, output, cfg.Format)
	if err != nil {
		return saveOptions{}, err
	}
	quality, err := resolveJPEGQuality(cmd, qualityFlag, cfg.JPEGQuality)
	if err != nil {
		return saveOptions{}, err
	}
	return saveOptions{format: format, quality: quality}, nil
}

// encodeImage converts the image bytes returned by the API into the requested format.
// PNG data is returned unchanged so the embedded generation parameters are kept.
func encodeImage(data []byte, opts saveOptions) ([]byte, error) {
	if opts.format != formatJPEG {
		return data, nil
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image returned by the API: %w", err)
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.quality}); err != nil {
		return nil, fmt.Errorf("failed to encode JPEG: %w", err)
	}
	return buf.Bytes(), nil
}
