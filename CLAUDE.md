# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

**CalSync by ACE** — a Go app that syncs Salesforce "Billable Utilization" events to Google Calendar. Single binary (`SyncApp`), local web UI on port 5001, no cloud backend.

## Building

```bash
go build -o SyncApp .                         # Mac/Linux
GOOS=windows GOARCH=amd64 go build -o SyncApp.exe .  # Windows
./scripts/build.sh                            # Build both + sync templates to dist/
./scripts/run-local.sh                        # Run locally at localhost:5001
./scripts/release.sh 1.2.3                    # Tag and prepare a release
```

## Key Files

| File | Purpose |
|---|---|
| `main.go` | Entire application — HTTP server, OAuth, sync logic, scheduler |
| `templates/index.html` | Web dashboard (edit here, copy to `dist/mac/` and `dist/windows/` when releasing) |
| `dist/mac/` / `dist/windows/` | Distributable packages for end users |

## Gitignored Runtime Files

These are generated at runtime — never commit them:
- `google_credentials.json`, `sf_credentials.json` — OAuth credentials
- `google_token.json`, `last_sync.json`, `sync_map.json`, `color.json` — runtime state
- `calsync.log`, `com.ace.calsync.plist`, `SyncApp`, `SyncApp.exe`

## Slash Commands

| Command | Purpose |
|---|---|
| `/branch 1.2.3` | Create a new branch (release, feature, fix, hotfix, docs, chore) |
| `/release 1.2.3` | Build, tag, and push a new release |
| `/test` | Verify app is running and healthy |
| `/push "message"` | Safe commit and push (blocks credentials) |
| `/diff` | Human-readable diff summary |
| `/summary` | Change summary as release notes / PR / Slack |
| `/compare v0.0.3 v0.0.4` | Side-by-side version comparison |
| `/log` | Summarise calsync.log in plain English |

## Docs

Full documentation is in [`docs/`](./docs/README.md).
