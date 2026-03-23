---
name: claudemd-setup
description: Bootstrap or update a CLAUDE.md file that gives Claude Code persistent project instructions. Use when starting a new project, onboarding Claude to an existing repo, or updating project conventions.
usage: /claudemd-setup
triggers:
  - "set up CLAUDE.md", "create Claude instructions", "bootstrap the project"
  - "write CLAUDE.md", "configure Claude for this project", "update Claude instructions"
skip_for: writing the spec (use spec-write), writing tasks (use tasks-write)
---

# CLAUDE.md Setup Skill

## Purpose

Produce a `CLAUDE.md` that gives Claude Code persistent, project-specific instructions. This file is loaded automatically at the start of every conversation in the project directory.

A good `CLAUDE.md` eliminates repetitive prompting — you write it once and Claude follows it every session.

---

## Step 1: Gather Context

Read the following files if they exist:
- `SPEC.md` — to understand the project's architecture and tech stack
- `TASKS.md` — to understand the implementation phases
- Existing `CLAUDE.md` — if updating, read it fully before editing

If none exist, ask the user:
> "What framework and language is this project using? What are the main architectural patterns I should follow?"

---

## Step 2: Write the CLAUDE.md

Use this structure, adapting sections to the actual tech stack:

```markdown
# CLAUDE.md

## Project Overview

<1-2 sentence description of what the project is and what it does.>

Claude Code should follow the architecture described in `SPEC.md` and implement the application incrementally per `TASKS.md`.

---

# Implementation Philosophy

Claude should:

- follow idiomatic <framework> conventions
- keep controllers/handlers thin
- implement business logic in service objects
- isolate external integrations in dedicated services
- use background jobs for automation
- prefer clear, modular architecture

Do not introduce:

- unnecessary abstraction layers
- over-engineered patterns for simple problems
- runtime agent frameworks or orchestration libraries

---

# Architecture Rules

<Describe the structural rules for this specific stack. Examples below:>

**For Rails:**
Controllers must only handle requests, call services, and return responses.
Business logic must live in `app/services/`.
Background jobs must live in `app/jobs/`.

**For Node/Express:**
Routes must only parse requests and call controllers.
Business logic must live in `src/services/`.

**For Python/Django or FastAPI:**
Views/endpoints must remain thin.
Business logic must live in `app/services/` or `app/domain/`.

Adapt these rules to match the actual project stack.

---

# External Integrations

List any third-party APIs and where their wrappers live:

- `<API Name>` integration lives in `<path>`
- Credentials must never be logged or committed

---

# AI Usage (if applicable)

<Describe what Claude API is used for within the application, and what it must NOT do.>

Example:
> Claude is used only for generating user-facing text (recommendations, reports). Claude must not perform deterministic analytics — those must be computed in code.

---

# Data Safety Rules

List any sensitive data that must never be logged or exposed:

- <secret type 1> must be encrypted before storage
- <secret type 2> must never appear in logs

---

# Folder Structure

<Show the expected directory layout for this project.>

---

# Background Job System (if applicable)

The application uses <Sidekiq / Celery / BullMQ / etc.> for job processing.

Jobs should be designed to be idempotent where possible.

---

# Code Quality Rules

Claude should:

- follow idiomatic patterns for this stack
- avoid over-engineering
- use clear naming conventions
- keep files focused and small
- write tests for services and jobs where practical

---

# Incremental Development Rule

Claude must **never generate the full application in a single step**.

Claude should:

1. complete the current task from `TASKS.md`
2. ensure the application runs
3. summarize changes made
4. wait for the next instruction

---

# Claude Agents (if applicable)

Agent definitions are located in: `.claude/agents/`

These are **build-time roles**, not application code.

---

# Claude Skills

Reusable skills are defined in: `.claude/skills/`
```

---

## Step 3: Write or Update the File

- If `CLAUDE.md` does not exist, create it using the template above.
- If it exists, read it fully first. Then make targeted edits — update only the sections that are wrong or missing. Do not rewrite sections that are already correct.

---

## Quality Rules

**ALWAYS:**
- Derive rules from the actual tech stack — don't paste Rails rules into a Python project
- Include an explicit "Incremental Development Rule" section — this is critical for agentic workflows
- Keep it concise — CLAUDE.md is read every session, so signal-to-noise ratio matters
- Reference `SPEC.md` and `TASKS.md` if they exist

**NEVER:**
- Include implementation code in CLAUDE.md
- Write rules so broad they don't add information ("write good code")
- Contradict rules that are already enforced by the framework itself
- Leave out the data safety section if the project handles credentials or PII
