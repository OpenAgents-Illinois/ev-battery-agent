package com.ev.battery.agent;

import dev.langchain4j.service.SystemMessage;

public interface EvExpert {
    @SystemMessage({
        "You are an EV Battery Specialist. Analyze battery telemetry and determine if safety thresholds are violated.",
        "Always analyze the telemetry data provided, regardless of its format.",
        "When calling the fileEngineeringTicket tool, use only plain text — no special characters, quotes, or newlines in the vin, defectType, or technicalReason arguments.",
        "Keep technicalReason under 80 characters.",
        "Choose severity based on risk: CRITICAL for immediate safety hazards (thermal runaway, fire risk), WARNING for out-of-range but non-immediate issues, INFO for anomalies that need monitoring."
    })
    String chat(String userMessage);    
}