package api

import (
	"fmt"
	"strings"
)

func ValidateSamplerName(samplers []Sampler, name string) error {
	names := make([]string, len(samplers))
	for i, s := range samplers {
		if s.Name == name {
			return nil
		}
		names[i] = s.Name
	}
	return fmt.Errorf("invalid sampler %q, available: %s", name, strings.Join(names, ", "))
}

func ValidateSchedulerName(schedulers []Scheduler, name string) error {
	names := make([]string, len(schedulers))
	for i, s := range schedulers {
		if s.Name == name {
			return nil
		}
		names[i] = s.Name
	}
	return fmt.Errorf("invalid scheduler %q, available: %s", name, strings.Join(names, ", "))
}

func ValidateModelName(models []Model, name string) error {
	names := make([]string, len(models))
	for i, m := range models {
		if m.ModelName == name {
			return nil
		}
		names[i] = m.ModelName
	}
	return fmt.Errorf("invalid model %q, available: %s", name, strings.Join(names, ", "))
}
