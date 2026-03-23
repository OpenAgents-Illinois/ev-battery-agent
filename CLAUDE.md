# EV Battery Agent — Claude Context

## Project Overview
An autonomous AI agent that monitors EV battery health and auto-files Jira tickets when safety thresholds are breached.

**Stack:** Java, LangChain4j, Google Vertex AI (Gemini 2.0 Flash + text-embedding-004), Atlassian Jira REST API, Gradle

## Architecture
- **`App.java`** — Entry point. Builds the agent: loads EV manuals from `docs/` into an in-memory vector store, wires up RAG retriever, chat memory, and the Jira tool.
- **`EvExpert.java`** — LangChain4j `AiService` interface defining the agent's chat contract.
- **`JiraTicketingTool.java`** — `@Tool`-annotated method the agent calls to POST a Jira issue via REST API (Jira Cloud, Atlassian Document Format body).

## RAG Setup
Documents in `docs/` (EV manuals, safety specs) are chunked (500 tokens, 50 overlap) and embedded at startup into an `InMemoryEmbeddingStore`. Top-5 segments are retrieved per query.

## Environment Variables (`.env`)
| Variable | Purpose |
|---|---|
| `GCLOUD_PROJECT_ID` | GCP project for Vertex AI |
| `JIRA_TOKEN` | Atlassian API token |

`JIRA_EMAIL` and `JIRA_PROJECT_KEY` are currently hardcoded in `JiraTicketingTool.java` — move to `.env` if needed.

## Build & Run
```bash
./gradlew run
```

## Key Design Decisions
- Model: `gemini-2.0-flash` with `temperature=0.0` for deterministic safety analysis
- Jira issue type: `Bug`, filed under configured project key
- Agent has a 10-message sliding window for chat memory
