package server

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/OpenAgents-Illinois/ev-battery-agent/internal/agent"
	"github.com/OpenAgents-Illinois/ev-battery-agent/internal/telemetry"
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

		severity := agent.DetectSeverity(analysis)
		writeJSON(w, http.StatusOK, analysisResponse{
			VIN:      t.VIN,
			Analysis: analysis,
			Severity: severity,
			Flag:     severityFlag(severity),
		})
	}
}

func severityFlag(severity string) string {
	switch severity {
	case agent.SeverityCritical:
		return "critical"
	case agent.SeverityWarning:
		return "warning"
	case agent.SeverityInfo:
		return "info"
	default:
		return "unknown"
	}
}
