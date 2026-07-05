# Architecture

## High-Level Overview

CalSync is a single Go binary (`SyncApp`) that runs as a background process on the user's machine. It has no cloud backend — everything runs locally.

```mermaid
graph TB
    subgraph Laptop["User's Laptop"]
        subgraph App["SyncApp (Go binary)"]
            WS["Web Server\n:5001"]
            SCH["Scheduler\n(9am daily)"]
            SE["Sync Engine"]
            LS[("Local Storage\nsf_credentials.json\ngoogle_token.json\nsync_map.json\nlast_sync.json\ncolor.json")]
        end
        BR["Browser\nhttp://localhost:5001"]
    end

    SF["Salesforce API\n(SOAP login + REST query)"]
    GC["Google Calendar API\n(OAuth2 + REST)"]

    BR <-->|"Dashboard UI"| WS
    WS --> SE
    SCH --> SE
    SE <--> LS
    SE -->|"Query events"| SF
    SE -->|"Create / Update / Delete"| GC
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

```mermaid
sequenceDiagram
    participant U as User (Browser)
    participant A as SyncApp
    participant SF as Salesforce SOAP API

    U->>A: Enter username / password / token
    A->>SF: POST /services/Soap/u/59.0 (login)
    SF-->>A: sessionId + instanceURL
    A->>A: Save credentials to sf_credentials.json
    A-->>U: Connected ✓

    Note over A,SF: On session expiry (401)
    A->>SF: Auto re-login with saved credentials
    SF-->>A: New sessionId
    A->>SF: Retry original request
```

### Google Calendar (OAuth 2.0 + PKCE)

```mermaid
sequenceDiagram
    participant U as User (Browser)
    participant A as SyncApp
    participant G as Google OAuth

    U->>A: Click "Connect Google Calendar"
    A->>A: Generate state + PKCE verifier
    A-->>U: Redirect to Google consent screen
    U->>G: Log in and approve permissions
    G-->>U: Redirect to localhost:5001/connect/google/callback
    U->>A: Authorization code + state
    A->>G: Exchange code for tokens (with PKCE verifier)
    G-->>A: Access token + Refresh token
    A->>A: Save tokens to google_token.json
    A-->>U: Connected ✓

    Note over A,G: On token expiry
    A->>G: Auto-refresh using refresh token
    G-->>A: New access token
```

---

## Data Flow — Startup & Scheduled Sync

```mermaid
flowchart TD
    A([User logs in to laptop]) --> B[SyncApp starts automatically]
    B --> C[Load saved credentials from disk]
    C --> D{Salesforce creds saved?}
    D -->|Yes| E[Re-authenticate with Salesforce]
    D -->|No| F[Wait for user to connect]
    E --> G{Google token saved?}
    F --> G
    G -->|Yes| H[Restore Google token]
    G -->|No| F
    H --> I[Start scheduler loop\nevery 5 min]
    I --> J{Past 9am AND\nnot synced today?}
    J -->|No| I
    J -->|Yes| K[Wait 5 min for network]
    K --> L{Both accounts\nconnected?}
    L -->|No| M[Retry in 5 min]
    M --> L
    L -->|Yes| N[Run Sync Engine]
    N --> O[Save last_sync.json]
    O --> P([Desktop notification sent])
    P --> I
```

---

## Sync Engine Logic

```mermaid
flowchart TD
    A([Start Sync]) --> B[Query Salesforce\nBillable Utilization events\nfrom today onwards]
    B --> C[Load sync_map.json\nSF ID → Google event ID]
    C --> D[For each Salesforce event]

    D --> E{Already in\nsync map?}

    E -->|No| F[Create new\nGoogle Calendar event]
    F --> G[Add to sync map]

    E -->|Yes| H[Fetch existing\nGoogle Calendar event]
    H --> I{Event changed?\ntitle / time /\nlocation / color}
    I -->|Yes| J[Update Google\nCalendar event]
    I -->|No| K[Skip — no change]

    D --> L[For each previously synced event\nno longer in Salesforce]
    L --> M[Delete from\nGoogle Calendar]
    M --> N[Remove from sync map]

    G --> O[Save sync_map.json]
    J --> O
    K --> O
    N --> O
    O --> P([Sync complete])
```

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
