package cmd

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func newFormatCmd(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	var s string
	var q int
	cmd.Flags().StringVar(&s, "format", "", "")
	cmd.Flags().IntVar(&q, "quality", 0, "")
	if err := cmd.Flags().Parse(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	return cmd
}

// pngBase64 returns a base64 PNG like the WebUI API returns.
// Noise makes the JPEG size depend on the quality.
func pngBase64(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	r := rand.New(rand.NewSource(1))
	for y := 0; y < 64; y++ {
		for x := 0; x < 64; x++ {
			img.Set(x, y, color.RGBA{uint8(r.Intn(256)), uint8(r.Intn(256)), uint8(r.Intn(256)), 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func assertJPEG(t *testing.T, path string) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	img, err := jpeg.Decode(f)
	if err != nil {
		t.Fatalf("%s is not a JPEG: %v", path, err)
	}
	if img.Bounds().Dx() != 64 || img.Bounds().Dy() != 64 {
		t.Errorf("%s: unexpected size %v", path, img.Bounds())
	}
}

func TestParseImageFormat(t *testing.T) {
	for in, want := range map[string]imageFormat{
		"png": formatPNG, "PNG": formatPNG,
		"jpeg": formatJPEG, "jpg": formatJPEG, "JPG": formatJPEG, "Jpeg": formatJPEG,
	} {
		got, err := parseImageFormat(in)
		if err != nil {
			t.Errorf("%q: unexpected error %v", in, err)
		}
		if got != want {
			t.Errorf("%q: expected %s, got %s", in, want, got)
		}
	}
}

func TestParseImageFormat_UnknownIsError(t *testing.T) {
	_, err := parseImageFormat("webp")
	if err == nil {
		t.Fatal("expected error for webp")
	}
	for _, s := range []string{`"webp"`, "png", "jpeg"} {
		if !strings.Contains(err.Error(), s) {
			t.Errorf("error should mention %s: %v", s, err)
		}
	}
}

func TestImageFormatFromPath(t *testing.T) {
	cases := map[string]struct {
		want imageFormat
		ok   bool
	}{
		"out.jpg":      {formatJPEG, true},
		"out.JPG":      {formatJPEG, true},
		"out.jpeg":     {formatJPEG, true},
		"dir/out.JPEG": {formatJPEG, true},
		"out.png":      {formatPNG, true},
		"out.PNG":      {formatPNG, true},
		"out":          {"", false},
		"out.webp":     {"", false},
	}
	for path, tc := range cases {
		got, ok := imageFormatFromPath(path)
		if got != tc.want || ok != tc.ok {
			t.Errorf("%q: expected (%q, %v), got (%q, %v)", path, tc.want, tc.ok, got, ok)
		}
	}
}

func TestResolveImageFormat_Precedence(t *testing.T) {
	dir := t.TempDir()
	cases := []struct {
		name   string
		args   []string
		output string
		cfg    string
		want   imageFormat
	}{
		{"nothing set keeps PNG", nil, "", "", formatPNG},
		{"config/env value", nil, "", "jpeg", formatJPEG},
		{"output extension beats config", nil, "out.png", "jpeg", formatPNG},
		{"upper case extension", nil, "out.JPG", "", formatJPEG},
		{"directory output ignores extension", nil, dir, "jpg", formatJPEG},
		{"unknown extension falls through", nil, "out.webp", "jpeg", formatJPEG},
		{"flag beats config", []string{"--format", "png"}, "", "jpeg", formatPNG},
		{"flag with matching extension", []string{"--format", "jpg"}, "out.jpeg", "png", formatJPEG},
		{"flag with unknown extension", []string{"--format", "jpeg"}, "out.webp", "", formatJPEG},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := newFormatCmd(t, tc.args...)
			flagVal, _ := cmd.Flags().GetString("format")
			got, err := resolveImageFormat(cmd, flagVal, tc.output, tc.cfg)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("expected %s, got %s", tc.want, got)
			}
		})
	}
}

func TestResolveImageFormat_FlagConflictsWithExtension(t *testing.T) {
	cmd := newFormatCmd(t, "--format", "jpeg")
	_, err := resolveImageFormat(cmd, "jpeg", "out.png", "")
	if err == nil || !strings.Contains(err.Error(), "out.png") {
		t.Errorf("expected conflict error naming out.png, got %v", err)
	}
}

func TestResolveImageFormat_UnknownValuesAreErrors(t *testing.T) {
	cmd := newFormatCmd(t, "--format", "gif")
	if _, err := resolveImageFormat(cmd, "gif", "", ""); err == nil || !strings.Contains(err.Error(), "--format") {
		t.Errorf("expected --format error, got %v", err)
	}
	cmd = newFormatCmd(t)
	if _, err := resolveImageFormat(cmd, "", "", "bmp"); err == nil || !strings.Contains(err.Error(), "SDCTL_FORMAT") {
		t.Errorf("expected SDCTL_FORMAT/config error, got %v", err)
	}
}

func TestResolveJPEGQuality(t *testing.T) {
	if got, _ := resolveJPEGQuality(newFormatCmd(t), 0, 0); got != defaultJPEGQuality {
		t.Errorf("expected default %d, got %d", defaultJPEGQuality, got)
	}
	if got, _ := resolveJPEGQuality(newFormatCmd(t), 0, 75); got != 75 {
		t.Errorf("expected config 75, got %d", got)
	}
	if got, _ := resolveJPEGQuality(newFormatCmd(t, "--quality", "60"), 60, 75); got != 60 {
		t.Errorf("expected flag 60, got %d", got)
	}
	for _, q := range []int{0, 101, -1} {
		if _, err := resolveJPEGQuality(newFormatCmd(t, "--quality", "0"), q, 0); err == nil {
			t.Errorf("expected error for --quality %d", q)
		}
	}
	if _, err := resolveJPEGQuality(newFormatCmd(t), 0, 101); err == nil {
		t.Error("expected error for config quality 101")
	}
}

func TestSaveImages_PNGKeepsAPIBytes(t *testing.T) {
	dir := t.TempDir()
	src := pngBase64(t)
	paths, err := saveImages([]string{src}, dir, saveOptions{format: formatPNG, quality: 90})
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	if !strings.HasSuffix(paths[0], ".png") {
		t.Errorf("expected .png name, got %s", paths[0])
	}
	want, _ := base64.StdEncoding.DecodeString(src)
	got, _ := os.ReadFile(paths[0])
	if !bytes.Equal(got, want) {
		t.Error("PNG output should be the API bytes unchanged")
	}
}

func TestSaveImages_JPEGAutoNames(t *testing.T) {
	dir := t.TempDir()
	fixed := time.Date(2026, 9, 29, 1, 2, 3, 0, time.Local)
	orig := now
	now = func() time.Time { return fixed }
	defer func() { now = orig }()
	// An existing file must not be overwritten.
	os.WriteFile(filepath.Join(dir, "output-20260929-010203-1.jpg"), []byte("keep"), 0644)

	src := pngBase64(t)
	paths, err := saveImages([]string{src, src}, dir, saveOptions{format: formatJPEG, quality: 90})
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	want := []string{
		filepath.Join(dir, "output-20260929-010203-2.jpg"),
		filepath.Join(dir, "output-20260929-010203-3.jpg"),
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("path[%d]: expected %s, got %s", i, want[i], paths[i])
		}
		assertJPEG(t, paths[i])
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "output-20260929-010203-1.jpg")); string(data) != "keep" {
		t.Error("existing file was overwritten")
	}
}

func TestSaveImages_JPEGBatchToFile(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "result.JPG")
	src := pngBase64(t)
	paths, err := saveImages([]string{src, src}, out, saveOptions{format: formatJPEG, quality: 90})
	if err != nil {
		t.Fatalf("saveImages failed: %v", err)
	}
	want := []string{filepath.Join(dir, "result.0001.JPG"), filepath.Join(dir, "result.0002.JPG")}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("path[%d]: expected %s, got %s", i, want[i], paths[i])
		}
		assertJPEG(t, paths[i])
	}
}

func TestSaveImages_JPEGQualityChangesSize(t *testing.T) {
	src := pngBase64(t)
	size := func(q int) int64 {
		dir := t.TempDir()
		paths, err := saveImages([]string{src}, dir, saveOptions{format: formatJPEG, quality: q})
		if err != nil {
			t.Fatalf("saveImages failed: %v", err)
		}
		info, err := os.Stat(paths[0])
		if err != nil {
			t.Fatalf("stat: %v", err)
		}
		return info.Size()
	}
	if low, high := size(10), size(100); low >= high {
		t.Errorf("quality 10 (%d bytes) should be smaller than quality 100 (%d bytes)", low, high)
	}
}

func TestSaveImages_JPEGRejectsNonImage(t *testing.T) {
	_, err := saveImages([]string{"aGVsbG8="}, t.TempDir(), saveOptions{format: formatJPEG, quality: 90})
	if err == nil {
		t.Error("expected error when the API data is not an image")
	}
}

// --- end-to-end ---

func TestCLI_JPEGFromEnv(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			srv.image = pngBase64(t)
			outDir := filepath.Join(t.TempDir(), "images")
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", "")
			t.Setenv("SDCTL_OUTPUT_DIR", outDir)
			t.Setenv("SDCTL_FORMAT", "jpeg")
			t.Setenv("SDCTL_JPEG_QUALITY", "85")

			stdout := runCLI(t, tc.args(t)...)

			paths := assertStdoutIsPathsIn(t, stdout, outDir, 1)
			if !strings.HasSuffix(paths[0], ".jpg") {
				t.Errorf("expected .jpg name, got %s", paths[0])
			}
			assertJPEG(t, paths[0])
		})
	}
}

func TestCLI_FormatFlagAndOutputExtension(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			srv.image = pngBase64(t)
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", "")
			t.Setenv("SDCTL_OUTPUT_DIR", "")
			t.Setenv("SDCTL_FORMAT", "png")

			out := filepath.Join(t.TempDir(), "flag.jpeg")
			stdout := runCLI(t, append(tc.args(t), "--format", "jpg", "--quality", "50", "-o", out)...)
			if strings.TrimRight(stdout, "\n") != out {
				t.Errorf("expected stdout %q, got %q", out, stdout)
			}
			assertJPEG(t, out)

			byExt := filepath.Join(t.TempDir(), "ext.JPG")
			runCLI(t, append(tc.args(t), "-o", byExt)...)
			assertJPEG(t, byExt)
		})
	}
}

func TestCLI_FormatConflictIsError(t *testing.T) {
	srv := newFakeServer(t)
	t.Setenv("SDCTL_URL", srv.URL)
	t.Setenv("SDCTL_PARAMS", "")

	out := filepath.Join(t.TempDir(), "out.png")
	_, err := runCLIErr(t, "txt2img", "a cat", "--format", "jpeg", "-o", out)
	if err == nil {
		t.Fatal("expected error for --format jpeg with -o out.png")
	}
	if srv.body("/sdapi/v1/txt2img") != nil {
		t.Error("the conflict should be reported before generating")
	}
}
