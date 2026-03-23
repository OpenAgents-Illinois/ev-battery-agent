---
name: spec-write
description: Draft or update a SPEC.md product specification for any project. Use when user says "write a spec", "draft the spec", "document what we're building", or "update the spec".
usage: /spec-write
triggers:
  - "write a spec", "draft the spec", "document what we're building"
  - "create a product spec", "update the spec", "spec this out"
skip_for: implementation tasks (use tasks-write), Claude instructions (use claudemd-setup)
---

# Spec Write Skill

## Purpose

Produce a `SPEC.md` that serves as the single source of truth for what a project is, what it does, and how it is built. This document drives `TASKS.md`, `CLAUDE.md`, and all implementation decisions.

---

## Step 1: Gather Context

Before writing, ask or infer:

1. **What is the project?** — One-sentence description
2. **Who is it for?** — Target user or persona
3. **What problem does it solve?** — Core pain point
4. **What are the key features?** — Top 3–7 capabilities
5. **What is explicitly out of scope?** — Non-goals
6. **What is the tech stack?** — Languages, frameworks, databases, external APIs
7. **What are the domain models?** — Key data entities

If the user has already described the project, extract these from the conversation. Do not ask for information that is already clear.

---

## Step 2: Write the SPEC.md

Use this exact structure:

```markdown
# <Project Name>

## Overview

<2-3 sentence description of what the application does and who it is for.>

---

# Goals

The system should help users:

- <goal 1>
- <goal 2>
- <goal 3>

---

# Non Goals

The application will NOT initially support:

- <non-goal 1>
- <non-goal 2>

The focus is strictly <core focus area>.

---

# Core Features

## <Feature Name>

<Description of the feature and what it must do.>

---

## <Feature Name>

<Description.>

---

# System Architecture

<Brief description of the architectural approach — monolith, microservices, serverless, etc.>

Core components include:

- <component 1>
- <component 2>

---

# Tech Stack

## Backend
<Language and framework>

## Frontend
<Framework and libraries>

## Database
<Database>

## Background Jobs
<Queue system if applicable>

## External APIs
<Third-party integrations>

---

# Domain Models

## <ModelName>

<What this model represents.>

Fields include:

- <field 1>
- <field 2>

---

# Background Jobs (if applicable)

## <Job Name>

<What this job does.>

---

# Data Flow

1. <Step 1>
2. <Step 2>
3. <Step 3>

---

# User Interface

## <Page Name>

Displays:

- <element 1>
- <element 2>

---

# Security

The application must:

- <security requirement 1>
- <security requirement 2>

---

# Milestones

## Milestone 1
<Name and brief description>

## Milestone 2
<Name and brief description>
```

---

## Step 3: Write or Update the File

- If `SPEC.md` does not exist, create it.
- If it already exists, read it first, then produce a targeted edit — do not rewrite sections that haven't changed.
- Keep descriptions factual and implementation-neutral where possible.

---

## Quality Rules

**ALWAYS:**
- Write in plain, unambiguous language
- Separate goals from non-goals clearly
- List domain models with their key fields
- Define milestones that map to discrete deliverables

**NEVER:**
- Include implementation code in the spec
- Over-specify UI details that belong in design docs
- Leave the Non Goals section empty — it is as important as Goals
- Write a spec longer than necessary — favor clarity over completeness
