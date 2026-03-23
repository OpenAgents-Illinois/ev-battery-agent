---
name: build
description: Begin or continue implementing the project from TASKS.md. Reads the spec, finds the next incomplete task, implements it incrementally, and stops. Use when user says "start building", "implement the next task", "continue", or "build this".
usage: /build
triggers:
  - "start building", "build this", "implement the next task"
  - "continue", "next task", "let's build", "start implementing"
skip_for: writing the spec (use spec-write), planning tasks (use tasks-write), reviewing PRs (use pr-review)
---

# Build Skill

## Purpose

Implement the project **one task at a time**, following the plan defined in `TASKS.md`. Each task must leave the application in a runnable state before stopping.

---

## Step 1: Read Project Context

Read all three files before writing any code:

1. `SPEC.md` — understand what is being built and why
2. `CLAUDE.md` — load architecture rules, naming conventions, and constraints
3. `TASKS.md` — find the next incomplete task

If any of these files are missing, stop and tell the user:
> "`<file>` is missing. Run `/spec-write`, `/tasks-write`, or `/claudemd-setup` first."

---

## Step 2: Identify the Current Task

Scan `TASKS.md` top-to-bottom and find the **first task with no `[DONE]` marker**.

Report to the user:
```
Current task: <Phase N> — Task N.M — <Task Name>
```

If all tasks are marked `[DONE]`, report:
> "All tasks in TASKS.md are complete. Would you like to add more tasks or start a new phase?"

---

## Step 3: Plan Before Implementing

Before writing code, produce a short implementation plan:

```markdown
## Implementation Plan: Task <N.M> — <Task Name>

### Files to Create
- `path/to/file` — [purpose]

### Files to Modify
- `path/to/file` — [what changes]

### Commands to Run
- [e.g., rails generate migration, bundle install, etc.]

### Verification
- [How to confirm this task is complete — e.g., "app boots", "test passes", "endpoint returns 200"]
```

**Do not write code until this plan is presented.** If the task is straightforward, keep the plan brief — one line per file is enough.

---

## Step 4: Implement

Follow `CLAUDE.md` rules strictly:

- Keep controllers thin — business logic goes in services
- Place files in the correct directories per the architecture rules
- Use the naming conventions already established in the codebase
- Do not implement anything beyond what the current task requires
- Do not scaffold future tasks "for convenience"

If the task requires installing dependencies, run the appropriate command for the stack:

| Stack | Command |
|-------|---------|
| Ruby / Rails | `bundle add <gem>` or add to `Gemfile` then `bundle install` |
| Node | `npm install <pkg>` or `yarn add <pkg>` |
| Python | `pip install <pkg>` or add to `requirements.txt` / `pyproject.toml` |
| Go | `go get <module>` |

---

## Step 5: Verify

After implementing, verify the task is complete per its expected outcomes:

- Run the app or relevant tests to confirm nothing is broken
- Check that the specific outcomes listed in `TASKS.md` for this task are met
- Fix any errors before reporting done — do not stop on a broken state

---

## Step 6: Report and Stop

When the task is complete, report to the user:

```markdown
## Task <N.M> Complete — <Task Name>

### Files Created
- `path/to/file` — [what it does]

### Files Modified
- `path/to/file` — [what changed]

### Verification
- [What was confirmed — e.g., "app boots", "all tests pass", "endpoint returns 200"]

### Next Task
Task <N.M+1> — <Next Task Name>

Ready to continue? Say "continue" or "/build" to implement the next task.
```

**Stop here.** Do not implement the next task without explicit instruction.

---

## Quality Rules

**ALWAYS:**
- Read `SPEC.md`, `CLAUDE.md`, and `TASKS.md` before writing a single line of code
- Present an implementation plan before coding
- Implement only the current task — nothing more
- Leave the application in a runnable state when done
- Report what was built and what comes next

**NEVER:**
- Skip ahead to future tasks
- Implement features not defined in `SPEC.md`
- Violate the architecture rules in `CLAUDE.md`
- Stop on a broken or non-booting application state
- Rewrite working code from previous tasks unless the current task explicitly requires it
