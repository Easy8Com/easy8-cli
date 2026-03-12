---
name: easy8-cli
description: "Use Easy8 CLI to fetch and update Issue/PBI details from user requests like 'fix issue #124' or 'fix pbi #42'."
---

# Easy8 CLI

Use this skill when a user asks to work with Easy8 tasks/issues/PBIs and details should be loaded via the local `easy8` CLI.

## Scope

- Fetch issue detail
- Search/list issues
- Fetch PBI detail
- List/update PBIs
- Prepare short actionable brief for follow-up implementation work

## Invocation Hints

- Use this skill explicitly for prompts such as:
  - `fetch issue details #1234`
  - `fix issue #1234`
  - `fetch pbi details #42`
- If numeric ID is present, go directly to `show <id> --quiet`.
- If ID is not present, run search/list command first and disambiguate from top hits.

## CLI Location

Run commands in the `easy8-cli` repository root.

Prefer compiled binary if present:

- `./easy8 ...`

Fallback:

- `go run ./cmd/easy8 ...`

## Prerequisites

Required config or env:

- Run `easy8 setup` (recommended), or
- Set `EASY8_BASE_URL` and `EASY8_API_KEY`

Quick check:

```bash
go run ./cmd/easy8 --help
go run ./cmd/easy8 version
```

If auth/config is missing, stop and ask for `easy8 setup` or env setup with one focused message.

## Intent Routing

- `task`, `issue`, `ticket` -> issue commands
- `pbi`, `product backlog item`, `easy product backlog item` -> pbi commands

ID parsing:

- Accept `#124`, `124`, `issue-124`, `pbi-42`
- Extract numeric ID

## Command Mapping

### Issue detail

If user says: `fix issue #124`

```bash
go run ./cmd/easy8 issue show 124 --quiet
```

### Issue search when ID missing

```bash
go run ./cmd/easy8 issue search --q "<user text>" --quiet
```

### PBI detail

If user says: `fix pbi #42`

```bash
go run ./cmd/easy8 pbi show 42 --quiet
```

### PBI search/list when ID missing

```bash
go run ./cmd/easy8 pbi list --q "<user text>" --quiet
```

### PBI update (only when explicitly requested)

```bash
go run ./cmd/easy8 pbi update 42 --status done --quiet
```

## Output Format

- `--json` returns envelope: `ok`, `data`, `summary`, optional `breadcrumbs/context`.
- `--quiet` returns raw API-shaped JSON without envelope.
- For skill automation, prefer `--quiet` for parsing and `--json` for rich UX hints.

After fetching entity, return a short brief:

- `Entity`: Issue or PBI
- `ID`
- `Title` (`subject` for issue, `name` for pbi)
- `Status`
- `Assignee/Owner`
- `Description` (shortened)
- `Related` (linked issues for PBI when available)
- `Next Action` (what to implement or clarify)

## Safety Rules

- Always prefer machine output (`--quiet` or `--json`) for agent parsing.
- Do not update Issue/PBI unless user explicitly asks to change fields.
- For ambiguous text without ID and with multiple search hits, provide top hits and ask one focused disambiguation question.
- If entity is not found, report that directly and show the exact command used.

## Examples

- `fix issue #124` -> `issue show 124 --quiet`
- `find pbi onboarding` -> `pbi list --q "onboarding" --quiet`
- `set pbi #42 to done` -> `pbi update 42 --status done --quiet`
