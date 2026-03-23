---
name: tasks-write
description: Generate or update a TASKS.md incremental implementation plan from a project spec. Use when user says "write the tasks", "create the task plan", "break this into tasks", or "update TASKS.md".
usage: /tasks-write
triggers:
  - "write the tasks", "create the task plan", "break this into tasks"
  - "generate TASKS.md", "update the task list", "plan the implementation"
skip_for: writing the spec (use spec-write), Claude instructions (use claudemd-setup)
---

# Tasks Write Skill

## Purpose

Produce a `TASKS.md` that breaks a project spec into an ordered, incremental implementation plan. Each task must leave the application in a runnable state.

---

## Step 1: Read the Spec

If `SPEC.md` exists, read it in full before writing any tasks. The task plan must be derived from the spec, not invented.

If no spec exists, prompt the user:
> "There's no SPEC.md yet. Should I draft the spec first with /spec-write, or do you want to describe what to build directly?"

---

## Step 2: Identify Phases

Map the spec's milestones to implementation phases. Each phase groups tasks that must be completed together before the next phase can begin.

Typical phase structure (adapt to the project):

| Phase | Theme |
|-------|-------|
| 1 | Application foundation (framework setup, auth, config) |
| 2 | Core data models and database schema |
| 3 | External integrations (APIs, SDKs) |
| 4 | Business logic and analytics |
| 5 | AI/automation layer (if applicable) |
| 6 | Reporting and data outputs |
| 7 | Background jobs and scheduling |
| 8 | User interface and pages |
| 9 | Scheduling and production automation |

---

## Step 3: Write TASKS.md

Use this exact structure:

```markdown
# TASKS.md

This file defines the incremental implementation tasks for <Project Name>.

Claude should only complete **one task at a time** and should not attempt to implement future tasks unless required for scaffolding.

Each task should leave the application in a **runnable state**.

---

# Phase <N> — <Phase Name>

## Task <N.M> — <Task Name>

<One sentence description of what this task accomplishes.>

Requirements:

- <requirement 1>
- <requirement 2>

Expected outcomes:

- <outcome 1>
- <outcome 2>

---
```

Repeat for each task within each phase.

---

## Step 4: Write or Update the File

- If `TASKS.md` does not exist, create it.
- If it already exists, read it first. Then:
  - Add new tasks at the appropriate phase
  - Update tasks whose requirements have changed
  - Do not delete completed tasks — mark them `[DONE]` instead if the user asks

---

## Task Writing Rules

**Each task must:**
- Be atomic — accomplishable in one focused session
- Leave the app in a runnable, testable state when complete
- Reference specific file paths, service names, or endpoint routes from the spec
- List concrete requirements, not vague goals
- List verifiable expected outcomes

**Phases must:**
- Be ordered by dependency — later phases must not require future phase outputs
- Each phase should be independently deployable if possible

---

## Quality Rules

**ALWAYS:**
- Derive tasks from the spec — never invent features not in SPEC.md
- Use the exact service namespace, model name, and file path conventions defined in CLAUDE.md (if present)
- Keep tasks small enough to complete one per session
- Write expected outcomes as observable facts ("app boots", "endpoint returns 200", "job enqueues")

**NEVER:**
- Create tasks that depend on unimplemented future tasks
- Write vague tasks like "implement auth" — be specific ("install Devise, generate User model, add sign-in/sign-out routes")
- Duplicate work across tasks
- Write tasks that skip directly to UI before data models and services exist
