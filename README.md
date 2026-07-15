# CalSync by ACE

[![Build Check](https://github.com/nidhijain-sf/CalSync/actions/workflows/build-check.yml/badge.svg)](https://github.com/nidhijain-sf/CalSync/actions/workflows/build-check.yml)
[![Latest Release](https://img.shields.io/badge/release-v0.0.9-brightgreen?logo=github)](https://github.com/nidhijain-sf/CalSync/releases/latest)
[![Go Version](https://img.shields.io/badge/go-1.25-blue?logo=go)](go.mod)
[![Platform](https://img.shields.io/badge/platform-mac%20%7C%20windows-lightgrey?logo=apple)](https://github.com/nidhijain-sf/CalSync/releases/latest)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Docs](https://img.shields.io/badge/docs-read-blue?logo=readme)](./docs/README.md)

**Salesforce to Google Calendar Sync**

Syncs your Salesforce calendar events (Billable Utilization) to your Google Calendar automatically every day at 9am. If your laptop was off at 9am, it will catch up automatically the next time you open it.

---

## Installation

### Mac

> **First time only:** macOS will block the installer. Follow these exact steps.

1. Extract the zip and move the `CalSync-Share` folder to your Applications folder
2. Open the `mac` folder
3. Double-click `Install.command`
   - macOS will show a warning and block it — this is expected
4. Open **System Settings → Privacy & Security**
5. Scroll down — you will see: *"Install.command was blocked"*
6. Click **Open Anyway**
7. Enter your Mac password if prompted
8. The installer runs — your browser opens automatically
9. Connect your Salesforce account
10. Connect your Google Calendar account
11. Pick an event color *(optional)*
12. Click **Sync Now** to run your first sync

To uninstall: double-click `Uninstall.command` and follow the same Privacy & Security steps if blocked.

### Windows

1. Extract the zip and open the `windows` folder
2. Double-click `Install.bat`
   - The app installs itself and opens in your browser automatically
3. Connect your Salesforce account
4. Connect your Google Calendar account
5. Pick an event color *(optional)*
6. Click **Sync Now** to run your first sync

To uninstall: double-click `Uninstall.bat`.

---

## How It Works

- Runs a local web server on **port 5001** — visit `http://localhost:5001` to manage the app
- Written in Go (`main.go`) — handles Salesforce + Google OAuth, sync scheduling, and sync logic
- **Mac**: persists via `~/Library/LaunchAgents/com.ace.calsync.plist` (auto-starts on login)
- **Windows**: persists via a startup entry in `%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\`
- Logs written to `calsync.log` in the app's working directory

---

## Repository Structure

```
main.go               App entry point — HTTP server, OAuth flows, sync logic
go.mod / go.sum       Go module dependencies
templates/            Web UI served by the app
Install.command       Mac installer (strips quarantine, codesigns, installs LaunchAgent)
Uninstall.command     Mac uninstaller
dist/
  mac/                Distributable Mac package (installer + templates)
  windows/            Distributable Windows package (installer + templates)
```

---

## Building

```bash
go build -o SyncApp .                    # Mac/Linux
GOOS=windows go build -o SyncApp.exe .   # Windows cross-compile
```

Copy the resulting binary into `dist/mac/` or `dist/windows/` to update the distributable package.

---

## Documentation

| Doc | Description |
|---|---|
| [Overview](./docs/overview.md) | What CalSync is and why it exists |
| [Installation](./docs/installation.md) | Detailed install steps for Mac and Windows |
| [Setup](./docs/setup.md) | Connecting Salesforce and Google Calendar |
| [How the Sync Works](./docs/how-sync-works.md) | What gets synced, when, and how |
| [Architecture](./docs/architecture.md) | System design and data flow diagrams |
| [Developer Guide](./docs/developer-guide.md) | Building, running, and releasing |
| [Troubleshooting](./docs/troubleshooting.md) | Common issues and fixes |
| [Scripts & Automation](./docs/scripts-and-automation.md) | Build scripts, releases, GitHub Actions, labels |

---

## Need Help?

Use the **Feedback** page inside the app to reach the ACE team on Slack.
