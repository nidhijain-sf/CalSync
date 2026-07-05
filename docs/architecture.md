# Architecture

## High-Level Overview

CalSync is a single Go binary (`SyncApp`) that runs as a background process on the user's machine. It has no cloud backend — everything runs locally.

```
┌─────────────────────────────────────────────────────────────────┐
│                        User's Laptop                            │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    SyncApp (Go binary)                   │   │
│  │                                                          │   │
│  │   ┌─────────────┐    ┌──────────────┐   ┌────────────┐  │   │
│  │   │  Web Server │    │  Scheduler   │   │  Sync      │  │   │
│  │   │  :5001      │    │  (9am daily) │   │  Engine    │  │   │
│  │   └──────┬──────┘    └──────┬───────┘   └─────┬──────┘  │   │
│  │          │                  │                 │          │   │
│  │          └──────────────────┴─────────────────┘          │   │
│  │                             │                            │   │
│  │              ┌──────────────▼──────────────┐             │   │
│  │              │        Local Storage        │             │   │
│  │              │  sf_credentials.json        │             │   │
│  │              │  google_token.json           │             │   │
│  │              │  sync_map.json               │             │   │
│  │              │  last_sync.json              │             │   │
│  │              │  color.json                  │             │   │
│  │              └─────────────────────────────┘             │   │
│  └──────────────────────────────────────────────────────────┘   │
│                                                                  │
│  ┌────────────────┐                                             │
│  │  Browser       │ ←── http://localhost:5001                   │
│  │  (Dashboard)   │                                             │
│  └────────────────┘                                             │
└──────────────────────────────────────────────────────────────────┘
                    │                          │
                    ▼                          ▼
        ┌───────────────────┐     ┌────────────────────────┐
        │   Salesforce API  │     │  Google Calendar API   │
        │  (SOAP login +    │     │  (OAuth2 + REST)       │
        │   REST query)     │     │                        │
        └───────────────────┘     └────────────────────────┘
```

---

## Components

### Web Server (`http://localhost:5001`)

Serves a single-page HTML dashboard (`templates/index.html`) that lets users:
- Connect / disconnect Salesforce
- Connect / disconnect Google Calendar
- Choose an event color
- Trigger a manual sync
- View last sync time and next scheduled sync time

All UI state is rendered server-side — no JavaScript framework, just Go HTML templates.

### Scheduler

Runs in a background goroutine and checks every 5 minutes whether a sync is due:
- A sync is due if the current time is past 9:00 AM and no sync has happened today
- Handles wake-from-sleep and power-on after 9am automatically
- Waits 5 minutes after launch to allow network connectivity to establish

### Sync Engine

The core logic:
1. Queries Salesforce for the user's Billable Utilization events (today onwards)
2. Compares them to the sync map (local record of previously synced events)
3. Creates, updates, or deletes Google Calendar events accordingly
4. Saves the updated sync map

### Local Storage

All state is stored as JSON files in the same directory as the binary:

| File | Contents |
|---|---|
| `sf_credentials.json` | Salesforce username, password, security token, domain |
| `google_token.json` | Google OAuth access + refresh token |
| `sync_map.json` | Map of Salesforce event ID → Google Calendar event ID |
| `last_sync.json` | Timestamp of the last successful sync |
| `color.json` | User's chosen Google Calendar event color ID |
| `calsync.log` | Application log (appended on every run) |

---

## Authentication Flows

### Salesforce

CalSync uses the **Salesforce SOAP Login API** (`/services/Soap/u/59.0`). After a successful login, it stores the session ID and instance URL in memory and saves the credentials to disk for automatic re-login after reboot.

If the session expires (401 response), CalSync automatically re-authenticates using the saved credentials and retries the failed request.

### Google Calendar

CalSync uses **OAuth 2.0 with PKCE** (Proof Key for Code Exchange) for security. The flow:

1. App generates a random state + PKCE code verifier
2. User is redirected to Google's consent screen
3. Google redirects back to `http://localhost:5001/connect/google/callback`
4. App exchanges the authorization code for access + refresh tokens
5. Tokens are saved to disk; the refresh token is used automatically when the access token expires

---

## Platform Persistence

### Mac — LaunchAgent

The installer creates `~/Library/LaunchAgents/com.ace.calsync.plist` with:
- `RunAtLoad: true` — starts when the user logs in
- `KeepAlive: true` — restarts automatically if the process crashes

### Windows — Startup Folder

The installer places a `launcher.bat` in:
`%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\`

This runs the app every time the user logs into Windows.

---

## Data Flow Diagram

```
User logs in to laptop
        │
        ▼
SyncApp starts automatically
        │
        ├─► Load saved credentials from disk
        ├─► Re-authenticate with Salesforce (if credentials saved)
        ├─► Restore Google token (if saved)
        └─► Start scheduler loop (checks every 5 min)
                │
                ▼ (when past 9am and not synced today)
        ┌───────────────────────────────────────┐
        │            Sync Engine                │
        │                                       │
        │  1. Query Salesforce for events       │
        │     (Billable Utilization, today+)    │
        │                 │                     │
        │  2. Load sync_map.json                │
        │                 │                     │
        │  3. For each SF event:                │
        │     ├─ New? → Create in Google Cal    │
        │     ├─ Changed? → Update in Google Cal│
        │     └─ Unchanged? → Skip              │
        │                 │                     │
        │  4. For each previously synced event  │
        │     no longer in SF:                  │
        │     └─ Delete from Google Cal         │
        │                 │                     │
        │  5. Save updated sync_map.json        │
        │  6. Save last_sync.json               │
        └───────────────────────────────────────┘
                │
                ▼
        Desktop notification sent
        ("X events synced, Y deleted, Z errors")
```
