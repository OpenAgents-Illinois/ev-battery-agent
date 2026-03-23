package com.ev.battery.agent;

/**
 * Structured representation of EV battery telemetry data.
 *
 * Supported input formats (handled by TelemetryParser):
 *   JSON: {"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving","vehicleModel":"R1S"}
 *   CSV:  VIN_789,55.0,3.1,82.0,driving[,R1S]
 *         Columns: vin, batteryTempC, voltageV [, stateOfChargePercent] [, drivingMode] [, vehicleModel]
 *
 * vehicleModel is optional — auto-detected from VIN prefix if omitted:
 *   7FCLS... → R1S,  7FCTL... → R1T,  otherwise → UNKNOWN
 */
public class BatteryTelemetry {
    public String vin;
    public double batteryTempC;
    public double voltageV;
    public Double stateOfChargePercent; // nullable — not always available
    public String drivingMode = "unknown";
    public String timestamp;   // nullable — ISO-8601 if provided
    public String vehicleModel; // nullable — auto-detected from VIN after parsing

    /** Validates required fields and sane ranges. Throws if invalid. */
    public void validate() {
        if (vin == null || vin.isBlank()) throw new IllegalArgumentException("VIN is required.");
        if (batteryTempC < -50 || batteryTempC > 200)
            throw new IllegalArgumentException("batteryTempC out of range: " + batteryTempC);
        if (voltageV < 0 || voltageV > 10)
            throw new IllegalArgumentException("voltageV out of range: " + voltageV);
        if (stateOfChargePercent != null && (stateOfChargePercent < 0 || stateOfChargePercent > 100))
            throw new IllegalArgumentException("stateOfChargePercent out of range: " + stateOfChargePercent);

        // Detect model from VIN if not explicitly set
        if (vehicleModel == null || vehicleModel.isBlank()) {
            vehicleModel = detectModelFromVin(vin);
        } else {
            vehicleModel = vehicleModel.toUpperCase();
        }
    }

    /**
     * Derives the Rivian model from the VIN's 4th-6th characters (World Manufacturer Identifier).
     * Rivian VINs: 7FCLS → R1S, 7FCTL → R1T. Falls back to UNKNOWN for non-Rivian or test VINs.
     */
    static String detectModelFromVin(String vin) {
        if (vin == null || vin.length() < 6) return "UNKNOWN";
        String prefix = vin.substring(0, 6).toUpperCase();
        if (prefix.startsWith("7FCLS")) return "R1S";
        if (prefix.startsWith("7FCTL")) return "R1T";
        return "UNKNOWN";
    }

    /** Formats telemetry as a readable string for the agent prompt. */
    public String toPromptString() {
        StringBuilder sb = new StringBuilder();
        sb.append("VIN: ").append(vin);
        sb.append(", Vehicle: ").append(vehicleModel != null ? vehicleModel : "UNKNOWN");
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
