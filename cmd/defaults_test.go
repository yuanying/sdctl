package cmd

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/yuanying/sdctl/internal/api"
)

func newFlagCmd(t *testing.T, args ...string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "test"}
	var s string
	cmd.Flags().StringVarP(&s, "output", "o", "", "")
	cmd.Flags().StringVar(&s, "params", "", "")
	if err := cmd.Flags().Parse(args); err != nil {
		t.Fatalf("parse flags: %v", err)
	}
	return cmd
}

func TestResolveDefaultString_FlagNotSet_UsesDefault(t *testing.T) {
	cmd := newFlagCmd(t)
	got := resolveDefaultString(cmd, "params", "", "/etc/sdctl/default.yaml")
	if got != "/etc/sdctl/default.yaml" {
		t.Errorf("expected default, got %q", got)
	}
}

func TestResolveDefaultString_FlagSet_Wins(t *testing.T) {
	cmd := newFlagCmd(t, "--params", "mine.yaml")
	got := resolveDefaultString(cmd, "params", "mine.yaml", "/etc/sdctl/default.yaml")
	if got != "mine.yaml" {
		t.Errorf("expected flag value, got %q", got)
	}
}

func TestResolveDefaultString_EmptyFlag_DisablesDefault(t *testing.T) {
	cmd := newFlagCmd(t, "--params", "")
	got := resolveDefaultString(cmd, "params", "", "/etc/sdctl/default.yaml")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestResolveOutput_FlagNotSet_NoDefault(t *testing.T) {
	cmd := newFlagCmd(t)
	got, err := resolveOutput(cmd, "", "")
	if err != nil {
		t.Fatalf("resolveOutput failed: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty output, got %q", got)
	}
}

func TestResolveOutput_DefaultDirIsCreated(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "out")
	cmd := newFlagCmd(t)

	got, err := resolveOutput(cmd, "", dir)
	if err != nil {
		t.Fatalf("resolveOutput failed: %v", err)
	}
	if got != dir {
		t.Errorf("expected %s, got %s", dir, got)
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Errorf("expected directory to be created: %s", dir)
	}
}

func TestResolveOutput_FlagWinsOverDefault(t *testing.T) {
	defaultDir := filepath.Join(t.TempDir(), "default")
	cmd := newFlagCmd(t, "-o", "mine.png")

	got, err := resolveOutput(cmd, "mine.png", defaultDir)
	if err != nil {
		t.Fatalf("resolveOutput failed: %v", err)
	}
	if got != "mine.png" {
		t.Errorf("expected flag value, got %q", got)
	}
	if _, err := os.Stat(defaultDir); !os.IsNotExist(err) {
		t.Errorf("default dir should not be created when -o is given")
	}
}

func TestSaveImagesToDir_SameSecondDoesNotOverwrite(t *testing.T) {
	dir := t.TempDir()
	fixed := time.Date(2026, 9, 26, 12, 34, 56, 0, time.Local)
	orig := now
	now = func() time.Time { return fixed }
	defer func() { now = orig }()

	opts := saveOptions{format: formatPNG}
	first, err := saveImages([]string{"Zmlyc3Q="}, dir, opts) // "first"
	if err != nil {
		t.Fatalf("first saveImages failed: %v", err)
	}
	second, err := saveImages([]string{"c2Vjb25k", "dGhpcmQ="}, dir, opts) // "second", "third"
	if err != nil {
		t.Fatalf("second saveImages failed: %v", err)
	}

	want := []string{
		filepath.Join(dir, "output-20260926-123456-1.png"),
		filepath.Join(dir, "output-20260926-123456-2.png"),
		filepath.Join(dir, "output-20260926-123456-3.png"),
	}
	got := append(first, second...)
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("path[%d]: expected %s, got %s", i, want[i], got[i])
		}
	}
	for path, content := range map[string]string{want[0]: "first", want[1]: "second", want[2]: "third"} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		if string(data) != content {
			t.Errorf("%s: expected %q, got %q", path, content, data)
		}
	}
}

func TestStartProgress_DisabledWritesNothing(t *testing.T) {
	srv := newFakeServer(t)
	client = api.NewClient(srv.URL)

	var buf bytes.Buffer
	stop := startProgress(&buf, false)
	time.Sleep(700 * time.Millisecond)
	stop()

	if buf.Len() != 0 {
		t.Errorf("expected no progress output, got %q", buf.String())
	}
}

func TestStartProgress_EnabledWritesToGivenWriter(t *testing.T) {
	srv := newFakeServer(t)
	client = api.NewClient(srv.URL)

	var buf syncBuffer
	stop := startProgress(&buf, true)
	time.Sleep(700 * time.Millisecond)
	stop()

	if !strings.Contains(buf.String(), "Generating") {
		t.Errorf("expected progress bar on writer, got %q", buf.String())
	}
}

// --- end-to-end tests against a fake WebUI API ---

type fakeServer struct {
	*httptest.Server
	mu     sync.Mutex
	bodies map[string]map[string]any
	// image is the base64 image returned by txt2img/img2img.
	image string
}

func newFakeServer(t *testing.T) *fakeServer {
	t.Helper()
	fs := &fakeServer{bodies: map[string]map[string]any{}, image: "aGVsbG8="}
	fs.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sdapi/v1/progress":
			json.NewEncoder(w).Encode(map[string]any{"progress": 0.5})
		case "/sdapi/v1/txt2img", "/sdapi/v1/img2img":
			// Take long enough for the progress watcher to tick at least once.
			time.Sleep(600 * time.Millisecond)
			var body map[string]any
			json.NewDecoder(r.Body).Decode(&body)
			fs.mu.Lock()
			fs.bodies[r.URL.Path] = body
			img := fs.image
			fs.mu.Unlock()
			json.NewEncoder(w).Encode(map[string]any{"images": []string{img}})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(fs.Close)
	return fs
}

func (fs *fakeServer) body(path string) map[string]any {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	return fs.bodies[path]
}

type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

func resetFlags(t *testing.T) {
	t.Helper()
	reset := func(f *pflag.Flag) {
		f.Value.Set(f.DefValue)
		f.Changed = false
	}
	rootCmd.PersistentFlags().VisitAll(reset)
	for _, c := range rootCmd.Commands() {
		c.Flags().VisitAll(reset)
	}
}

// runCLI executes sdctl with args and returns what was written to stdout.
func runCLI(t *testing.T, args ...string) string {
	t.Helper()
	out, err := runCLIErr(t, args...)
	if err != nil {
		t.Fatalf("sdctl %v failed: %v", args, err)
	}
	return out
}

// runCLIErr executes sdctl with args and returns stdout and the command error.
func runCLIErr(t *testing.T, args ...string) (string, error) {
	t.Helper()
	resetFlags(t)
	t.Cleanup(func() { resetFlags(t) })

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = w
	done := make(chan []byte)
	go func() {
		data, _ := io.ReadAll(r)
		done <- data
	}()

	configFile := filepath.Join(t.TempDir(), "missing-config.yaml")
	rootCmd.SetArgs(append([]string{"--config", configFile}, args...))
	execErr := rootCmd.Execute()

	w.Close()
	os.Stdout = origStdout
	out := <-done
	return string(out), execErr
}

func writeParams(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "params.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write params: %v", err)
	}
	return path
}

func writePNG(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "input.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, 64, 64))); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	return path
}

func assertStdoutIsPathsIn(t *testing.T, stdout, dir string, n int) []string {
	t.Helper()
	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != n {
		t.Fatalf("expected %d line(s) on stdout, got %q", n, stdout)
	}
	for _, l := range lines {
		if filepath.Dir(l) != dir {
			t.Errorf("expected path in %s, got %q", dir, l)
		}
		if _, err := os.Stat(l); err != nil {
			t.Errorf("stdout line is not a saved file: %q", l)
		}
	}
	return lines
}

type cliCase struct {
	name     string
	endpoint string
	args     func(t *testing.T) []string
}

func cliCases() []cliCase {
	return []cliCase{
		{"txt2img", "/sdapi/v1/txt2img", func(t *testing.T) []string {
			return []string{"txt2img", "a cat"}
		}},
		{"img2img", "/sdapi/v1/img2img", func(t *testing.T) []string {
			return []string{"img2img", "a cat", writePNG(t)}
		}},
		{"hires", "/sdapi/v1/img2img", func(t *testing.T) []string {
			return []string{"hires", "a cat", writePNG(t)}
		}},
	}
}

func TestCLI_EnvDefaultsApplied(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			outDir := filepath.Join(t.TempDir(), "images")
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", writeParams(t, "steps: 33\n"))
			t.Setenv("SDCTL_OUTPUT_DIR", outDir)

			stdout := runCLI(t, tc.args(t)...)

			assertStdoutIsPathsIn(t, stdout, outDir, 1)
			if got := srv.body(tc.endpoint)["steps"]; got != float64(33) {
				t.Errorf("expected steps 33 from SDCTL_PARAMS, got %v", got)
			}
		})
	}
}

func TestCLI_FlagsOverrideEnvDefaults(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			envOut := filepath.Join(t.TempDir(), "env-images")
			flagOut := filepath.Join(t.TempDir(), "flag.png")
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", writeParams(t, "steps: 33\n"))
			t.Setenv("SDCTL_OUTPUT_DIR", envOut)

			args := append(tc.args(t), "--params", writeParams(t, "steps: 44\n"), "-o", flagOut)
			stdout := runCLI(t, args...)

			if strings.TrimRight(stdout, "\n") != flagOut {
				t.Errorf("expected stdout %q, got %q", flagOut, stdout)
			}
			if _, err := os.Stat(envOut); !os.IsNotExist(err) {
				t.Errorf("SDCTL_OUTPUT_DIR should not be created when -o is given")
			}
			if got := srv.body(tc.endpoint)["steps"]; got != float64(44) {
				t.Errorf("expected steps 44 from --params, got %v", got)
			}
		})
	}
}

func TestCLI_EmptyParamsFlagDisablesEnvDefault(t *testing.T) {
	srv := newFakeServer(t)
	t.Setenv("SDCTL_URL", srv.URL)
	t.Setenv("SDCTL_PARAMS", filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	t.Setenv("SDCTL_OUTPUT_DIR", t.TempDir())

	runCLI(t, "txt2img", "a cat", "--params", "")

	if got := srv.body("/sdapi/v1/txt2img")["steps"]; got != float64(20) {
		t.Errorf("expected built-in default steps 20, got %v", got)
	}
}

func TestCLI_NoEnv_KeepsCurrentBehavior(t *testing.T) {
	srv := newFakeServer(t)
	t.Setenv("SDCTL_URL", srv.URL)
	t.Setenv("SDCTL_PARAMS", "")
	os.Unsetenv("SDCTL_PARAMS")
	t.Setenv("SDCTL_OUTPUT_DIR", "")
	os.Unsetenv("SDCTL_OUTPUT_DIR")

	cwd := t.TempDir()
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(cwd)

	stdout := runCLI(t, "txt2img", "a cat")

	assertStdoutIsPathsIn(t, stdout, ".", 1)
	if got := srv.body("/sdapi/v1/txt2img")["steps"]; got != float64(20) {
		t.Errorf("expected built-in default steps 20, got %v", got)
	}
}
