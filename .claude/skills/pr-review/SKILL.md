---
name: pr-review
description: Review GitHub pull requests by fetching diffs, analyzing code quality, checking conventions, and optionally posting review comments. Use when user says "review this PR", "check PR #N", or shares a PR URL.
usage: /pr-review [PR number or URL]
triggers:
  - "review this PR", "check PR #N", "review pull request"
  - User shares a GitHub PR URL
  - User asks to assess code quality of a PR
  - Before merging a PR
skip_for: local code review not tied to a PR, bug fixing (use issue-fix), writing code
---

# PR Review Skill

---

## Phase 1: Fetch PR Context

### Step 1.1: Extract PR Details

Run in parallel:

```bash
gh pr view <number> --json title,body,author,baseRefName,headRefName,files,additions,deletions,labels,reviewDecision
gh pr diff <number>
gh pr checks <number>
```

### Step 1.2: Parse Linked Issues

Check the PR body for issue references (`#N`, `Fixes #N`, `Closes #N`). If found:

```bash
gh issue view <N> --json title,body,labels
```

### Step 1.3: Read Project Conventions

Check for:
- `CLAUDE.md` — commit format, architectural patterns, test requirements
- `.github/` — PR templates, contributing guidelines
- Language-specific config (`Gemfile`, `package.json`, `pyproject.toml`) to understand the stack

---

## Phase 2: Analyze Changes

### Step 2.1: File-by-File Review

For each changed file, assess:

| Check | What to Look For |
|-------|-----------------|
| **Correctness** | Logic errors, off-by-one, null handling, edge cases |
| **Security** | Injection, XSS, hardcoded secrets, auth bypasses |
| **Conventions** | Naming, file placement, import ordering, project patterns |
| **Tests** | New code has tests; key paths are covered |
| **Types** | Proper annotations, no unsafe suppressions |

### Step 2.2: Cross-File Analysis

- Do changes maintain consistency across modules?
- Are imports and dependencies updated correctly?
- Does the change break existing contracts or public APIs?

### Step 2.3: Plan Alignment (if linked issue exists)

| Question | Action if Failed |
|----------|------------------|
| Does the PR address the linked issue? | Flag incomplete work |
| Are there unrelated changes? | Flag scope creep |
| Is the approach consistent with project architecture? | Note alternatives |

---

## Phase 3: Produce Review

### Severity Levels

| Level | Definition | Action Required |
|-------|------------|-----------------|
| **CRITICAL** | Breaks functionality, security vulnerability, data loss risk | Must fix before merge |
| **HIGH** | Significant bug, major deviation from plan, performance issue | Should fix before merge |
| **MEDIUM** | Code smell, minor deviation, maintainability concern | Fix in this PR or follow-up issue |
| **LOW** | Style issues, minor improvements, suggestions | Nice to have |

### Review Report Template

```markdown
## PR Review: #<number> — <title>

### Overview
- **Author:** <author>
- **Branch:** <head> → <base>
- **Files Changed:** <count> (+<additions>, -<deletions>)
- **CI Status:** <pass/fail/pending>

### What Was Done Well
- [Positive point 1]
- [Positive point 2]

### Issues Found

#### Critical (Must Fix)
[Issues that would cause bugs, security holes, or data loss]

#### High (Should Fix)
[Issues that affect correctness, performance, or maintainability]

#### Medium (Recommended)
[Code quality improvements, missing tests, convention violations]

#### Low (Suggestions)
[Style nits, optional improvements]

### Verdict
- [ ] APPROVE — Ready to merge
- [ ] COMMENT — Non-blocking feedback only
- [ ] REQUEST CHANGES — Must address issues before merge
```

---

## Phase 4: Human Approval Gate (MANDATORY)

**Never skip this phase.** Always present a summary and wait for explicit user approval before posting anything to GitHub.

### Step 4.1: Present Summary

```markdown
## Review Summary for PR #<number>

- **Verdict:** APPROVE / COMMENT / REQUEST CHANGES
- **Critical issues:** <count>
- **High issues:** <count>
- **Medium issues:** <count>
- **Low issues:** <count>

**Proposed action:** Post as GitHub review with <verdict>

Shall I:
1. Post this review to GitHub as <verdict>
2. Post with edits (tell me what to change)
3. Show findings locally only (no GitHub action)
```

### Step 4.2: Wait for Explicit Approval

- Do **NOT** post to GitHub until the user explicitly picks option 1 or 2
- If option 2, apply edits and re-present the summary
- If option 3, stop here — no GitHub interaction

### Step 4.3: Post to GitHub (only after approval)

```bash
gh pr review <number> --<approve|comment|request-changes> --body "$(cat <<'EOF'
<review body>
EOF
)"
```

---

## Quality Rules

**ALWAYS:**
- Fetch the actual diff — never review from memory
- Read `CLAUDE.md` and project config for conventions
- Lead with positives before listing issues
- Classify every issue by severity with `file:line` references
- Present the approval summary and wait for explicit confirmation before posting

**NEVER:**
- Guess about code you haven't read
- Post a review without user approval
- Nitpick style when unresolved logic issues exist
- Approve PRs with CRITICAL issues
