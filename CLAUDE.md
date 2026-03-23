# EV Battery Agent — Claude Context

## Project Overview
An autonomous AI agent that monitors EV battery health and auto-files Jira tickets when safety thresholds are breached.

**Stack:** Java, LangChain4j, Google Vertex AI (Gemini 2.0 Flash + text-embedding-004), Atlassian Jira REST API, Javalin HTTP, Gradle

## Architecture
- **`App.java`** — Entry point. Initializes `AgentFactory`, then runs interactive mode or HTTP server based on args.
- **`AgentFactory.java`** — Initializes shared expensive resources (models, embedding store) once at startup. `newAgent()` returns a fresh `EvExpert` with clean chat memory — call per request.
- **`EvExpert.java`** — LangChain4j `AiService` interface defining the agent's chat contract and system prompt.
- **`JiraTicketingTool.java`** — `@Tool`-annotated method the agent calls to POST a Jira issue via REST API (Jira Cloud, Atlassian Document Format body). Supports severity-based priority and issue type.
- **`BatteryTelemetry.java`** — POJO for structured telemetry: `vin`, `batteryTempC`, `voltageV`, `stateOfChargePercent` (nullable), `drivingMode`, `timestamp`.
- **`TelemetryParser.java`** — Parses telemetry from JSON or CSV into `BatteryTelemetry`. Validates ranges before hitting the agent.
- **`BatteryServer.java`** — Javalin HTTP server. `POST /analyze` accepts telemetry, `GET /health` is liveness check.

## RAG Setup
Documents in `docs/` (EV manuals, safety specs) are chunked (500 tokens, 50 overlap) and embedded at startup into an `InMemoryEmbeddingStore`. Top-5 segments are retrieved per query.

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
# Interactive mode (stdin loop)
./gradlew run

# HTTP server mode on port 8080
./gradlew run --args="--server"

# HTTP server on custom port
./gradlew run --args="--server 9090"
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
- Model: `gemini-2.0-flash` with `temperature=0.0` for deterministic safety analysis
- Jira issue type: `Bug` for CRITICAL, `Task` for WARNING/INFO; priority maps to Highest/High/Medium
- `AgentFactory` is initialized once (expensive PDF embedding); `newAgent()` is cheap and called per request
- Each HTTP request gets a fresh agent with clean memory — no cross-request state leakage
- Interactive mode reuses a single agent across turns so chat memory carries over

## Workflow Instruction
**Commit after every feature or meaningful change before continuing to the next.** Use `./gradlew compileJava` to verify before committing.
