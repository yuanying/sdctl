package api_test

import (
	"testing"

	"github.com/yuanying/sdctl/internal/api"
)

func TestValidateSamplerName(t *testing.T) {
	samplers := []api.Sampler{
		{Name: "Euler a"},
		{Name: "DPM++ 2M"},
	}

	if err := api.ValidateSamplerName(samplers, "Euler a"); err != nil {
		t.Errorf("expected no error for valid sampler, got: %v", err)
	}

	err := api.ValidateSamplerName(samplers, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid sampler, got nil")
	}
	msg := err.Error()
	if !contains(msg, "invalid") {
		t.Errorf("error should mention the invalid value, got: %s", msg)
	}
	if !contains(msg, "Euler a") {
		t.Errorf("error should list available samplers, got: %s", msg)
	}
}

func TestValidateSchedulerName(t *testing.T) {
	schedulers := []api.Scheduler{
		{Name: "automatic", Label: "Automatic"},
		{Name: "karras", Label: "Karras"},
	}

	if err := api.ValidateSchedulerName(schedulers, "karras"); err != nil {
		t.Errorf("expected no error for valid scheduler, got: %v", err)
	}

	err := api.ValidateSchedulerName(schedulers, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid scheduler, got nil")
	}
	msg := err.Error()
	if !contains(msg, "invalid") {
		t.Errorf("error should mention the invalid value, got: %s", msg)
	}
	if !contains(msg, "karras") {
		t.Errorf("error should list available schedulers, got: %s", msg)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
