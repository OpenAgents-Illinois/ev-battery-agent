package com.ev.battery.agent;

import java.net.URI;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.net.http.HttpClient;
import java.util.Base64;

import dev.langchain4j.agent.tool.P;
import dev.langchain4j.agent.tool.Tool;

public class JiraTicketingTool {
    private static final String JIRA_DOMAIN = System.getenv("JIRA_DOMAIN");
    private static final String EMAIL = System.getenv("JIRA_EMAIL");
    private static final String API_TOKEN = System.getenv("JIRA_TOKEN");
    private static final String PROJECT_KEY = System.getenv("JIRA_SPACE_KEY");

    @Tool("Creates an engineering ticket in Jira when a battery defect is detected. Severity must be one of: CRITICAL, WARNING, or INFO.")
    public String fileEngineeringTicket(
        @P("Vehicle Identification Number (VIN)") String vin,
        @P("Type of defect detected") String defectType,
        @P("Brief technical explanation without special characters") String technicalReason,
        @P("Severity level: CRITICAL, WARNING, or INFO") String severity
     ) {
        if (API_TOKEN == null || PROJECT_KEY == null || JIRA_DOMAIN == null || EMAIL == null) {
            return "ERROR: Jira configuration missing. Set JIRA_TOKEN, JIRA_SPACE_KEY, JIRA_DOMAIN, and JIRA_EMAIL env vars.";
        }

        vin = sanitize(vin);
        defectType = sanitize(defectType);
        technicalReason = sanitize(technicalReason);
        severity = sanitize(severity).toUpperCase();

        String issueType = severityToIssueType(severity);
        String priority = severityToPriority(severity);

        String jsonPayload = """
        {
            "fields": {
                "project": { "key": "%s" },
                "summary": "[%s] EV Battery Alert: %s (VIN: %s)",
                "issuetype": { "name": "%s" },
                "priority": { "name": "%s" },
                "description": {
                "type": "doc",
                "version": 1,
                "content": [
                    {
                    "type": "paragraph",
                    "content": [
                        { "type": "text", "text": "Reasoning: %s" }
                    ]
                    }
                ]
                }
            }
        }
        """.formatted(PROJECT_KEY, severity, defectType, vin, issueType, priority, technicalReason);

        String auth = EMAIL + ":" + API_TOKEN;
        String encodedAuth = Base64.getEncoder().encodeToString(auth.getBytes());

        HttpClient client = HttpClient.newHttpClient();
        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create("https://" + JIRA_DOMAIN + "/rest/api/3/issue"))
                .header("Authorization", "Basic " + encodedAuth)
                .header("Content-Type", "application/json")
                .header("Accept", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(jsonPayload))
                .build();

        try {
            HttpResponse<String> response = client.send(request, HttpResponse.BodyHandlers.ofString());
            if (response.statusCode() == 201) {
                String body = response.body();
                String ticketKey = "UNKNOWN";
                if (body.contains("\"key\":\"")) {
                    ticketKey = body.split("\"key\":\"")[1].split("\"")[0];
                }
                return "SUCCESS: Ticket created with Key: " + ticketKey + " [" + issueType + " / " + priority + " priority]";
            } else {
                return "FAILED: Jira API returned " + response.statusCode() + ". Error: " + response.body();
            }
        } catch (Exception e) {
            return "ERROR: Could not connect to Jira. " + e.getMessage();
        }
    }

    private String severityToIssueType(String severity) {
        return switch (severity) {
            case "CRITICAL", "EMERGENCY" -> "Bug";
            case "WARNING" -> "Task";
            default -> "Task";
        };
    }

    private String severityToPriority(String severity) {
        return switch (severity) {
            case "CRITICAL", "EMERGENCY" -> "Highest";
            case "WARNING" -> "High";
            case "INFO" -> "Medium";
            default -> "Medium";
        };
    }

    private String sanitize(String input) {
        if (input == null) { return ""; }
        return input
            .replace("\"", "'")
            .replace("\n", " ")
            .replace("\r", " ")
            .replace("°", " degrees")
            .trim();
    }
}
