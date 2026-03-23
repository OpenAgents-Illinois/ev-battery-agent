package com.ev.battery.agent;

/**
 * Structured representation of EV battery telemetry data.
 *
 * Supported input formats (handled by TelemetryParser):
 *   JSON: {"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving"}
 *   CSV:  VIN_789,55.0,3.1,82.0,driving
 *         Columns: vin, batteryTempC, voltageV [, stateOfChargePercent] [, drivingMode]
 */
public class BatteryTelemetry {
    public String vin;
    public double batteryTempC;
    public double voltageV;
    public Double stateOfChargePercent; // nullable — not always available
    public String drivingMode = "unknown";
    public String timestamp; // nullable — ISO-8601 if provided

    /** Validates required fields and sane ranges. Throws if invalid. */
    public void validate() {
        if (vin == null || vin.isBlank()) throw new IllegalArgumentException("VIN is required.");
        if (batteryTempC < -50 || batteryTempC > 200)
            throw new IllegalArgumentException("batteryTempC out of range: " + batteryTempC);
        if (voltageV < 0 || voltageV > 10)
            throw new IllegalArgumentException("voltageV out of range: " + voltageV);
        if (stateOfChargePercent != null && (stateOfChargePercent < 0 || stateOfChargePercent > 100))
            throw new IllegalArgumentException("stateOfChargePercent out of range: " + stateOfChargePercent);
    }

    /** Formats telemetry as a readable string for the agent prompt. */
    public String toPromptString() {
        StringBuilder sb = new StringBuilder();
        sb.append("VIN: ").append(vin);
        sb.append(", Battery Temp: ").append(batteryTempC).append("C");
        sb.append(", Voltage: ").append(voltageV).append("V");
        if (stateOfChargePercent != null)
            sb.append(", State of Charge: ").append(stateOfChargePercent).append("%");
        sb.append(", Mode: ").append(drivingMode);
        if (timestamp != null)
            sb.append(", Timestamp: ").append(timestamp);
        return sb.toString();
    }
}
