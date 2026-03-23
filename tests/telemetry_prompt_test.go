package tests

import (
	"strings"
	"testing"

	"ev-battery-agent/internal/telemetry"
)

func TestToPromptStringIncludesAllFields(t *testing.T) {
	tel, err := telemetry.Parse("VIN_789,55.0,3.1,82.0,driving")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	prompt := tel.ToPromptString()
	for _, want := range []string{"VIN_789", "55.0", "3.1", "82.0", "driving"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}
