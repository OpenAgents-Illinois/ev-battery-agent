package telemetry

import (
	"fmt"
	"strings"
)

// ToPromptString formats telemetry as plain English bullet points for the agent.
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
