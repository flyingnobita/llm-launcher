# Project Template

Instructions for using this repo as a template and what to customize.

---

## General

| Item        | Purpose                                                       |
| ----------- | ------------------------------------------------------------- |
| `README.md` | Project overview, setup, usage                                |
| `mise.toml` | Tool versions and tasks (test, lint, format)                  |
| `AGENTS.md` | Project-specific AI instructions; extend/override agent rules |

## Git

| Item                       | Purpose                                            |
| -------------------------- | -------------------------------------------------- |
| `.github/workflows/ci.yml` | CI: prettier + markdownlint + stack-specific       |
| `.pre-commit-config.yaml`  | Pre-commit hooks (e.g., ruff, mypy)                |
| `.gitignore`               | Patterns for files to exclude from version control |
| `.gitattributes`           | Line endings (LF in repo; cross-platform support)  |

**Line endings:** Store LF in the repo. `.gitattributes` (`* text=auto eol=lf`)
normalizes this; Git converts on checkout/commit for Windows. `.editorconfig`
enforces `end_of_line = lf` in editors.

## Formatting & Linting

| Item                 | Purpose                                        |
| -------------------- | ---------------------------------------------- |
| `package.json`       | Prettier (and other) dev dependencies          |
| `.prettierrc`        | Prettier config (format first)                 |
| `.prettierignore`    | Files/dirs Prettier skips                      |
| `.markdownlint.yaml` | Markdownlint config                            |
| `.editorconfig`      | Indent, line endings, charset (cross-platform) |

**Tasks:** `mise run format` formats md, yaml, yml, json. Broaden `.gitignore`
with patterns for Node, Rust, Go, etc. as needed.

## Dev Docs

Spec-Driven Development (SDD) and documentation hierarchy:

| Type           | Index                   | Details                                                           |
| -------------- | ----------------------- | ----------------------------------------------------------------- |
| PRD (optional) | `dev-docs/PRD.md`       | Product requirements; high-level vision                           |
| Specs          | `dev-docs/SPECS.md`     | `dev-docs/YYYYMMDD-SPECS-short-title.md` (detailed behavior)      |
| Plans          | `dev-docs/PLANS.md`     | `dev-docs/YYYYMMDD-PLANS-short-title.md` (feature implementation) |
| Progress       | —                       | `dev-docs/plan-progress/*.md` (low-level details & checklists)    |
| ADRs           | `dev-docs/DECISIONS.md` | `dev-docs/adr/YYYYMMDD-short-title.md` (decisions details)        |

---

## Others

| Item                  | Purpose                               |
| --------------------- | ------------------------------------- |
| `LICENSE`             | License file (e.g., MIT, Apache 2.0)  |
| `.env.example`        | Env: API keys, tokens only            |
| `config.toml.example` | Paths, retention, log level, profiles |
| `CHANGELOG.md`        | Changelog / release notes             |

---

## Logging

| Setting      | Choice                                                                 |
| ------------ | ---------------------------------------------------------------------- |
| Location     | `logs/` folder                                                         |
| Filename     | `app.log`                                                              |
| Format       | Human-readable: `%(asctime)s - %(name)s - %(levelname)s - %(message)s` |
| Timestamp    | ISO 8601: `%Y-%m-%dT%H:%M:%S`                                          |
| Rotation     | `TimedRotatingFileHandler`, daily at midnight                          |
| Backup count | config.toml `logging.<profile>.retention_days` (default 14)            |
| Output       | File + console (both)                                                  |

---

## Optional

### Community Health (Language-Agnostic)

| File                 | Purpose                                             |
| -------------------- | --------------------------------------------------- |
| `CONTRIBUTING.md`    | How to contribute, dev setup, PR flow, commit style |
| `SECURITY.md`        | Vulnerability reporting (GitHub security policy)    |
| `CODE_OF_CONDUCT.md` | Community standards (Contributor Covenant)          |

Standard across CNCF, OpenSSF, and small-project templates.

### GitHub Templates & Automation

| Item                               | Purpose                                      |
| ---------------------------------- | -------------------------------------------- |
| `.github/ISSUE_TEMPLATE/`          | Bug report, feature request, config template |
| `.github/pull_request_template.md` | PR checklist and description                 |
| `.github/dependabot.yml`           | Dependency updates (adapt per ecosystem)     |
| `.github/FUNDING.yml`              | Optional; sponsor links                      |

### Editor Consistency (Language-Agnostic)

`.editorconfig` is included. It sets indent (2 spaces), `end_of_line = lf`, and
`charset = utf-8` for macOS, Linux, and Windows.
