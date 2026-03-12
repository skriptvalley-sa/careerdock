# Claude Code Session & Branch Management Guide

> How to run Claude Code sessions during the build phase to maintain a clean commit and PR history.

---

## 1. Key Principle

**One branch per logical task group. One session per branch. One PR per branch.**

This keeps the git history clean, PRs reviewable, and rollbacks easy.

---

## 2. Why Not Worktrees?

During the design phase, we used Claude Code's "create worktree" feature. This worked because design was a single stream of document creation. **Do not use worktrees for the build phase.** Here's why:

| | Worktree | Regular branches |
|---|---|---|
| Branch names | Random (`claude/serene-hypatia`) | Descriptive (`feature/sprint-0-auth`) |
| PR scope | Everything accumulates on one branch | One branch = one focused PR |
| Code review | Massive PRs mixing unrelated work | Focused, reviewable changes |
| Rollback | Hard to isolate what to revert | Revert one PR cleanly |
| Parallel work | Locked to one branch | Multiple branches in flight |
| Session continuity | Must resume same worktree | Checkout same branch in new session |

---

## 3. Branch Naming Convention

Per [CODE-STRUCTURE.md](../CODE-STRUCTURE.md), use trunk-based development with short-lived feature branches:

```
feature/<sprint>-<description>    # New functionality
fix/<sprint>-<description>        # Bug fixes
chore/<sprint>-<description>      # Tooling, config, docs
```

### Examples for Sprint 0

Sprint 0 has 31 tasks (~81 hours). Break it into 5-6 branches:

```
feature/sprint-0-project-scaffold     # Tasks 0.1-0.9   (repo init, Docker, CI, Makefile, config files)
feature/sprint-0-backend-foundation   # Tasks 0.10-0.13  (config, domain layer, migrations, migration runner)
feature/sprint-0-auth                 # Tasks 0.14-0.25  (repos, services, middleware, handlers, email, routes, server)
feature/sprint-0-frontend-shell       # Tasks 0.26-0.29  (app shell, auth pages, API client, auth store)
chore/sprint-0-seed-and-adrs          # Tasks 0.30-0.31  (seed script, ADRs)
```

### Sizing guideline

Each branch should be **10-20 hours of work** — large enough to be a meaningful unit, small enough to review in one sitting. If a branch exceeds ~20 files changed, consider splitting it.

---

## 4. Session Workflow

### 4.1 Starting a new session

```bash
# 1. Always start from the main repo directory (not a worktree)
cd /Users/sujaykumar/go/src/github.com/skriptvalley/careerdock

# 2. Ensure main is up to date
git checkout main && git pull

# 3. Create a new branch
git checkout -b feature/sprint-0-project-scaffold

# 4. Start Claude Code (do NOT select "create worktree")
claude
```

### 4.2 Prompting Claude at session start

Be specific about scope. Reference the design docs:

```
Starting Sprint 0 implementation. Work on tasks 0.1-0.9 from
docs/BUILD-PLAN.md — project scaffold: Go module init, Next.js init,
Docker Compose, Makefile, Air config, CI pipeline, pre-commit hooks,
.env.example, .gitignore, and PR template.

Reference docs/CODE-STRUCTURE.md for patterns and conventions.
```

### 4.3 During the session

- Claude writes code, you review
- Commit frequently (after each logical unit — e.g., after Docker Compose works, after CI pipeline is written)
- Use conventional commit messages: `feat:`, `fix:`, `chore:`, `docs:`

### 4.4 Ending a session

```
# Ask Claude to:
1. Commit all remaining changes
2. Push the branch
3. Create the PR

# Then you:
1. Review the PR on GitHub
2. Merge when satisfied
```

### 4.5 If a session runs out of context mid-branch

This will happen on larger branches. Simply:

```bash
# Start a new Claude Code session on the SAME branch (no worktree)
cd /Users/sujaykumar/go/src/github.com/skriptvalley/careerdock
git checkout feature/sprint-0-auth   # Already exists, just switch to it

claude
```

Then tell Claude:

```
Continuing work on feature/sprint-0-auth. Read docs/BUILD-PLAN.md
(Sprint 0, tasks 0.14-0.25) and the existing code in backend/internal/
to understand where we left off. Continue from the next incomplete task.
```

Claude will read the existing code and pick up where the previous session stopped.

---

## 5. PR Workflow

### 5.1 Before creating a PR

Always fetch and merge main to avoid conflicts:

```bash
git fetch origin main
git merge origin/main --no-edit
# Resolve any conflicts if they arise
git push origin <branch-name>
```

### 5.2 PR structure

Each PR should have:
- **Title:** Short, descriptive (e.g., "feat: Sprint 0 — project scaffold and CI pipeline")
- **Summary:** What tasks from BUILD-PLAN.md are covered
- **Test plan:** How to verify the changes work

### 5.3 After merge

The branch is done. Next session starts a new branch from updated main.

---

## 6. Sprint-to-Branch Mapping Template

Use this as a starting point. Adjust based on actual task dependencies and session scope.

### Sprint 0 — Foundation

| Branch | Tasks | Scope |
|--------|-------|-------|
| `feature/sprint-0-project-scaffold` | 0.1-0.9 | Go/Next.js init, Docker Compose, Makefile, CI, hooks, config files |
| `feature/sprint-0-backend-foundation` | 0.10-0.13 | Config module, domain layer, DB migrations, migration runner |
| `feature/sprint-0-auth` | 0.14-0.25 | User repo, auth service, middleware, handlers, email, routes, API server, worker skeleton |
| `feature/sprint-0-frontend-shell` | 0.26-0.29 | App layout, auth pages, API client, auth store |
| `chore/sprint-0-seed-and-adrs` | 0.30-0.31 | Seed script skeleton, initial ADRs |

### Sprint 1 — Company Directory

| Branch | Tasks | Scope |
|--------|-------|-------|
| `feature/sprint-1-company-api` | 1.1-1.6 | Company repo, service, handlers, routes, seed data, seed runner |
| `feature/sprint-1-company-frontend` | 1.7-1.16 | Company list page, profile page, components, service worker, landing page, pricing page, header/footer, cache headers |

### Sprint 2 — Lists & Tracking

| Branch | Tasks | Scope |
|--------|-------|-------|
| `feature/sprint-2-list-api` | 2.1-2.5, 2.12-2.15 | List/entry repos, services, handlers, user service, notifications, SSE |
| `feature/sprint-2-list-frontend` | 2.6-2.11 | Lists page, detail page, tracker, dashboard, layout, settings |

### Sprint 3 — Payments & Resume

| Branch | Tasks | Scope |
|--------|-------|-------|
| `feature/sprint-3-payments` | 3.1-3.8 | Payment/credit repos, payment service, Razorpay adapter, handlers, webhook, premium middleware |
| `feature/sprint-3-resume-and-ai` | 3.9-3.20 | Resume repo/service/handlers, PDF extraction, AI providers, prompts, cache, worker tasks |
| `feature/sprint-3-frontend` | 3.21-3.23 | Checkout UI, resume management, credit display |

### Sprint 4 — AI Features

| Branch | Tasks | Scope |
|--------|-------|-------|
| `feature/sprint-4-ats-and-curation` | 4.1-4.11, 4.17 | ATS repo/service/handlers, curated list repo/service/handlers, worker tasks, prompts, validation, scheduler |
| `feature/sprint-4-frontend` | 4.12-4.16 | ATS pages, curated lists page, premium dashboard, SSE integration |

### Sprint 5 — Admin & Polish

| Branch | Tasks | Scope |
|--------|-------|-------|
| `feature/sprint-5-admin-api` | 5.1-5.11 | Admin/audit/feature-flag/moderation services, handlers, worker tasks |
| `feature/sprint-5-admin-frontend` | 5.12-5.20 | All admin UI pages |
| `feature/sprint-5-security-monitoring` | 5.21-5.27 | Prometheus metrics, Sentry, Nginx headers, brute-force lockout, log redaction, feature flag seed, data export |

### Sprint 6 — Launch Prep

| Branch | Tasks | Scope |
|--------|-------|-------|
| `chore/sprint-6-infra` | 6.1-6.6 | EC2, RDS, S3, CloudFront, ECR, DNS, SSL, secrets, deploy workflow, prod Docker Compose, first deploy |
| `chore/sprint-6-monitoring-setup` | 6.7-6.11 | Grafana Cloud, UptimeRobot, Sentry projects, CloudWatch, Dependabot |
| `chore/sprint-6-launch` | 6.12-6.20 | Seed prod data, load testing, security/monitoring checklists, bug fixes, Vercel deploy, beta, README, v1.0.0 tag |

---

## 7. Prompt Templates

### Starting a new sprint branch

```
Starting Sprint {N} implementation. Work on tasks {range} from
docs/BUILD-PLAN.md — {brief description}.

Reference:
- docs/CODE-STRUCTURE.md for patterns and conventions
- docs/LLD/{relevant-doc}.md for detailed specifications
- docs/SECURITY.md for auth/security requirements (if applicable)

Check docs/STATUS.md for current progress.
```

### Continuing a session on the same branch

```
Continuing work on {branch-name}. Read docs/BUILD-PLAN.md
(Sprint {N}, tasks {range}) and the existing code to understand
where we left off. Continue from the next incomplete task.
```

### Ending a session

```
Commit all changes, push the branch, and create a PR.
Ensure there are no merge conflicts with main.
```

---

## 8. STATUS.md Updates

After each PR is merged, `docs/STATUS.md` should reflect progress. Update sprint status as:

- `⬜ Not started` — No tasks begun
- `🔄 In progress` — Some tasks complete
- `✅ Complete` — All tasks done, DoD met

Example during Sprint 0:

```markdown
## Implementation Sprints
- Sprint 0 (Foundation): 🔄 In progress (scaffold done, auth in progress)
- Sprint 1 (Company Directory): ⬜ Not started
```

---

## 9. Checklist for Each Session

- [ ] Start from `main` (or existing feature branch if continuing)
- [ ] Branch name follows convention: `feature/sprint-N-description`
- [ ] Claude prompt references BUILD-PLAN.md tasks and relevant design docs
- [ ] Commits use conventional commit format
- [ ] Before PR: fetch and merge `origin/main`
- [ ] PR created with summary and test plan
- [ ] After merge: update STATUS.md if sprint milestone reached
