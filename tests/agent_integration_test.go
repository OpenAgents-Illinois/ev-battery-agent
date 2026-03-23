//go:build integration

package tests

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/joho/godotenv"

	"ev-battery-agent/internal/agent"
)

// TestOverheatingScenario requires live GCP credentials and Jira access.
// Run with: go test -tags integration ./tests/
func TestOverheatingScenario(t *testing.T) {
	_ = godotenv.Load("../.env")

	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	if projectID == "" {
		t.Skip("GCLOUD_PROJECT_ID not set — skipping integration test")
	}

	ctx := context.Background()
	factory, err := agent.NewFactory(ctx, projectID, "us-central1")
	if err != nil {
		t.Fatalf("NewFactory: %v", err)
	}

	telemetryStr := "VIN_789, Temp: 58C, Voltage: 3.1V, Status: Driving."
	prompt := agent.InteractivePrompt("Analyze this telemetry: " + telemetryStr + ". If it's a defect, file a ticket and tell me the ticket ID.")

	response, err := factory.Chat(ctx, "UNKNOWN", prompt)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	t.Logf("Agent response: %s", response)

	lower := strings.ToLower(response)
	if !strings.Contains(lower, "jira") && !strings.Contains(lower, "batt") && !strings.Contains(lower, "ticket") {
		t.Errorf("agent response does not mention ticket or battery analysis: %s", response)
	}
}
