# Developer Guide

## Prerequisites

- [Go](https://go.dev/dl/) 1.21 or later
- Access to a Salesforce org with Billable Utilization events
- A Google Cloud project with the Calendar API enabled and OAuth credentials

## Project Structure

```
/
├── main.go                  Single-file Go application
├── go.mod                   Go module definition
├── go.sum                   Dependency checksums
├── templates/
│   ├── index.html           Web dashboard (served at localhost:5001)
│   ├── ace-logo.svg         ACE logo
│   └── ace-banner.svg       ACE banner graphic
├── dist/
│   ├── mac/                 Mac distributable (binary + installer + templates)
│   └── windows/             Windows distributable (binary + installer + templates)
├── docs/                    Project documentation
└── .gitignore               Excludes credentials, tokens, logs, binaries
```

## Building

```bash
# Mac / Linux
go build -o SyncApp .

# Windows (cross-compile from Mac/Linux)
GOOS=windows GOARCH=amd64 go build -o SyncApp.exe .
```

## Running locally

```bash
# Place google_credentials.json in the same directory first
go run main.go
# Open http://localhost:5001
```

## Google OAuth Setup

1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Create a project (or use an existing one)
3. Enable the **Google Calendar API**
4. Go to **APIs & Services → Credentials → Create Credentials → OAuth 2.0 Client ID**
5. Application type: **Web application**
6. Add `http://localhost:5001/connect/google/callback` as an Authorized Redirect URI
7. Download the credentials JSON and save it as `google_credentials.json` next to the binary

## Updating the Web UI

Edit `templates/index.html`. Changes take effect on the next app restart (no rebuild needed).

When releasing, copy the updated file to both:
- `dist/mac/templates/index.html`
- `dist/windows/templates/index.html`

## Releasing a New Version

1. Update `main.go` as needed
2. Build both binaries:
   ```bash
   go build -o dist/mac/SyncApp .
   GOOS=windows GOARCH=amd64 go build -o dist/windows/SyncApp.exe .
   ```
3. Copy updated templates to `dist/mac/templates/` and `dist/windows/templates/`
4. Test the installer on both platforms
5. Zip and distribute the `dist/mac/` and `dist/windows/` folders

## Key Code Areas in main.go

| Lines | What it does |
|---|---|
| `sfLogin()` | Salesforce SOAP authentication |
| `sfQuery()` | Salesforce REST query with auto-relogin on 401 |
| `runSync()` | Core sync logic — create/update/delete Google events |
| `startScheduler()` | Background goroutine — checks every 5 min if sync is due |
| `handleConnectGoogle()` | Initiates OAuth2 PKCE flow |
| `handleGoogleCallback()` | Receives OAuth code, exchanges for tokens |
| `withBackoff()` | Retries Google API calls on rate limit errors |
| `loadSavedCredentials()` | Restores persisted session on startup |

## Runtime Files (never commit these)

| File | Purpose |
|---|---|
| `google_credentials.json` | Google OAuth client ID + secret |
| `sf_credentials.json` | Salesforce username/password/token |
| `google_token.json` | Google OAuth access + refresh tokens |
| `sync_map.json` | SF event ID → Google event ID mapping |
| `last_sync.json` | Timestamp of last successful sync |
| `color.json` | User's chosen event color |
| `calsync.log` | Application log |
