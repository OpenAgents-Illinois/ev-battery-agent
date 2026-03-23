package com.ev.battery.agent;

import java.util.Scanner;

import io.github.cdimascio.dotenv.Dotenv;

public class App {

    private static final String INTERACTIVE_PROMPT =
        "A Rivian EV owner has reported the following battery issue. Analyze it, determine if any safety " +
        "thresholds are violated, and if so file a Jira ticket using the fileEngineeringTicket tool " +
        "with severity CRITICAL, WARNING, or INFO.\n\nReport: %s";

    /** Used by BatteryServer for structured API input. */
    static String buildPrompt(BatteryTelemetry t) {
        StringBuilder sb = new StringBuilder();
        sb.append("A Rivian ").append(t.vehicleModel).append(" has reported the following battery readings:\n");
        sb.append("- VIN: ").append(t.vin).append("\n");
        sb.append("- Battery Temperature: ").append(t.batteryTempC).append(" degrees Celsius\n");
        sb.append("- Cell Voltage: ").append(t.voltageV).append(" V\n");
        if (t.stateOfChargePercent != null)
            sb.append("- State of Charge: ").append(t.stateOfChargePercent).append("%\n");
        sb.append("- Driving Mode: ").append(t.drivingMode).append("\n");
        sb.append("\nAre any of these readings outside safe operating limits? ");
        sb.append("If yes, file a Jira ticket using the fileEngineeringTicket tool with severity CRITICAL, WARNING, or INFO.");
        return sb.toString();
    }

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
        System.out.println("Describe the battery issue in plain English, or type 'exit' to quit.");
        System.out.println("  Example: My R1S VIN ABC123 battery is running at 62 degrees, voltage 2.9V, 15% charge while driving.");
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

            // Route to the right vehicle's docs based on mention of R1S/R1T in the message
            String vehicleModel = detectModelFromText(input);
            System.out.println("Analyzing... (vehicle: " + vehicleModel + ")");

            EvExpert agent = factory.newAgent(vehicleModel);
            String result = agent.chat(INTERACTIVE_PROMPT.formatted(input));
            System.out.println(result);
            System.out.println();
        }

        scanner.close();
    }

    /** Detects R1S or R1T from free text. Defaults to R1S if not mentioned. */
    static String detectModelFromText(String text) {
        String upper = text.toUpperCase();
        if (upper.contains("R1T")) return "R1T";
        if (upper.contains("R1S")) return "R1S";
        return "R1S"; // default
    }
}
