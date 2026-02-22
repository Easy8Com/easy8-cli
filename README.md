# easy8-cli
Small Go CLI for Easy8. Current scope: Issues (tasks) only.

## Goals
- Create, list, search, and update issues.
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
List issues:

```bash
easy8 issue list --limit 10 --sort "priority:desc,due_date"
```

Search issues (fulltext):

```bash
easy8 issue search --q "onboarding"
```

Search issues with filters:

```bash
easy8 issue search --q "petr" --assignee-id 51 --status-id 2 --priority-id 3 --due-date 2024-01-10 --subject "Login" --task-type-id 1
```

Search issues with name lookups:

```bash
easy8 issue search --q "petr" --assignee "Alice Doe" --status "New" --priority "High" --task-type "Task" --project "Project A"
```

Notes:
- For assignee, status, priority, task type, and project you can use either name or ID.
- Name lookups are resolved via `/users.json`, `/issue_statuses.json`, `/enumerations/issue_priorities.json`, `/trackers.json`, `/projects.json`.

Create issue:

```bash
easy8 issue create \
  --subject "Fix onboarding" \
  --project-id 1 \
  --tracker-id 1 \
  --status-id 1 \
  --priority-id 1 \
  --author-id 1 \
  --assigned-to-id 2 \
  --description "Short summary"
```

Update issue:

```bash
easy8 issue update --id 123 --status-id 5 --done-ratio 80
```

Machine readable output:

```bash
easy8 issue list --json
```

## Roadmap
- Additional entities (projects, users, time, etc.)
- Config profiles
- Convenience commands (quick create, templates)

## Testing
- Every code change must be covered by tests; aim for line-level coverage.
- Always run:

```bash
go test ./...
```
