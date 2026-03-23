package server

import (
	"fmt"
	"net/http"

	"ev-battery-agent/internal/agent"
)

// Start launches the HTTP server on the given port.
// Endpoints:
//
//	GET  /health  — liveness check
//	POST /analyze — accepts BatteryTelemetry as JSON or CSV, returns analysis
func Start(factory *agent.Factory, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler())
	mux.HandleFunc("POST /analyze", analyzeHandler(factory))

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	fmt.Println("  POST /analyze  — submit battery telemetry JSON or CSV")
	fmt.Println("  GET  /health   — liveness check")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
