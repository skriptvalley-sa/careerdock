# How to Use This Prompt with Claude Code — Quick Guide

## Setup

```bash
# 1. Create your repo
mkdir careerdock && cd careerdock
git init

# 2. Create the docs folder
mkdir -p docs

# 3. Copy the master prompt into your repo
cp career-platform-prompt.md CLAUDE-PROMPT.md

# 4. Initialize Claude Code
claude
```

## Workflow: Phase-by-Phase

### Step 1: Kick off Phase 1
Paste the entire `CLAUDE-PROMPT.md` as your first message to Claude Code. It will start with Phase 1 (Requirements Refinement) and wait for your input.

### Step 2: Iterate on each phase
- Claude Code will present analysis and recommendations.
- You discuss, push back, approve.
- Claude Code writes the finalised doc (e.g., `docs/PRD.md`) directly in your repo.
- You review the file, commit it, and say "move to Phase 2".

### Step 3: Context management across sessions
Claude Code sessions have limited context. For long projects:

```
# At the start of a new session, say:
"Read all files in docs/ to get up to speed. We are currently on Phase X."
```

Keep a `docs/STATUS.md` file tracking:
```markdown
# Project Status
- Phase 1: ✅ Complete (PRD.md)
- Phase 2: ✅ Complete (ARCHITECTURE.md)
- Phase 3: 🔄 In Progress (LLD/database.md done, LLD/api.md next)
- Phase 4-9: ⬜ Not started
```

### Step 4: Building phase
Once all design docs are finalised:
```
"Read docs/BUILD-PLAN.md. Start Sprint 0. For each task, implement it, 
test it, and commit before moving to the next task."
```

## Tips for Effective Claude Code Usage

1. **One phase at a time** — Don't rush. The discussion phase saves massive rework later.

2. **Commit after each document** — This gives you rollback points:
   ```bash
   git add docs/PRD.md && git commit -m "docs: finalise PRD"
   ```

3. **Use Claude Code's file editing** — Let it write directly to your repo. Review diffs with `git diff`.

4. **For implementation, be specific**:
   ```
   # Instead of: "build the auth system"
   # Say: "Implement the auth module per docs/LLD/api.md section 2.
   #        Start with signup and login endpoints. Write tests."
   ```

5. **Break big implementations into commits**:
   ```
   "Implement the company CRUD handlers. After each endpoint, write a test 
    and commit with a conventional commit message."
   ```

6. **Seed data strategy** — When you reach Sprint 1, have Claude Code:
   - Generate a Go script that uses Claude API to research and build company profiles
   - Store as JSON seed files in `seeds/`
   - Create a `make seed` target to load them into the DB

7. **AI prompt engineering** — For Phase 3 (AI Service Design), tell Claude Code:
   ```
   "For each AI operation (resume parsing, ATS scoring, etc.), write the 
    actual prompt template, test it with a sample resume, and iterate 
    until the output schema is reliable. Save prompts in 
    backend/internal/ai/prompts/"
   ```

8. **Local dev parity** — Insist on a working `docker-compose.yml` from Sprint 0:
   ```yaml
   # Should include: postgres, redis, minio, mailhog, meilisearch
   # One command: docker compose up -d && make dev
   ```

## Cost Estimation Prep

Before finalising the Pay Model in Phase 1, ask Claude Code to estimate:
- Claude API cost per resume parse (~input tokens for a 2-page PDF)
- Cost per ATS check (input: resume + company/job data, output: score + analysis)
- Cost per curated list generation
- Monthly infra cost for ~1000 users

This will inform your pricing decisions.

## What If You Get Stuck?

- **Design disagreement**: Ask Claude Code to present 3 options with pros/cons table.
- **Scope creep**: Reference the PRD. If it's not in the PRD, it's v2.
- **Technical uncertainty**: Ask for a quick proof-of-concept before committing to an approach.
- **Context lost**: Point Claude Code to the relevant doc file. It can read and resume.
