package com.ev.battery.agent;

import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.*;

class TelemetryParserTest {

    @Test
    void testParseJsonFull() {
        String input = "{\"vin\":\"VIN_789\",\"batteryTempC\":55.0,\"voltageV\":3.1,\"stateOfChargePercent\":82.0,\"drivingMode\":\"driving\"}";
        BatteryTelemetry t = TelemetryParser.parse(input);
        assertEquals("VIN_789", t.vin);
        assertEquals(55.0, t.batteryTempC);
        assertEquals(3.1, t.voltageV);
        assertEquals(82.0, t.stateOfChargePercent);
        assertEquals("driving", t.drivingMode);
    }

    @Test
    void testParseJsonMinimal() {
        String input = "{\"vin\":\"VIN_001\",\"batteryTempC\":30.0,\"voltageV\":3.7}";
        BatteryTelemetry t = TelemetryParser.parse(input);
        assertEquals("VIN_001", t.vin);
        assertNull(t.stateOfChargePercent);
        assertEquals("unknown", t.drivingMode);
    }

    @Test
    void testParseCsvFull() {
        BatteryTelemetry t = TelemetryParser.parse("VIN_789,55.0,3.1,82.0,driving");
        assertEquals("VIN_789", t.vin);
        assertEquals(55.0, t.batteryTempC);
        assertEquals(3.1, t.voltageV);
        assertEquals(82.0, t.stateOfChargePercent);
        assertEquals("driving", t.drivingMode);
    }

    @Test
    void testParseCsvMinimal() {
        BatteryTelemetry t = TelemetryParser.parse("VIN_001,30.0,3.7");
        assertEquals("VIN_001", t.vin);
        assertNull(t.stateOfChargePercent);
        assertEquals("unknown", t.drivingMode);
    }

    @Test
    void testValidationRejectsBlankVin() {
        assertThrows(IllegalArgumentException.class, () ->
            TelemetryParser.parse(",55.0,3.1"));
    }

    @Test
    void testValidationRejectsOutOfRangeTemp() {
        assertThrows(IllegalArgumentException.class, () ->
            TelemetryParser.parse("VIN_001,999.0,3.7"));
    }

    @Test
    void testValidationRejectsOutOfRangeVoltage() {
        assertThrows(IllegalArgumentException.class, () ->
            TelemetryParser.parse("VIN_001,30.0,99.0"));
    }

    @Test
    void testToPromptStringIncludesAllFields() {
        BatteryTelemetry t = TelemetryParser.parse("VIN_789,55.0,3.1,82.0,driving");
        String prompt = t.toPromptString();
        assertTrue(prompt.contains("VIN_789"));
        assertTrue(prompt.contains("55.0"));
        assertTrue(prompt.contains("3.1"));
        assertTrue(prompt.contains("82.0"));
        assertTrue(prompt.contains("driving"));
    }
}
