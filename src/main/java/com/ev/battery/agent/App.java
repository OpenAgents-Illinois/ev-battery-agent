package com.ev.battery.agent;

import java.util.Scanner;

import io.github.cdimascio.dotenv.Dotenv;

public class App {

    private static final String ANALYZE_PROMPT =
        "Analyze this telemetry: %s. If this violates safety thresholds, file a Jira ticket with the appropriate severity (CRITICAL, WARNING, or INFO).";

    public static void main(String[] args) {
        Dotenv dotenv = Dotenv.load();
        String projectId = dotenv.get("GCLOUD_PROJECT_ID");

        System.out.println("=== EV Battery Agent ===");
        System.out.println("Loading documents and initializing agent...");
        AgentFactory factory = new AgentFactory(projectId, "us-central1");
        System.out.println("Agent ready.");
        System.out.println();

        if (args.length > 0 && args[0].equals("--server")) {
            int port = args.length > 1 ? Integer.parseInt(args[1]) : 8080;
            BatteryServer.start(factory, port);
        } else {
            runInteractive(factory);
        }
    }

    private static void runInteractive(AgentFactory factory) {
        EvExpert agent = factory.newAgent();
        System.out.println("Paste telemetry as JSON or CSV, or type 'exit' to quit.");
        System.out.println("  JSON: {\"vin\":\"VIN_789\",\"batteryTempC\":55.0,\"voltageV\":3.1,\"stateOfChargePercent\":82.0,\"drivingMode\":\"driving\"}");
        System.out.println("  CSV:  VIN_789,55.0,3.1,82.0,driving   (vin,tempC,voltageV[,soc%][,mode])");
        System.out.println();

        Scanner scanner = new Scanner(System.in);
        while (true) {
            System.out.print("> ");
            String input = scanner.nextLine().trim();

            if (input.isEmpty()) continue;
            if (input.equalsIgnoreCase("exit") || input.equalsIgnoreCase("quit")) {
                System.out.println("Exiting.");
                break;
            }

            BatteryTelemetry telemetry;
            try {
                telemetry = TelemetryParser.parse(input);
            } catch (IllegalArgumentException e) {
                System.out.println("Invalid input: " + e.getMessage());
                System.out.println();
                continue;
            }

            String result = agent.chat(ANALYZE_PROMPT.formatted(telemetry.toPromptString()));
            System.out.println(result);
            System.out.println();
        }

        scanner.close();
    }
}
