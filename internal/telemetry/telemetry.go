package telemetry

import (
	"fmt"
	"strings"
)

// BatteryTelemetry holds structured EV battery readings.
// Supports JSON and CSV input via Parse.
//
// JSON: {"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving"}
// CSV:  VIN_789,55.0,3.1[,soc%][,mode][,vehicleModel]
type BatteryTelemetry struct {
	VIN                  string   `json:"vin"`
	BatteryTempC         float64  `json:"batteryTempC"`
	VoltageV             float64  `json:"voltageV"`
	StateOfChargePercent *float64 `json:"stateOfChargePercent,omitempty"`
	DrivingMode          string   `json:"drivingMode"`
	Timestamp            string   `json:"timestamp,omitempty"`
	VehicleModel         string   `json:"vehicleModel,omitempty"`
}

// Validate checks required fields, sane ranges, and auto-detects vehicle model from VIN.
func (t *BatteryTelemetry) Validate() error {
	if strings.TrimSpace(t.VIN) == "" {
		return fmt.Errorf("VIN is required")
	}
	if t.BatteryTempC < -50 || t.BatteryTempC > 200 {
		return fmt.Errorf("batteryTempC out of range: %v", t.BatteryTempC)
	}
	if t.VoltageV < 0 || t.VoltageV > 10 {
		return fmt.Errorf("voltageV out of range: %v", t.VoltageV)
	}
	if t.StateOfChargePercent != nil && (*t.StateOfChargePercent < 0 || *t.StateOfChargePercent > 100) {
		return fmt.Errorf("stateOfChargePercent out of range: %v", *t.StateOfChargePercent)
	}
	if strings.TrimSpace(t.VehicleModel) == "" {
		t.VehicleModel = DetectModelFromVIN(t.VIN)
	} else {
		t.VehicleModel = strings.ToUpper(t.VehicleModel)
	}
	return nil
}

// DetectModelFromVIN derives the Rivian model from the VIN prefix.
// 7FCLS → R1S, 7FCTL → R1T, otherwise UNKNOWN.
func DetectModelFromVIN(vin string) string {
	if len(vin) < 6 {
		return "UNKNOWN"
	}
	prefix := strings.ToUpper(vin[:6])
	switch {
	case strings.HasPrefix(prefix, "7FCLS"):
		return "R1S"
	case strings.HasPrefix(prefix, "7FCTL"):
		return "R1T"
	default:
		return "UNKNOWN"
	}
}
