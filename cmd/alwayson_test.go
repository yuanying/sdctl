package cmd

import (
	"path/filepath"
	"testing"
)

const adetailerParams = `
alwayson_scripts:
  ADetailer:
    args:
      - true
      - false
      - ad_model: face_yolov8n.pt
        ad_denoising_strength: 0.4
`

func assertADetailerInBody(t *testing.T, body map[string]any) {
	t.Helper()
	scripts, ok := body["alwayson_scripts"].(map[string]any)
	if !ok {
		t.Fatalf("expected alwayson_scripts in request body, got %v", body["alwayson_scripts"])
	}
	adetailer, ok := scripts["ADetailer"].(map[string]any)
	if !ok {
		t.Fatalf("expected ADetailer in alwayson_scripts, got %v", scripts)
	}
	args, ok := adetailer["args"].([]any)
	if !ok || len(args) != 3 {
		t.Fatalf("unexpected ADetailer args: %v", adetailer["args"])
	}
	if args[0] != true || args[1] != false {
		t.Errorf("unexpected leading args: %v", args[:2])
	}
	unit, ok := args[2].(map[string]any)
	if !ok || unit["ad_model"] != "face_yolov8n.pt" || unit["ad_denoising_strength"] != 0.4 {
		t.Errorf("unexpected ADetailer unit: %v", args[2])
	}
}

func TestCLI_AlwaysonScriptsFromParams(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", "")
			t.Setenv("SDCTL_OUTPUT_DIR", t.TempDir())

			args := append(tc.args(t), "--params", writeParams(t, adetailerParams))
			runCLI(t, args...)

			assertADetailerInBody(t, srv.body(tc.endpoint))
		})
	}
}

func TestCLI_AlwaysonScriptsFromEnvParams(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", writeParams(t, adetailerParams))
			t.Setenv("SDCTL_OUTPUT_DIR", t.TempDir())

			runCLI(t, tc.args(t)...)

			assertADetailerInBody(t, srv.body(tc.endpoint))
		})
	}
}

func TestCLI_NoAlwaysonScripts_OmittedFromBody(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", "")
			t.Setenv("SDCTL_OUTPUT_DIR", t.TempDir())

			args := append(tc.args(t), "--params", writeParams(t, "steps: 25\n"))
			runCLI(t, args...)

			if _, ok := srv.body(tc.endpoint)["alwayson_scripts"]; ok {
				t.Errorf("expected no alwayson_scripts in request body, got %v", srv.body(tc.endpoint)["alwayson_scripts"])
			}
		})
	}
}

// --params replaces SDCTL_PARAMS as a whole file, so alwayson_scripts written
// only in the default params does not leak into the request.
func TestCLI_ParamsFlagReplacesEnvAlwaysonScripts(t *testing.T) {
	srv := newFakeServer(t)
	t.Setenv("SDCTL_URL", srv.URL)
	t.Setenv("SDCTL_PARAMS", writeParams(t, adetailerParams))
	t.Setenv("SDCTL_OUTPUT_DIR", filepath.Join(t.TempDir(), "images"))

	runCLI(t, "txt2img", "a cat", "--params", writeParams(t, "steps: 25\n"))

	if got, ok := srv.body("/sdapi/v1/txt2img")["alwayson_scripts"]; ok {
		t.Errorf("expected --params to replace SDCTL_PARAMS, got alwayson_scripts %v", got)
	}
}
