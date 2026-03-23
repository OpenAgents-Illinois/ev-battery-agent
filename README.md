# ev-battery-agent

An autonomous AI agent built in **Go** that monitors Rivian EV battery health, detects safety threshold violations using **Gemini 2.0 Flash** + RAG over owner manuals, and auto-files Jira tickets directly within your Atlassian Jira workspace

If a problem occurs (such as a battery thermal risk or a voltage anomaly) the agent creates a high-priority **Atlassian Jira** ticket so engineers can act before it leads to a recall.

## How it works

1. **Monitor** — Submit a battery report in plain English (TUI) or structured telemetry JSON/CSV (REST API). Readings include temperature, voltage, state of charge, and driving mode.
2. **Think** — The agent retrieves relevant chunks from Rivian R1S/R1T owner manuals via RAG (cosine similarity over Vertex AI embeddings), then calls **Gemini 2.0 Flash** to reason about whether any safety thresholds are violated.
3. **Act** — If a violation is detected, the agent calls the `fileEngineeringTicket` tool, which creates a Jira issue with severity-based priority (CRITICAL → Bug/Highest, WARNING → Task/High, INFO → Task/Medium). Duplicate tickets for the same VIN are skipped automatically.

## Stack

| Layer | Technology |
|---|---|
| Language | Go 1.22 |
| LLM | Gemini 2.0 Flash (Vertex AI) |
| Embeddings | text-embedding-004 (Vertex AI) |
| LLM framework | tmc/langchaingo |
| Jira client | andygrunwald/go-jira |
| TUI | charmbracelet/bubbletea + lipgloss |
| RAG | In-memory cosine similarity vector store |

## Setup

**1. Clone and install dependencies**
```bash
git clone <repo>
cd ev-battery-agent
go mod download
```

**2. Create a `.env` file**
```env
GCLOUD_PROJECT_ID=your-gcp-project
JIRA_DOMAIN=yourname.atlassian.net
JIRA_EMAIL=you@example.com
JIRA_TOKEN=your-atlassian-api-token
JIRA_SPACE_KEY=KAN
```

**3. Authenticate with GCP**
```bash
gcloud auth application-default login
```

## Build & Run

```bash
go build -o ev-battery-agent .

# Bubble Tea TUI (interactive mode)
./ev-battery-agent

# HTTP server on port 8080
./ev-battery-agent --server

# HTTP server on custom port
./ev-battery-agent --server 9090
```

> **First run** embeds the Rivian owner manual PDFs via Vertex AI and caches them locally (~30-60s). Every subsequent run loads from cache and starts in seconds.

## REST API

```bash
# Analyze telemetry (JSON)
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: application/json" \
  -d '{"vin":"VIN_789","batteryTempC":55.0,"voltageV":3.1,"stateOfChargePercent":82.0,"drivingMode":"driving"}'

# Analyze telemetry (CSV: vin,tempC,voltageV[,soc%][,mode])
curl -X POST http://localhost:8080/analyze \
  -H "Content-Type: text/plain" \
  -d 'VIN_789,55.0,3.1,82.0,driving'

# Health check
curl http://localhost:8080/health
```
