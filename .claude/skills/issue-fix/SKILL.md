---
name: issue-fix
description: Triage and fix GitHub issues end-to-end. Fetches the issue, explores relevant code, plans the fix, implements with TDD, and verifies. Use when user says "fix issue #N", "work on this issue", or references a GitHub issue number.
usage: /issue-fix <issue-number>
triggers:
  - "fix issue #N", "work on issue #N", "resolve issue #N"
  - User references a GitHub issue number
skip_for: known bugs with clear cause (just fix directly), PR review (use pr-review)
---

# Issue Fix Skill

---

## Phase 1: Issue Intake

### Step 1.1: Fetch Issue Details

```bash
gh issue view <number> --json title,body,labels,assignees,comments
```

### Step 1.2: Triage

Classify the issue:

| Type | Indicators | Next Action |
|------|-----------|-------------|
| **Bug report** | Error message, "not working", reproduction steps | Investigate root cause |
| **Feature request** | "Add", "implement", "support" | Plan implementation |
| **Unclear** | Vague description, no reproduction steps | Ask user for clarification before proceeding |

### Step 1.3: Create Working Branch

```bash
git checkout -b fix/<issue-number>-<short-description> main
```

---

## Phase 2: Investigate

### Step 2.1: Locate Relevant Code

Fire an `Explore` sub-agent to find:
- Files mentioned in the issue
- Code related to the error or feature area
- Existing tests for the affected area
- Recent commits that may have caused the issue (`git log --oneline -10`)

### Step 2.2: Root Cause Analysis (for bugs)

Gather:
- Issue title and body
- Affected file paths from Step 2.1
- Error messages or stack traces from the issue
- Related test failures if any

### Step 2.3: Scope Analysis (for features)

Determine:
- What files need to change
- What new files are needed
- What tests must be written
- What existing tests might break

---

## Phase 3: Plan the Fix

### Fix Plan Template

```markdown
## Fix Plan: Issue #<number>

### Summary
[One sentence: what this fix does]

### Root Cause (bugs) / Approach (features)
[Why the issue exists OR how the feature will be implemented]

### Files to Modify
- `path/to/file1` — [what changes]
- `path/to/file2` — [what changes]

### New Files
- `path/to/new_file` — [purpose]

### Tests
- `path/to/test_file` — [what it verifies]

### Success Criteria
- [ ] [Specific testable outcome]
- [ ] [Specific testable outcome]
- [ ] All existing tests still pass
```

**Present this plan to the user and wait for approval before implementing.**

---

## Phase 4: Implement (TDD)

### Step 4.1: Write Failing Tests First

Write tests that define the expected behavior:
- For bugs: a test that reproduces the bug
- For features: a test that defines the new behavior

### Step 4.2: Run Tests to Confirm Failure

Detect the project's test runner from `CLAUDE.md`, `package.json`, `Gemfile`, `pyproject.toml`, `Cargo.toml`, etc., then run:

| Stack | Command |
|-------|---------|
| Ruby / Rails | `bundle exec rspec <path>` or `rails test <path>` |
| Python | `pytest <path> -v` |
| Node / JS | `npm test` or `yarn test` |
| Go | `go test ./...` |
| Rust | `cargo test` |

Expected: **FAIL**

### Step 4.3: Implement the Fix

Make the minimal code changes needed to pass the tests. Do not refactor unrelated code.

### Step 4.4: Run Full Test Suite

```bash
# Same runner as above, but for the full suite
```

Expected: **ALL PASS** (including new tests)

---

## Phase 5: Verify & Present

### Step 5.1: Final Checklist

- [ ] New tests pass
- [ ] All existing tests pass
- [ ] No lint or type errors
- [ ] Changes are minimal and focused on the issue

### Step 5.2: Commit

Follow the `/commit` skill. Use this format:

```bash
git commit -m "$(cat <<'EOF'
fix(<scope>): resolve issue #<number> — <short description>

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

### Step 5.3: Report to User

```markdown
## Issue #<number> — Fixed

### Changes Made
- `file1:line` — [what changed]
- `file2:line` — [what changed]

### Tests Added
- `test_name` — [what it verifies]

### Verification
- All tests passing (N/N)
- No regressions detected

### Next Steps
- Push and open a PR?
- Close the issue?
```

---

## Quality Rules

**ALWAYS:**
- Read the full issue before acting
- Create a branch before making changes
- Write tests before implementation (TDD)
- Run the full test suite, not just new tests
- Present a plan and wait for user approval before implementing

**NEVER:**
- Start coding without reading the issue
- Skip the planning phase
- Commit without running tests
- Close the issue without user confirmation
- Mix fixes for multiple issues in one branch
- Use `--no-verify` to bypass hooks
