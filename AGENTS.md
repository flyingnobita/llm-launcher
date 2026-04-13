# AGENTS.md

> **This is a template repository.** Before using it for development, complete
> the setup steps below. An AI agent or user should follow this guide to adapt
> the template for a new project.

---

## Setup: Adapt This Template

When helping a user set up a new project from this template, run through the
following steps. Ask the user the questions; then apply their answers.

---

### Step 1: Project Identity

Ask the user:

1. **Project name** — Display name for the project (e.g. "My Awesome API")
2. **Folder/repo name** — Directory and repo name (e.g. `my-awesome-api`)
3. **Brief description** — One sentence describing the project

**Action:** Rename the project folder from `project-template` to the chosen
folder name (e.g. `mv project-template my-awesome-api`). Update:

- `README.md` — Replace "Project Name" and "Brief description"
- `LICENSE` — Replace `<year>` and `<copyright holders>` if applicable

---

### Step 2: Tech Stack

Ask the user:

1. **Runtime/language** — Python, Node, Rust, Go, or other?
2. **Version** — Specific version (e.g. Python 3.12, Node 20 LTS)
3. **Framework** — If any (e.g. FastAPI, Next.js, Actix)
4. **Database** — If any (e.g. PostgreSQL, SQLite)
5. **Package manager** — uv, npm/pnpm, cargo, go mod, etc.

**Action:** Update `dev-docs/SPECS.md` Tech Stack table. Update `mise.toml` with
`[tools]` and `[tasks]` for the chosen stack.

---

### Step 3: Config & Environment

Ask the user:

1. **Environment variables** — What API keys, tokens, or secrets are needed?
2. **Config needs** — Paths, log levels, or other config beyond env vars?

**Action:** Update `.env.example` with actual variable names (remove
placeholders). Adjust `config.toml.example` if the project needs different
config (or remove if not needed). Update `dev-docs/SPECS.md` Config section.

---

### Step 4: CI & Pre-commit

**Action:** Replace the placeholder `check` job in `.github/workflows/ci.yml`
with stack-specific setup and commands (test, lint). Uncomment or add
stack-specific hooks in `.pre-commit-config.yaml` (e.g. ruff, eslint, cargo
fmt).

---

### Step 5: Source Layout

Ask the user:

1. **Source layout** — Does the project use `src/`, `cmd/`, `app/`, or something
   else? (See `dev-docs/SPECS.md` Project Structure for examples.)

**Action:** Create the source directory structure. Update `dev-docs/SPECS.md` if
the example structure does not fit.

---

### Step 6: Final Checks

- [ ] Run `npm install` (or `pnpm install`) to install Prettier and other deps
- [ ] Run `mise run format` — formats md, yaml, yml, json
- [ ] Run `mise run check` (or equivalent) to verify setup
- [ ] Remove or update this AGENTS.md section if project-specific agent rules
      are needed

---

## After Setup

Once setup is complete, replace this setup guide with project-specific AI
instructions (e.g. coding standards, stack conventions, where to find key
modules). See `TEMPLATE.md` for the full list of customizable items.
