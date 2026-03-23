---
name: create-issue
description: Create well-structured GitHub issues following a consistent format. Use when user says "create an issue", "draft a task", "log a bug", or "document this feature".
usage: /create-issue
triggers:
  - "create an issue", "draft a task", "log a bug", "open an issue", "document a feature"
skip_for: fixing issues (use issue-fix), reviewing PRs (use pr-review)
---

# Create Issue Skill

## Step 1: Check for Duplicates

Before creating, search for related open issues:

```bash
gh issue list --state open --search "<keywords>"
```

If a duplicate or closely related issue exists, prefer commenting on it rather than opening a new one.

---

## Step 2: Categorize the Issue

Select exactly **one** prefix:

| Prefix | When to Use |
|--------|-------------|
| `[BUG]` | Functional errors, crashes, or unintended behavior |
| `[FEAT]` | Brand new functionality or modules |
| `[ENH]` | Improvements to existing features (performance, UX, logic) |
| `[REFACTOR]` | Code cleanup or structural changes with no functional impact |
| `[DOCS]` | README, inline comments, or documentation changes |
| `[CHORE]` | Maintenance, dependency updates, or CI/CD changes |

---

## Step 3: Draft the Issue

Use this exact Markdown structure:

```markdown
**Title:** `[PREFIX] Short, descriptive title`

**Summary:**
> 1-2 sentence high-level overview of the goal or problem.

**Details:**

- **Context:** (module, file paths, environment, or affected area)
- **Description:** (Steps to reproduce for Bugs; User stories for Features; Reasoning for Refactors)

**Reproduction (if applicable):**

1. [Step one]
2. [Step two]
3. [Observed vs expected behavior]

**Proposed Implementation:**

- [Technical approach or logic change required]
- [Files likely involved — use backticks for paths, e.g. `app/services/auth_service.rb`]

**Labels:** `bug` | `enhancement` | `documentation` | `refactor` | `chore`

---
*Co-authored by @<github-username> and @claude*
```

Replace `<github-username>` with the repository owner or the user requesting the issue.

---

## Step 4: Create the Issue

If the GitHub CLI is available, offer to create it directly:

```bash
gh issue create \
  --title "[PREFIX] Short title" \
  --body "$(cat <<'EOF'
<issue body from Step 3>
EOF
)" \
  --label "<label>"
```

Otherwise, present the formatted issue body for the user to copy into GitHub.

---

## Quality Rules

**ALWAYS:**
- Check for duplicates before creating
- Reference specific file paths in backticks (not hardcoded to any project)
- Use actionable language — describe what should happen, not just what went wrong
- Keep the title under 72 characters

**NEVER:**
- Create vague issues without a clear description or reproduction steps
- Use project-specific jargon that wouldn't make sense to a new contributor
- Skip the co-authorship attribution line
