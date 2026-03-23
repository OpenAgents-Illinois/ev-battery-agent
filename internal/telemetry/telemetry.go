package telemetry

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// BatteryTelemetry holds structured EV battery readings.
// Supports JSON and CSV input (handled by Parse).
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

// Parse parses telemetry from JSON or CSV input and validates it.
func Parse(input string) (*BatteryTelemetry, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("input is empty")
	}

	var t *BatteryTelemetry
	var err error
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
	t.BatteryTempC, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid batteryTempC: %w", err)
	}
	t.VoltageV, err = strconv.ParseFloat(strings.TrimSpace(parts[2]), 64)
	if err != nil {
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

// Validate checks required fields and sane ranges, and auto-detects vehicle model from VIN.
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

// DetectModelFromVIN derives the Rivian model from VIN prefix.
// 7FCLS → R1S, 7FCTL → R1T, otherwise UNKNOWN.
func DetectModelFromVIN(vin string) string {
	if len(vin) < 6 {
		return "UNKNOWN"
	}
	prefix := strings.ToUpper(vin[:6])
	if strings.HasPrefix(prefix, "7FCLS") {
		return "R1S"
	}
	if strings.HasPrefix(prefix, "7FCTL") {
		return "R1T"
	}
	return "UNKNOWN"
}

// ToPromptString formats telemetry as plain English for the agent.
func (t *BatteryTelemetry) ToPromptString() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("A Rivian %s has reported the following battery readings:\n", t.VehicleModel))
	sb.WriteString(fmt.Sprintf("- VIN: %s\n", t.VIN))
	sb.WriteString(fmt.Sprintf("- Battery Temperature: %.1f degrees Celsius\n", t.BatteryTempC))
	sb.WriteString(fmt.Sprintf("- Cell Voltage: %.2f V\n", t.VoltageV))
	if t.StateOfChargePercent != nil {
		sb.WriteString(fmt.Sprintf("- State of Charge: %.1f%%\n", *t.StateOfChargePercent))
	}
	sb.WriteString(fmt.Sprintf("- Driving Mode: %s\n", t.DrivingMode))
	sb.WriteString("\nAre any of these readings outside safe operating limits? ")
	sb.WriteString("If yes, file a Jira ticket using the fileEngineeringTicket tool with severity CRITICAL, WARNING, or INFO.")
	return sb.String()
}
