package server

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"ev-battery-agent/internal/agent"
	"ev-battery-agent/internal/telemetry"
)

func healthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	}
}

func analyzeHandler(factory *agent.Factory) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}
