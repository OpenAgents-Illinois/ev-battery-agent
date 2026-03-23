package telemetry

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// Parse parses telemetry from JSON or CSV input, validates it, and returns a BatteryTelemetry.
func Parse(input string) (*BatteryTelemetry, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("input is empty")
	}

	var (
		t   *BatteryTelemetry
		err error
	)
	if strings.HasPrefix(input, "{") {
		t, err = parseJSON(input)
	} else {
		t, err = parseCSV(input)
	}
	if err != nil {
		return nil, err
	}
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

func parseJSON(input string) (*BatteryTelemetry, error) {
	var t BatteryTelemetry
	if err := json.Unmarshal([]byte(input), &t); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return &t, nil
}

func parseCSV(input string) (*BatteryTelemetry, error) {
	parts := strings.Split(input, ",")
	if len(parts) < 3 {
		return nil, fmt.Errorf("CSV requires at least 3 columns: vin,batteryTempC,voltageV — got %d", len(parts))
	}

	t := &BatteryTelemetry{DrivingMode: "unknown"}
	t.VIN = strings.TrimSpace(parts[0])

	var err error
	if t.BatteryTempC, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err != nil {
		return nil, fmt.Errorf("invalid batteryTempC: %w", err)
	}
	if t.VoltageV, err = strconv.ParseFloat(strings.TrimSpace(parts[2]), 64); err != nil {
		return nil, fmt.Errorf("invalid voltageV: %w", err)
	}
	if len(parts) >= 4 && strings.TrimSpace(parts[3]) != "" {
		soc, err := strconv.ParseFloat(strings.TrimSpace(parts[3]), 64)
		if err != nil {
			return nil, fmt.Errorf("invalid stateOfChargePercent: %w", err)
		}
		t.StateOfChargePercent = &soc
	}
	if len(parts) >= 5 && strings.TrimSpace(parts[4]) != "" {
		t.DrivingMode = strings.TrimSpace(parts[4])
	}
	if len(parts) >= 6 && strings.TrimSpace(parts[5]) != "" {
		t.VehicleModel = strings.TrimSpace(parts[5])
	}
	return t, nil
}
