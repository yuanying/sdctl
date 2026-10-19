package cmd

import "testing"

func TestCLI_DistilledCFGScaleFromFlag(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", "")
			t.Setenv("SDCTL_OUTPUT_DIR", t.TempDir())

			args := append(tc.args(t), "--distilled-cfg-scale", "5")
			runCLI(t, args...)

			if got := srv.body(tc.endpoint)["distilled_cfg_scale"]; got != 5.0 {
				t.Errorf("expected distilled_cfg_scale 5, got %v", got)
			}
		})
	}
}

func TestCLI_DistilledCFGScaleFromParams(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", "")
			t.Setenv("SDCTL_OUTPUT_DIR", t.TempDir())

			args := append(tc.args(t), "--params", writeParams(t, "distilled_cfg_scale: 5\n"))
			runCLI(t, args...)

			if got := srv.body(tc.endpoint)["distilled_cfg_scale"]; got != 5.0 {
				t.Errorf("expected distilled_cfg_scale 5, got %v", got)
			}
		})
	}
}

func TestCLI_DistilledCFGScaleFlagOverridesParams(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", "")
			t.Setenv("SDCTL_OUTPUT_DIR", t.TempDir())

			args := append(tc.args(t),
				"--params", writeParams(t, "distilled_cfg_scale: 5\n"),
				"--distilled-cfg-scale", "3.5")
			runCLI(t, args...)

			if got := srv.body(tc.endpoint)["distilled_cfg_scale"]; got != 3.5 {
				t.Errorf("expected distilled_cfg_scale 3.5, got %v", got)
			}
		})
	}
}

// Without a flag or params value, the field is left out so the server keeps
// its own default (Forge: 3.5) and non-Forge servers never see it.
func TestCLI_NoDistilledCFGScale_OmittedFromBody(t *testing.T) {
	for _, tc := range cliCases() {
		t.Run(tc.name, func(t *testing.T) {
			srv := newFakeServer(t)
			t.Setenv("SDCTL_URL", srv.URL)
			t.Setenv("SDCTL_PARAMS", "")
			t.Setenv("SDCTL_OUTPUT_DIR", t.TempDir())

			args := append(tc.args(t), "--params", writeParams(t, "steps: 25\n"))
			runCLI(t, args...)

			if got, ok := srv.body(tc.endpoint)["distilled_cfg_scale"]; ok {
				t.Errorf("expected no distilled_cfg_scale in request body, got %v", got)
			}
		})
	}
}
