# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

**CalSync by ACE** — a distribution workspace for a Salesforce-to-Google Calendar sync app. This is **not a source code repository**; it contains pre-compiled binaries (`SyncApp` / `SyncApp.exe`) and the deployment scaffolding around them.

The app runs a local web server on **port 5001** that users visit to connect accounts and trigger syncs. It syncs Salesforce "Billable Utilization" events to Google Calendar daily at 9am, with catch-up on restart if the machine was off.

## Layout

```
mac/          — macOS distribution
  SyncApp             pre-compiled binary (Go or similar)
  Install.command     sets up a LaunchAgent plist, strips quarantine, codesigns
  Uninstall.command   unloads and removes the LaunchAgent
  google_credentials.json   OAuth client credentials (sensitive — do not log or expose)
  templates/index.html      single-page web UI served by SyncApp

windows/      — Windows distribution
  SyncApp.exe         pre-compiled binary
  Install.bat         registers startup via %APPDATA%\...\Startup\CalSync.bat
  Uninstall.bat       kills process and removes startup link
  google_credentials.json   same OAuth credentials as mac/
  templates/index.html      same web UI as mac/
```

## Architecture

- **SyncApp binary** reads `google_credentials.json` from its working directory and serves `templates/index.html` as its UI.
- **Mac persistence**: `~/Library/LaunchAgents/com.ace.calsync.plist` (created by Install.command). `RunAtLoad` + `KeepAlive` keep it alive across reboots.
- **Windows persistence**: a `launcher.bat` copied into the Windows Startup folder.
- **Logs**: written to `<install-dir>/calsync.log` on both platforms.
- **OAuth redirect URI**: `http://localhost:5001/connect/google/callback` (must match the Google Cloud Console config).

## Modifying the UI

The only editable source file is `templates/index.html` (identical copies in `mac/templates/` and `windows/templates/`). Changes here are picked up on the next app start — no rebuild needed. Keep both copies in sync when editing.

## Sensitive Files

`google_credentials.json` contains real OAuth client credentials (client ID + secret). Do not commit changes that expose these values, print them in logs, or embed them in the HTML template.

## Updating the Binary

To ship a new version, replace `mac/SyncApp` and `windows/SyncApp.exe` with the newly compiled builds, then re-run the installer scripts on the target machines (the installer re-codesigns on Mac).

## Checking App Status (Mac)

```bash
# Is it running?
pgrep -fl SyncApp

# LaunchAgent registered?
launchctl list | grep calsync

# Recent logs
tail -50 ~/path/to/mac/calsync.log
```

## Checking App Status (Windows)

```bat
tasklist /fi "imagename eq SyncApp.exe"
```
