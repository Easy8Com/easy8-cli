# easy8-cli

Small Go CLI for Easy8. Current scope: Issues and Product Backlog Items (PBIs).

## Goals

- Create, show, list, search, and update issues.
- List, show, and update product backlog items.
- Provide JSON output for automation and skills.
- Stay small, fast, and easy to extend.

## Requirements

- Go 1.22+
- Easy8 API key

## Install

Build locally:

```bash
go build -o easy8 ./cmd/easy8
```

With version stamp:

```bash
go build -ldflags "-X easy8-cli/internal/cli.Version=1.0.0" -o easy8 ./cmd/easy8
```

## Run

Run the compiled binary:

```bash
./easy8 issue list --limit 10
```

Or run without building:

```bash
go run ./cmd/easy8 issue list --limit 10
```

## Configuration

Environment variables:

```bash
export EASY8_BASE_URL="https://demo.easysoftware.com"
export EASY8_API_KEY="<your-key>"
```

Optional default IDs (avoid repeating on every create):

```bash
export EASY8_DEFAULT_PROJECT_ID=1
export EASY8_DEFAULT_TRACKER_ID=1
export EASY8_DEFAULT_STATUS_ID=1
export EASY8_DEFAULT_PRIORITY_ID=1
export EASY8_DEFAULT_AUTHOR_ID=1
export EASY8_DEFAULT_ASSIGNED_TO_ID=1
```

Invalid integer values produce a warning on stderr.

For local development, copy `.env.example` to `.env` and use:

```bash
source ./setup_env.sh
```

Optional config file (env overrides config):

`~/.config/easy8/config.json`

```json
{
  "base_url": "https://demo.easysoftware.com",
  "api_key": "<your-key>",
  "defaults": {
    "project_id": 1,
    "tracker_id": 1,
    "status_id": 1,
    "priority_id": 1,
    "author_id": 1,
    "assigned_to_id": 1
  }
}
```

## Usage

### Agent skill (source of truth)

This repository contains the canonical skill file for agent-driven task/PBI fetches:

```text
skills/easy8-cli/SKILL.md
```

The skill is agent-agnostic and can be used with OpenCode, Claude Code, and Codex-style workflows.

In your current local setup, copy this file into your OpenCode workspace skill path:

```bash
cp skills/easy8-cli/SKILL.md /home/petr/_projects/devel/.opencode/skills/easy8-cli/SKILL.md
```

Examples of prompts:

```text
oprav ukol #1234
oprav pbi #42
najdi pbi onboarding
```

Typical command mapping used by the skill:

```bash
easy8 issue show 1234 --json
easy8 pbi show 42 --json
easy8 pbi list --q "onboarding" --json
```

If the copied skill is not visible immediately, restart the OpenCode session so the skill index reloads.

### Show issue detail

```bash
easy8 issue show 123
easy8 issue show 123 --include journals,attachments
easy8 issue show 123 --json
easy8 issue show --id 123  # legacy compatible form
```

### List issues

```bash
easy8 issue list --limit 10 --sort "priority:desc,due_date"
```

### Search issues (fulltext)

```bash
easy8 issue search --q "onboarding"
```

### Search issues with filters

```bash
easy8 issue search --q "petr" --assignee-id 51 --status-id 2 --priority-id 3 --due-date 2024-01-10 --subject "Login" --task-type-id 1
```

### Search issues with name lookups

```bash
easy8 issue search --q "petr" --assignee "Alice Doe" --status "New" --priority "High" --task-type "Task" --project "Project A"
```

Notes:

- For assignee, status, priority, task type, and project you can use either name or ID.
- Name lookups are resolved via `/users.json`, `/issue_statuses.json`, `/enumerations/issue_priorities.json`, `/trackers.json`, `/projects.json`.

### Create issue

```bash
easy8 issue create \
  --subject "Fix onboarding" \
  --project-id 1 \
  --tracker-id 1 \
  --status-id 1 \
  --priority-id 1 \
  --author-id 1 \
  --assigned-to-id 2 \
  --description "Short summary" \
  --done-ratio 0
```

### Update issue

```bash
easy8 issue update 123 --status-id 5 --done-ratio 80
easy8 issue update --id 123 --status-id 5 --done-ratio 80  # legacy compatible form
```

`--done-ratio` must be between 0 and 100.

## Product Backlog Items (PBIs)

### List PBIs

```bash
easy8 pbi list --limit 10
easy8 pbi list --status to_do --board-id 17
easy8 pbi list --q "design" --author-id 51
```

Filters: `--status` (to_do, realization, done, deleted), `--author-id`, `--board-id`, `--q` (fulltext).

### Show PBI detail

```bash
easy8 pbi show 42
easy8 pbi show 42 --json
easy8 pbi show --id 42  # legacy compatible form
```

### Update PBI

```bash
easy8 pbi update 42 --status done
easy8 pbi update 42 --name "New name" --estimate 5 --description "Details"
easy8 pbi update --id 42 --status done  # legacy compatible form
```

Updatable fields: `--name`, `--description`, `--status`, `--estimate`.

### Version

```bash
easy8 version
```

### Machine readable output

Any command supports `--json`:

```bash
easy8 issue list --json
easy8 issue show 123 --json
easy8 pbi list --json
easy8 pbi show 42 --json
```

## Testing

Unit tests (always run after any change):

```bash
go test ./...
```

Integration tests (require a running Easy8 server):

```bash
source ./setup_env.sh && go test -tags integration -v -timeout 600s ./internal/api/
```

Integration tests use the `//go:build integration` build tag and skip
automatically when `EASY8_BASE_URL` / `EASY8_API_KEY` are not set.

## Roadmap

- Additional entities (projects, users, time, etc.)
- Config profiles
- Convenience commands (quick create, templates)
