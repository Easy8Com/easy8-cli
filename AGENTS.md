# easy8-cli AGENTS

This document defines how agents (human or automated) should extend the
easy8-cli project. Keep it short, pragmatic, and consistent with the API.

## Project intent
- Provide a small, fast Go CLI for Easy8.
- Current scope: Issues (tasks) only.
- Supported actions: create, list, search, update.
- Design for future entities without breaking CLI UX.

## API basics
- Base URL: configurable (default for demo use: https://demo.easysoftware.com).
- Authentication:
  - Preferred: header api key `X-Redmine-API-Key`.
  - Optional fallback: query parameter `key`.

## Endpoints in scope
- `GET /issues.json` list issues
- `POST /issues.json` create issue
- `PUT /issues/{id}.json` update issue
- `GET /search.json` fulltext search

## Create issue required fields (per swagger)
- `subject`
- `project_id`
- `tracker_id`
- `status_id`
- `priority_id`
- `author_id`
- `assigned_to_id`

## CLI command map
- `easy8 issue create`
- `easy8 issue list`
- `easy8 issue search`
- `easy8 issue update`

## Configuration
- Environment variables (recommended):
  - `EASY8_BASE_URL`
  - `EASY8_API_KEY`
  - Optional defaults: `EASY8_DEFAULT_PROJECT_ID`, `EASY8_DEFAULT_TRACKER_ID`,
    `EASY8_DEFAULT_STATUS_ID`, `EASY8_DEFAULT_PRIORITY_ID`,
    `EASY8_DEFAULT_AUTHOR_ID`, `EASY8_DEFAULT_ASSIGNED_TO_ID`
- Optional config file:
  - `~/.config/easy8/config.json`
  - Env vars override config values.

## Output format
- Human-readable table by default.
- `--json` flag for machine-readable output (skills integration).

## Error handling
- Non-2xx responses must include status code and error body in stderr.
- Exit non-zero on API errors.

## Testing
- Every code change must be covered by tests; aim for line-level coverage.
- Always run `go test ./...` after any change.

## Extension guidance
- Keep API client in a small internal package (e.g., `internal/api`).
- Add new entity commands under a dedicated package to avoid monolith.
- Do not change existing CLI flags unless there is a strong compatibility reason.
