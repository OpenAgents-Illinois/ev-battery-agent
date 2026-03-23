package com.ev.battery.agent;

import com.google.gson.Gson;
import com.google.gson.JsonSyntaxException;

/**
 * Parses battery telemetry from JSON or CSV strings into a BatteryTelemetry object.
 *
 * JSON example:
 *   {"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving"}
 *
 * CSV example (columns: vin, batteryTempC, voltageV [,stateOfChargePercent] [,drivingMode]):
 *   VIN_789,55.0,3.1,82.0,driving
 *   VIN_789,55.0,3.1
 */
public class TelemetryParser {
    private static final Gson GSON = new Gson();

    public static BatteryTelemetry parse(String input) {
        if (input == null || input.isBlank()) {
            throw new IllegalArgumentException("Input is empty.");
        }

        String trimmed = input.trim();
        BatteryTelemetry telemetry = trimmed.startsWith("{") ? parseJson(trimmed) : parseCsv(trimmed);
        telemetry.validate();
        return telemetry;
    }

    private static BatteryTelemetry parseJson(String json) {
        try {
            BatteryTelemetry t = GSON.fromJson(json, BatteryTelemetry.class);
            if (t == null) throw new IllegalArgumentException("JSON parsed to null.");
            return t;
        } catch (JsonSyntaxException e) {
            throw new IllegalArgumentException("Invalid JSON: " + e.getMessage());
        }
    }

    private static BatteryTelemetry parseCsv(String csv) {
        String[] parts = csv.split(",", -1);
        if (parts.length < 3) {
            throw new IllegalArgumentException(
                "CSV requires at least 3 columns: vin,batteryTempC,voltageV. Got: " + parts.length);
        }

        BatteryTelemetry t = new BatteryTelemetry();
        try {
            t.vin = parts[0].trim();
            t.batteryTempC = Double.parseDouble(parts[1].trim());
            t.voltageV = Double.parseDouble(parts[2].trim());
            if (parts.length >= 4 && !parts[3].isBlank())
                t.stateOfChargePercent = Double.parseDouble(parts[3].trim());
            if (parts.length >= 5 && !parts[4].isBlank())
                t.drivingMode = parts[4].trim();
        } catch (NumberFormatException e) {
            throw new IllegalArgumentException("CSV number parse error: " + e.getMessage());
        }
        return t;
    }
}
