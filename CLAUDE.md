# EV Battery Agent — Claude Context

## Project Overview
An autonomous AI agent that monitors EV battery health and auto-files Jira tickets when safety thresholds are breached.

**Stack:** Go 1.22, tmc/langchaingo, Google Vertex AI (Gemini 2.0 Flash + text-embedding-004), andygrunwald/go-jira, Bubble Tea TUI, net/http server

## Architecture

```
main.go                        — entry point (TUI mode or --server mode)
internal/
  agent/
    factory.go   — Factory struct, NewFactory, buildStore (init/startup)
    chat.go      — Chat (RAG retrieval) + runAgentLoop (LLM tool-calling loop)
    prompts.go   — systemPrompt, InteractivePrompt, DetectModel
    store.go     — cosine similarity vector store + JSON disk cache
    docs.go      — PDF loading, text chunking, battery-keyword filtering
  jira/
    client.go    — Client struct, NewClient (go-jira wrapper + auth)
    ticket.go    — FileTicket (create Jira issue)
    search.go    — findExistingTicket (JQL deduplication)
    helpers.go   — severityToIssueType, severityToPriority, sanitize
  telemetry/
    telemetry.go — BatteryTelemetry struct, Validate, DetectModelFromVIN
    parser.go    — Parse, parseJSON, parseCSV
    prompt.go    — ToPromptString (plain-English prompt builder)
  server/
    server.go    — Start (register routes, listen)
    handlers.go  — healthHandler, analyzeHandler
    response.go  — analysisResponse, errorResponse, writeJSON
  tui/
    tui.go       — Start (Bubble Tea program entry point)
    model.go     — model struct, newModel, Init
    update.go    — Update, submit, state helpers
    view.go      — View, renderLines, renderStatus, wordWrap
    styles.go    — lipgloss style definitions
    messages.go  — agentResultMsg, agentErrMsg
```

## RAG Setup
PDFs in `docs/R1S/` and `docs/R1T/` are loaded, chunked (~300 words, 30-word overlap), and filtered to battery-relevant chunks. Embeddings are cached in `embeddings/` (git-ignored) — first run embeds via Vertex AI, subsequent runs load from disk.

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
go build -o ev-battery-agent .
./ev-battery-agent             # Bubble Tea TUI
./ev-battery-agent --server    # HTTP server on :8080
./ev-battery-agent --server 9090
```

### REST API
```bash
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving"}'

curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: text/plain" \
  -d 'VIN_789,55.0,3.1,82.0,driving'

curl http://localhost:8080/health
```

## Key Design Decisions
- Model: `gemini-2.0-flash-001` at temperature 0 for deterministic safety analysis
- Tool calling: manual agent loop (GenerateContent → handle ToolCalls → add ToolCallResponse → repeat) — max 5 iterations
- RAG: embed query → cosine similarity search → prepend top-5 chunks to user message
- Jira: go-jira library (andygrunwald/go-jira) — deduplication via JQL search before filing
- Jira severity routing: CRITICAL/EMERGENCY → Bug/Highest, WARNING → Task/High, INFO → Task/Medium
- Embedding cache: custom JSON format (`go-v1`) in `embeddings/` — git-ignored, auto-regenerated
- Vehicle routing: R1S → docs/R1S, R1T → docs/R1T (free-text detection in TUI, VIN prefix in server mode)
- TUI: Bubble Tea with AltScreen, viewport, textinput, spinner, lipgloss

## Workflow Instruction
**Commit after every feature or meaningful change before continuing to the next.** Use `go build ./...` to verify before committing.
