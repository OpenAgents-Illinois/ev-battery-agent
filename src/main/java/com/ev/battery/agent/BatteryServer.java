package com.ev.battery.agent;

import com.google.gson.Gson;
import io.javalin.Javalin;

/**
 * HTTP server exposing the EV Battery Agent as a REST API.
 *
 * Endpoints:
 *   GET  /health          — liveness check, returns 200 OK
 *   POST /analyze         — accepts BatteryTelemetry as JSON or CSV, returns analysis
 *
 * POST /analyze request body (JSON):
 *   {"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving"}
 *
 * POST /analyze response (JSON):
 *   {"vin":"VIN_789","analysis":"Battery temperature of 55C exceeds... Ticket created: KAN-42"}
 *
 * POST /analyze error response (JSON, 400):
 *   {"error":"batteryTempC out of range: 999.0"}
 */
public class BatteryServer {
    private static final Gson GSON = new Gson();
    private static final String ANALYZE_PROMPT =
        "Analyze this telemetry: %s. If this violates safety thresholds, file a Jira ticket with the appropriate severity (CRITICAL, WARNING, or INFO).";

    public static void start(AgentFactory factory, int port) {
        Javalin app = Javalin.create(config -> {
            config.useVirtualThreads = true;
        }).start(port);

        app.get("/health", ctx -> ctx.result("OK"));

        app.post("/analyze", ctx -> {
            BatteryTelemetry telemetry;
            try {
                telemetry = TelemetryParser.parse(ctx.body());
            } catch (IllegalArgumentException e) {
                ctx.status(400)
                   .contentType("application/json")
                   .result(GSON.toJson(new ErrorResponse(e.getMessage())));
                return;
            }

            EvExpert agent = factory.newAgent(telemetry.vehicleModel);
            String analysis = agent.chat(ANALYZE_PROMPT.formatted(telemetry.toPromptString()));

            ctx.contentType("application/json")
               .result(GSON.toJson(new AnalysisResponse(telemetry.vin, analysis)));
        });

        System.out.println("Server listening on http://localhost:" + port);
        System.out.println("  POST /analyze  — submit battery telemetry JSON or CSV");
        System.out.println("  GET  /health   — liveness check");
    }

    record AnalysisResponse(String vin, String analysis) {}
    record ErrorResponse(String error) {}
}
