---
name: commit
description: Create well-structured git commits following Conventional Commits format. Use when user says "commit", "create a commit", or "save these changes".
usage: /commit
triggers:
  - "commit", "make a commit", "save changes", "git commit"
skip_for: creating PRs (use pr-review), fixing issues (use issue-fix)
---

# Git Commit Skill

## Conventional Commits Format

```
<type>(<scope>): <description>
```

### Types

| Type | When to Use |
|------|-------------|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `style` | Formatting, no logic change |
| `refactor` | Code restructuring, no behavior change |
| `test` | Adding or fixing tests |
| `chore` | Maintenance, dependencies, CI/CD |

---

## Process

### Step 1: Understand What Changed

Run in parallel:
```bash
git status
git diff --staged
git log --oneline -5
```

Use `git log` to confirm existing scope naming conventions in this repo before choosing a scope.

### Step 2: Draft the Commit Message

- Use **imperative mood**: "add feature" not "added feature"
- Keep the subject line under 72 characters
- If the commit resolves an open GitHub issue, append `(fixes #N)` or `(closes #N)`
- If a `CLAUDE.md` or `.github/` directory exists, check for any commit conventions defined there

### Step 3: Stage Specific Files

Prefer staging named files over `git add -A` to avoid accidentally including secrets or unrelated changes:

```bash
git add path/to/file1 path/to/file2
```

### Step 4: Create the Commit

Always pass the message via HEREDOC to preserve formatting:

```bash
git commit -m "$(cat <<'EOF'
<type>(<scope>): <description> (fixes #N if applicable)

Co-Authored-By: Claude <noreply@anthropic.com>
EOF
)"
```

### Step 5: Verify

```bash
git status
git log --oneline -3
```

---

## Attribution

When Claude creates a commit, add a co-author line in the commit body:

- `Co-Authored-By: Claude <noreply@anthropic.com>`

The human remains the commit `author`; the co-author line credits Claude.

---

## Examples

```
feat(auth): add OAuth2 login with GitHub (closes #14)
fix(api): handle null response from payment gateway (fixes #31)
refactor(jobs): extract retry logic into base job class
chore(ci): add matrix testing for Ruby 3.2 and 3.3
docs(readme): update local setup instructions
```

---

## Quality Rules

**ALWAYS:**
- Read `git diff --staged` before writing the message
- Stage specific files, not everything
- Confirm scope naming matches the repo's existing convention

**NEVER:**
- Use `--no-verify` to skip hooks
- Amend a previously pushed commit without explicit user approval
- Commit files that may contain secrets (`.env`, credentials)
