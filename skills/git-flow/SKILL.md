---
name: git-flow
description: "Run Easy8 GitLab workflow end-to-end: ensure correct feature/fix branch, create conventional commits, and create merge request with conventional title."
---

# Git Flow (Easy8)

Use this skill for the standard Easy8 workflow:

1. start work on task branch,
2. commit with Conventional Commits,
3. create GitLab merge request with conventional title via `glab`.

## First Run (for non-developers)

### Check tools

```bash
git --version
glab --version
```

### If `glab` is missing in PATH

Tool installation is user responsibility.

If `glab` is not found, do not install it from this workflow. Ask the user to install/configure `glab`, and continue in preview mode only.

Checks:

```bash
command -v glab
```

If missing, provide help:

- Official docs: https://docs.gitlab.com/cli/

Based on current OS, provide common installation methods for the user to run:

- Homebrew (macOS/Linux): `brew install glab`
- apt (Debian/Ubuntu): `sudo apt install glab`
- pacman (Arch): `sudo pacman -S glab`

### Authenticate glab

```bash
glab auth login
glab auth status
```

If `glab auth status` fails:

1. Re-run interactive login:

```bash
glab auth login
```

2. Verify you selected the correct host/account.
3. Check token permissions in GitLab (API access required for MR operations).
4. Re-check status:

```bash
glab auth status
```

5. If still failing, continue in preview mode (generate branch/commit/MR title only) and ask user to complete auth.

### Configure glab for Easy8 GitLab (`git.easy8.com`)

Use this repository-style setup by default:

Before authentication, user must obtain a GitLab Personal Access Token in profile settings:

- https://git.easy8.com/-/user_settings/personal_access_tokens
- Required scope for MR operations: `api`

```bash
glab config set -g host git.easy8.com
glab config set -g git_protocol ssh
glab auth login --hostname git.easy8.com --git-protocol ssh --token
glab auth status
glab config get host -g
```

### Set MR defaults alias (recommended)

To always create MRs with source-branch cleanup, squash enabled, and self-assignment,
create and use a dedicated alias:

```bash
glab alias set mrc 'mr create --remove-source-branch --squash-before-merge --assignee me'
glab alias list
```

Then always prefer `glab mrc ...` instead of `glab mr create ...` in this workflow.
If a different assignee is required, still use `glab mrc` and override with explicit flags
(`--assignee`, `--reviewer`, etc.).

Notes:

- Do not copy tokens from glab config into chat or logs.
- If repository remote uses another alias (for example `git.easy.cz`), use explicit host per command:

```bash
GITLAB_HOST=git.easy.cz glab mr list
```

## Source of Truth

Follow these repository rules:

- `docs/Getting_started/Contributing.md`
- `docs/Release guides/How-to-contribute.md`
- `docs/Release guides/Branch-naming-conventions.md`
- `AGENTS.md`

If docs conflict, prefer:

1. `docs/Getting_started/Contributing.md`
2. `AGENTS.md`
3. older release-guide docs

## Canonical Naming Rules

### Branch names

- Feature branch (with reference): `feature/%task_id%_%task_subject%`
- Feature branch (without reference): `feature/%task_subject%`
- Bugfix branch (with reference): `fix/%task_id%_%task_subject%`
- Bugfix branch (without reference): `fix/%task_subject%`

`%task_subject%` rules:

- lowercase
- words joined by `_`
- ascii only
- remove unsupported characters

Examples:

- `feature/672697_move_heading_to_toggling_container_title`
- `feature/move_heading_to_toggling_container_title`
- `fix/665117_end_date_expiration`
- `fix/end_date_expiration`

### Commit message format

Use Conventional Commit format:

`<type>(<scope>): <description><optional_ref_suffix>`

Supported `type`:

- `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`

Supported reference suffix formats:

- `(refs #1234)`
- `(refs #issue-1234)`
- `(refs #pbi-1234)`

If issue/PBI is unknown, omit suffix entirely.

Example:

- `test(easy_hosting_services): cover designed content heading helper (refs #issue-1234)`

### Merge request title format

MR title uses the same style as commit message:
`<type>(<scope>): <description><optional_ref_suffix>`

MR title must not use branch naming format.

Example:

- `feat(easy_pages): move heading to toggling container title (refs #672697)`

## Required Inputs

- `work_type`: `feature` or `fix`
- `task_id`: numeric id (optional, required only when reference is used)
- `task_subject`: short task subject
- `type`: conventional commit type
- `scope`: affected module/component
- `description`: imperative short sentence
- `ref_kind`: `none` | `issue` | `pbi` | `number` (default `none` if id is not provided)
- `ref_id`: value for reference (required if `ref_kind != none`)
- `target_branch`: optional override

Optional:

- `reviewers`
- `assignee`
- `labels`
- `draft` (boolean)

MR default routing when not provided:

- `assignee`: `me`
- `reviewers`: `merge_request_pool` (ID `93`)

## Minimal Input Mode (non-developers)

If the user provides only task basics, auto-fill defaults and continue.

Minimal required from user:

- `work_type`
- `task_subject`

Auto-fill defaults:

- `type`: `feat` for `feature`, `fix` for `fix`
- `scope`: map explicitly using these rules:
  1. if `task_subject` contains `easy_pages` -> `easy_pages`
  2. if `task_subject` contains `easy_hosting_services` -> `easy_hosting_services`
  3. else use first token from `task_subject` normalized to snake_case
  4. if result is empty or invalid -> `general`
- `description`: normalized `task_subject`
- `ref_kind`: `number` if `task_id` is provided, otherwise `none`
- `ref_id`: `task_id` only when `ref_kind != none`
- `target_branch`: from target branch defaults
- `assignee`: `me`
- `reviewers`: `merge_request_pool` (ID `93`)

Before commit/MR creation, print generated values and proceed.

## Target Branch Defaults

- `feature` -> `next/minor`
- `fix` -> `next/minor`
- `next/bugs` or `master` only when explicitly requested by the user

## Execution Workflow

### 0) Decide execution mode

- If user asks for preview/dry-run, generate outputs only.
- If prerequisites fail (for example missing `glab` or invalid `glab` auth), switch to preview mode.
- Preview mode must not run `git commit`, `git push`, `glab mrc`, or `glab mr create`.

Preview output must include:

- expected branch name
- commit message
- MR title
- target branch

### 1) Validate prerequisites

- Ensure current directory is a git repo.
- Verify `glab` is installed.
- Verify `glab auth status` is valid.
- Verify alias `mrc` exists (`glab alias list`). If missing, create it and continue using `glab mrc`.
- If `glab` is missing, do not install it in this workflow; ask user to install it and continue in preview mode.
- If auth is invalid, ask user to fix auth and continue in preview mode.

### 2) Ensure working branch

- Build expected branch name from `work_type`, `task_id`, `task_subject`.
- If `task_id` exists, use `<work_type>/<task_id>_<task_subject>`.
- If `task_id` is missing, use `<work_type>/<task_subject>`.
- If current branch already matches expected, keep it.
- Otherwise create/switch to expected branch from proper base branch.
- Base branch default is `next/minor` for both `feature` and `fix`, unless user explicitly overrides target/base branch.

Command pattern:

- `git fetch origin`
- `git checkout <base_branch>`
- `git pull --ff-only`
- `git checkout -b <expected_branch>` (or `git checkout <expected_branch>` if exists)

### 3) Stage and commit

- Stage intended files only.
- Build commit message from template.
- Commit using generated message.

Template:

- `message = "<type>(<scope>): <description><suffix>"`

Suffix builder:

- `none` -> `""`
- `number` -> `" (refs #<ref_id>)"`
- `issue` -> `" (refs #issue-<ref_id>)"`
- `pbi` -> `" (refs #pbi-<ref_id>)"`

### 4) Push branch

- `git push -u origin <expected_branch>` (first push)
- no force push unless explicitly requested

### 5) Create MR title and merge request

- MR title = same format as commit message
- choose target branch by defaults unless overridden
- create MR via `glab mrc` by default

Command pattern:

- `glab mrc --title "<mr_title>" --target-branch "<target_branch>"`

No-reference example (preferred):

- `glab mrc --title "chore(easy_pages): rename heading helper for clarity" --target-branch "next/minor"`

Fallback example (only if alias is unavailable):

- `glab mr create --remove-source-branch --squash-before-merge --assignee me --title "chore(easy_pages): rename heading helper for clarity" --target-branch "next/minor"`

Optional flags:

- `--description "<text>"`
- `--reviewer "user1,user2"`
- `--assignee "user"`
- `--label "label1,label2"`
- `--draft`

Default merge options (when not provided explicitly):

- add `--remove-source-branch`
- add `--squash-before-merge`

Default routing (when not provided explicitly):

- add `--assignee "me"`
- add reviewer as Merge Request Pool (`merge_request_pool`, ID `93`)

### 6) Report result

Return:

- final branch name
- commit message used
- MR title used
- target branch
- MR URL

In preview mode, report `MR URL: not created (preview mode)`.

## Validation Guards

Before commit/MR:

- Branch must match `^(feature|fix)/([0-9]+_)?[a-z0-9_]+$`
- Commit/MR first part must match:
  `^(feat|fix|docs|style|refactor|test|chore|perf)\([a-z0-9_.-]+\): .+`
- Ref suffix (if present) must match one of:
  - `\(refs #[0-9]+\)`
  - `\(refs #issue-[0-9]+\)`
  - `\(refs #pbi-[0-9]+\)`

## Safety Rules

- Never amend commits unless user explicitly requests.
- Never use force push by default.
- Never commit secrets (`.env`, credentials, tokens).
- If there are unrelated local changes, do not revert them.
- If missing required inputs, ask one focused question with recommended default.

## Quick Examples

### Feature with numeric issue ref

- Branch: `feature/672697_move_heading_to_toggling_container_title`
- Commit: `feat(easy_pages): move heading to toggling container title (refs #672697)`
- MR title: `feat(easy_pages): move heading to toggling container title (refs #672697)`

### Test change with issue-prefixed ref

- Branch: `fix/1234_cover_heading_helper`
- Commit: `test(easy_hosting_services): cover designed content heading helper (refs #issue-1234)`
- MR title: `test(easy_hosting_services): cover designed content heading helper (refs #issue-1234)`

### No known task reference

- Commit: `chore(easy_pages): rename heading helper for clarity`
- MR title: `chore(easy_pages): rename heading helper for clarity`
