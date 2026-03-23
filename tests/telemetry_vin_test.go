package tests

import (
	"testing"

	"ev-battery-agent/internal/telemetry"
)

func TestDetectModelFromVINR1S(t *testing.T) {
	if got := telemetry.DetectModelFromVIN("7FCLS12345678901"); got != "R1S" {
		t.Errorf("got %q, want %q", got, "R1S")
	}
}

func TestDetectModelFromVINR1T(t *testing.T) {
	if got := telemetry.DetectModelFromVIN("7FCTL12345678901"); got != "R1T" {
		t.Errorf("got %q, want %q", got, "R1T")
	}
}

func TestDetectModelFromVINUnknown(t *testing.T) {
	if got := telemetry.DetectModelFromVIN("VIN_789"); got != "UNKNOWN" {
		t.Errorf("got %q, want %q", got, "UNKNOWN")
	}
}
