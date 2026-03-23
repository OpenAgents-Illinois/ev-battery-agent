# EV Battery Agent — Claude Context

## Project Overview
An autonomous AI agent that monitors EV battery health and auto-files Jira tickets when safety thresholds are breached.

**Stack:** Go, tmc/langchaingo, Google Vertex AI (Gemini 2.0 Flash + text-embedding-004), Atlassian Jira REST API, Bubble Tea TUI, net/http server

> Legacy Java implementation is still present (src/, build.gradle) but the primary codebase is now Go.

## Architecture (Go)

```
cmd/ev-battery-agent/main.go   — entry point (TUI mode or --server mode)
internal/
  agent/
    factory.go   — AgentFactory: LLM init, RAG retrieval, tool-calling loop
    store.go     — in-memory vector store with cosine similarity + JSON cache
    docs.go      — PDF loading, text chunking, battery-keyword filtering
  jira/
    client.go    — Jira REST client: file ticket + deduplication JQL search
  telemetry/
    telemetry.go — BatteryTelemetry struct + JSON/CSV parser + validate
  server/
    server.go    — HTTP server: POST /analyze, GET /health
  tui/
    tui.go       — Bubble Tea TUI (viewport + textinput + spinner + lipgloss)
```

## RAG Setup
PDFs in `docs/R1S/` and `docs/R1T/` are loaded, chunked (~300 words, 30 overlap), and filtered to battery-relevant chunks (battery, thermal, voltage, etc.). Embeddings are cached in `embeddings/go_R1S.json` and `embeddings/go_R1T.json` — first run re-embeds, subsequent runs load from disk (~3s startup).

## Environment Variables (`.env`)
| Variable | Purpose |
|---|---|
| `GCLOUD_PROJECT_ID` | GCP project for Vertex AI |
| `JIRA_TOKEN` | Atlassian API token |
| `JIRA_SPACE_KEY` | Jira project key (e.g. `KAN`) |
| `JIRA_DOMAIN` | Atlassian domain (e.g. `yourname.atlassian.net`) |
| `JIRA_EMAIL` | Atlassian account email |

## Build & Run
```bash
# Build binary
go build -o ev-battery-agent ./cmd/ev-battery-agent/

# Launch Bubble Tea TUI
./ev-battery-agent

# HTTP server mode on port 8080
./ev-battery-agent --server

# HTTP server on custom port
./ev-battery-agent --server 9090
```

### REST API
```bash
# Analyze telemetry (JSON body)
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving"}'

# Analyze telemetry (CSV body)
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: text/plain" \
  -d 'VIN_789,55.0,3.1,82.0,driving'

# Health check
curl http://localhost:8080/health
```

## Key Design Decisions
- Model: `gemini-2.0-flash-001` at temperature 0 for deterministic safety analysis
- Tool calling: manual agent loop (GenerateContent → handle ToolCalls → add ToolCallResponse → repeat) — max 5 iterations
- RAG: embed query → cosine similarity search → prepend top-5 chunks to user message
- Jira deduplication: JQL search before filing — skips if open ticket for VIN exists
- Jira severity routing: CRITICAL/EMERGENCY → Bug/Highest, WARNING → Task/High, INFO → Task/Medium
- Embedding cache: custom JSON format (`go-v1`) at `embeddings/go_R1S.json` / `embeddings/go_R1T.json`
- Vehicle routing: R1S → docs/R1S, R1T → docs/R1T (detected from free text in TUI, from VIN in server mode)
- TUI: Bubble Tea with AltScreen, viewport for scrollable output, textinput, spinner, lipgloss styling

## Workflow Instruction
**Commit after every feature or meaningful change before continuing to the next.** Use `go build ./...` to verify before committing.
