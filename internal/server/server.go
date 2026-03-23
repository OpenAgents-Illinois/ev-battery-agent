package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"ev-battery-agent/internal/agent"
	"ev-battery-agent/internal/telemetry"
)

type analysisResponse struct {
	VIN      string `json:"vin"`
	Analysis string `json:"analysis"`
}

type errorResponse struct {
	Error string `json:"error"`
}

// Start launches the HTTP server on the given port.
// Endpoints:
//
//	GET  /health  — liveness check
//	POST /analyze — accepts BatteryTelemetry as JSON or CSV, returns analysis
func Start(factory *agent.Factory, port int) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	mux.HandleFunc("POST /analyze", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{"failed to read request body"})
			return
		}

		t, err := telemetry.Parse(string(body))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse{err.Error()})
			return
		}

		analysis, err := factory.Chat(context.Background(), t.VehicleModel, t.ToPromptString())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse{err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, analysisResponse{VIN: t.VIN, Analysis: analysis})
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Server listening on http://localhost%s\n", addr)
	fmt.Println("  POST /analyze  — submit battery telemetry JSON or CSV")
	fmt.Println("  GET  /health   — liveness check")

	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
