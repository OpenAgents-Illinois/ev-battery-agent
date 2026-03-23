package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"ev-battery-agent/internal/agent"
	"ev-battery-agent/internal/server"
	"ev-battery-agent/internal/tui"
)

func main() {
	if err := godotenv.Load(); err != nil {
		// .env is optional if env vars are already set
		fmt.Println("Note: .env file not found, using environment variables")
	}

	projectID := os.Getenv("GCLOUD_PROJECT_ID")
	if projectID == "" {
		log.Fatal("GCLOUD_PROJECT_ID is not set")
	}

	fmt.Println("Loading EV Battery Agent...")

	ctx := context.Background()
	factory, err := agent.NewFactory(ctx, projectID, "us-central1")
	if err != nil {
		log.Fatalf("Failed to initialize agent: %v", err)
	}

	if len(os.Args) > 1 && os.Args[1] == "--server" {
		port := 8080
		if len(os.Args) > 2 {
			if p, err := strconv.Atoi(os.Args[2]); err == nil {
				port = p
			}
		}
		server.Start(factory, port)
	} else {
		if err := tui.Start(factory); err != nil {
			log.Fatal(err)
		}
	}
}
