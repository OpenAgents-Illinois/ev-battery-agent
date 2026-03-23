package tests

import (
	"strings"
	"testing"

	"ev-battery-agent/internal/telemetry"
)

// --- Parse: JSON ---

func TestParseJSONFull(t *testing.T) {
	input := `{"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving"}`
	tel, err := telemetry.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tel.VIN != "VIN_789" {
		t.Errorf("VIN: got %q, want %q", tel.VIN, "VIN_789")
	}
	if tel.BatteryTempC != 55.0 {
		t.Errorf("BatteryTempC: got %v, want 55.0", tel.BatteryTempC)
	}
	if tel.VoltageV != 3.1 {
		t.Errorf("VoltageV: got %v, want 3.1", tel.VoltageV)
	}
	if tel.StateOfChargePercent == nil || *tel.StateOfChargePercent != 82.0 {
		t.Errorf("StateOfChargePercent: got %v, want 82.0", tel.StateOfChargePercent)
	}
	if tel.DrivingMode != "driving" {
		t.Errorf("DrivingMode: got %q, want %q", tel.DrivingMode, "driving")
	}
}

func TestParseJSONMinimal(t *testing.T) {
	input := `{"vin":"VIN_001","batteryTempC":30.0,"voltageV":3.7}`
	tel, err := telemetry.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tel.VIN != "VIN_001" {
		t.Errorf("VIN: got %q, want %q", tel.VIN, "VIN_001")
	}
	if tel.StateOfChargePercent != nil {
		t.Errorf("StateOfChargePercent: expected nil, got %v", *tel.StateOfChargePercent)
	}
	// JSON omits drivingMode → empty string (CSV defaults to "unknown")
	if tel.DrivingMode != "" {
		t.Errorf("DrivingMode: got %q, want empty string", tel.DrivingMode)
	}
}

func TestParseJSONVehicleModel(t *testing.T) {
	input := `{"vin":"VIN_789","batteryTempC":30.0,"voltageV":3.7,"vehicleModel":"R1T"}`
	tel, err := telemetry.Parse(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tel.VehicleModel != "R1T" {
		t.Errorf("VehicleModel: got %q, want %q", tel.VehicleModel, "R1T")
	}
}

// --- Parse: CSV ---

func TestParseCSVFull(t *testing.T) {
	tel, err := telemetry.Parse("VIN_789,55.0,3.1,82.0,driving")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tel.VIN != "VIN_789" {
		t.Errorf("VIN: got %q, want %q", tel.VIN, "VIN_789")
	}
	if tel.BatteryTempC != 55.0 {
		t.Errorf("BatteryTempC: got %v, want 55.0", tel.BatteryTempC)
	}
	if tel.VoltageV != 3.1 {
		t.Errorf("VoltageV: got %v, want 3.1", tel.VoltageV)
	}
	if tel.StateOfChargePercent == nil || *tel.StateOfChargePercent != 82.0 {
		t.Errorf("StateOfChargePercent: got %v, want 82.0", tel.StateOfChargePercent)
	}
	if tel.DrivingMode != "driving" {
		t.Errorf("DrivingMode: got %q, want %q", tel.DrivingMode, "driving")
	}
}

func TestParseCSVMinimal(t *testing.T) {
	tel, err := telemetry.Parse("VIN_001,30.0,3.7")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tel.VIN != "VIN_001" {
		t.Errorf("VIN: got %q, want %q", tel.VIN, "VIN_001")
	}
	if tel.StateOfChargePercent != nil {
		t.Errorf("StateOfChargePercent: expected nil, got %v", *tel.StateOfChargePercent)
	}
	if tel.DrivingMode != "unknown" {
		t.Errorf("DrivingMode: got %q, want %q", tel.DrivingMode, "unknown")
	}
}

func TestParseCSVVehicleModel(t *testing.T) {
	tel, err := telemetry.Parse("VIN_789,30.0,3.7,,,R1S")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tel.VehicleModel != "R1S" {
		t.Errorf("VehicleModel: got %q, want %q", tel.VehicleModel, "R1S")
	}
}

// --- Validate: rejection cases ---

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

// --- DetectModelFromVIN ---

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

// --- ToPromptString ---

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
