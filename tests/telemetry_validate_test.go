package tests

import (
	"testing"

	"ev-battery-agent/internal/telemetry"
)

func TestValidateRejectsBlankVIN(t *testing.T) {
	_, err := telemetry.Parse(",55.0,3.1")
	if err == nil {
		t.Error("expected error for blank VIN, got nil")
	}
}

func TestValidateRejectsOutOfRangeTemp(t *testing.T) {
	_, err := telemetry.Parse("VIN_001,999.0,3.7")
	if err == nil {
		t.Error("expected error for temp 999.0, got nil")
	}
}

func TestValidateRejectsOutOfRangeVoltage(t *testing.T) {
	_, err := telemetry.Parse("VIN_001,30.0,99.0")
	if err == nil {
		t.Error("expected error for voltage 99.0, got nil")
	}
}
