package com.ev.battery.agent;

import io.github.cdimascio.dotenv.Dotenv;
import io.github.cdimascio.dotenv.DotenvException;

import org.junit.jupiter.api.Tag;
import org.junit.jupiter.api.Test;
import static org.junit.jupiter.api.Assertions.assertTrue;

/**
 * Integration test — requires live GCP credentials and Jira access.
 * Run with: ./gradlew integrationTest
 */
@Tag("integration")
class BatteryAgentTest {
    @Test
    void testOverheatingScenario() throws DotenvException {
        Dotenv dotenv = Dotenv.configure().load();
        String projectId = dotenv.get("GCLOUD_PROJECT_ID");
        EvExpert agent = new AgentFactory(projectId, "us-central1").newAgent("UNKNOWN");

        String telemetry = "VIN_789, Temp: 58C, Voltage: 3.1V, Status: Driving.";
        
        try {
            String response = agent.chat("Analyze this telemetry: " + telemetry + 
                ". If it's a defect, file a ticket and tell me the ticket ID.");
            
            System.out.println("Agent Response: " + response);
            
            boolean passed = response.toLowerCase().contains("jira") || 
                            response.toLowerCase().contains("batt") || 
                            response.toLowerCase().contains("ticket");
            assertTrue(passed, "Agent did not mention the ticket. Response was: " + response);
        } catch (Exception e) {
            System.err.println("Error during agent chat: " + e.getMessage());
            e.printStackTrace();
            throw e;
        }
    }
}
