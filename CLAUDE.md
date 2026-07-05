# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

**CalSync by ACE** — a Salesforce-to-Google Calendar sync app written in Go. It runs a local web server on **port 5001** that users visit to connect accounts and trigger syncs. It syncs Salesforce "Billable Utilization" events to Google Calendar daily at 9am, with catch-up on restart if the machine was off.

## Source Layout

```
CalSync/
  main.go               Go application entry point — HTTP server, sync logic, OAuth flows
  go.mod / go.sum       Go module dependencies
  templates/index.html  single-page web UI served by the app
  Install.command       Mac: strips quarantine, codesigns binary, installs LaunchAgent
  Uninstall.command     Mac: removes LaunchAgent
  dist/
    mac/                distributable Mac package (binary + installer + templates)
    windows/            distributable Windows package (binary + installer + templates)
```

## Architecture

- **main.go** serves `templates/index.html` and handles Salesforce + Google OAuth, sync scheduling, and the sync logic itself.
- **Mac persistence**: `~/Library/LaunchAgents/com.ace.calsync.plist` (`RunAtLoad` + `KeepAlive`). Created by `Install.command`.
- **Windows persistence**: `launcher.bat` copied into `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\`.
- **Logs**: written to `calsync.log` in the app's working directory.
- **OAuth redirect URI**: `http://localhost:5001/connect/google/callback` — must match Google Cloud Console config.

## Building

```bash
cd CalSync
go build -o SyncApp .          # Mac/Linux
GOOS=windows go build -o SyncApp.exe .   # Windows cross-compile
```

Copy the resulting binary into `dist/mac/` or `dist/windows/` to update the distributable package.

## Modifying the UI

Edit `CalSync/templates/index.html`. Changes are picked up on the next app restart. When updating `dist/`, copy the updated file into both `dist/mac/templates/` and `dist/windows/templates/`.

## Runtime Files (gitignored)

These are generated at runtime and excluded from version control:
- `google_credentials.json` / `sf_credentials.json` — OAuth credentials (sensitive)
- `google_token.json` — stored user token
- `last_sync.json` / `sync_map.json` — sync state
- `calsync.log` — log output
- `com.ace.calsync.plist` — generated LaunchAgent config

## Checking App Status (Mac)

```bash
pgrep -fl SyncApp
launchctl list | grep calsync
tail -50 CalSync/calsync.log
```

## Checking App Status (Windows)

```bat
tasklist /fi "imagename eq SyncApp.exe"
```
