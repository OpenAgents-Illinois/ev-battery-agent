---
name: create-pr
description: Create a GitHub pull request from the current branch or fork. Use when user says "create a PR", "open a PR", "make a pull request", or "PR these changes".
usage: /create-pr [optional target repo]
triggers:
  - "create a PR", "open a PR", "make a pull request", "submit a PR"
  - "PR these changes", "push and PR"
skip_for: reviewing PRs (use pr-review), committing changes (use commit)
---

# Create PR Skill

---

## Phase 1: Understand the Context

Run in parallel:

```bash
git status
git remote -v
git log --oneline -10
```

Then determine:
- **Current branch** — is it `master`/`main` or a feature branch?
- **Remotes** — is there an `upstream` (fork scenario) or just `origin`?
- **Ahead/behind** — how many commits are ahead of the target base?

```bash
git log upstream/master..origin/master --oneline   # if upstream exists
git log origin/master..HEAD --oneline              # if no upstream
```

---

## Phase 2: Determine PR Parameters

### Head and Base

| Scenario | Head | Base |
|----------|------|------|
| Feature branch → same repo | `current-branch` | `master` or `main` |
| Fork → upstream repo | `fork-owner:master` | `master` |

To find the upstream org/repo:
```bash
git remote get-url upstream
```

To find the fork owner:
```bash
gh repo view --json nameWithOwner -q .nameWithOwner
```

### Commits in this PR

```bash
git log <base>..<head> --oneline
```

Read the actual diff to understand what changed:

```bash
git diff <base>..<head> --stat
```

---

## Phase 3: Draft the PR

### Title
- Follow Conventional Commits format: `<type>(<scope>): <description>`
- Keep under 70 characters
- If a single commit, use that commit's subject line

### Body Template

```markdown
## Summary

- [bullet summarising change 1]
- [bullet summarising change 2]
- [bullet summarising change 3]

## Test plan
- [ ] CI checks pass
- [ ] [Any manual verification steps relevant to the changes]

🤖 Generated with [Claude Code](https://claude.com/claude-code)
```

If the PR closes a GitHub issue, add `Closes #N` on its own line in the summary.

---

## Phase 4: Push and Create

### Push if needed

If the branch hasn't been pushed yet:

```bash
git push -u origin <branch>
```

### Create the PR

```bash
gh pr create \
  --repo <target-repo> \
  --head <head> \
  --base <base> \
  --title "<title>" \
  --body "$(cat <<'EOF'
<body>
EOF
)"
```

Then output the PR URL.

---

## Quality Rules

**ALWAYS:**
- Read the actual commits and diff before writing the title/body
- Detect fork vs. same-repo scenario and set `--head` and `--repo` correctly
- Push the branch before creating the PR if it hasn't been pushed

**NEVER:**
- Open a PR with no commits ahead of base
- Use `git push --force` without explicit user approval
- Include secrets or credentials in the PR body
