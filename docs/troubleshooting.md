# Troubleshooting

## The app isn't running / browser won't open

**Mac:** Check if the process is running:
```bash
pgrep -fl SyncApp
```
If nothing appears, start it manually:
```bash
open /Applications/CalSync-Share/mac/SyncApp
```
Or check the log for startup errors:
```
/Applications/CalSync-Share/mac/calsync.log
```

**Windows:** Open Task Manager and look for `SyncApp.exe` under Background Processes. If it's not there, double-click `SyncApp.exe` directly to start it manually.

---

## Dashboard won't load at http://localhost:5001

The app is not running. See above to start it. Port 5001 must be free — if another app is using it, CalSync won't start. Check the log file for a "port already in use" error.

---

## Salesforce connection fails

**"Invalid username, password, security token"**
- Double-check your username (full email address)
- Reset your security token: Salesforce → Settings → My Personal Information → Reset My Security Token
- Make sure you're appending the token to your password in the token field (not the password field)

**"Login failed" with no other message**
- Your company may use a custom Salesforce domain. Enter your org's domain name in the Domain field (e.g. `mycompany` for `mycompany.salesforce.com`)

**Connection succeeds but no events appear after sync**
- Confirm you have events with the **Billable Utilization** record type in Salesforce
- Check that the events are scheduled for **today or in the future** (past events are not synced)

---

## Google connection fails

**"redirect_uri_mismatch"**
- This is a configuration issue. Contact the ACE team — the Google OAuth app may need to be reconfigured.

**"Access blocked" or consent screen issues**
- Make sure you're logging into the correct Google account
- Try opening `http://localhost:5001` in an incognito window

---

## Events are not appearing in Google Calendar

1. Check the log file for errors during the last sync
2. Confirm the sync actually ran — check **Last Sync** on the dashboard
3. Click **Sync Now** to trigger a manual sync and watch for errors
4. Make sure both Salesforce and Google are shown as connected (green) on the dashboard

---

## Events appeared in Google Calendar but then disappeared

CalSync deleted them because they were removed from Salesforce or no longer match the Billable Utilization record type. If this is unexpected, check the event in Salesforce.

---

## Edits I made in Google Calendar got overwritten

CalSync syncs one-way from Salesforce to Google. Any manual changes to synced events in Google Calendar will be reverted on the next sync if the original Salesforce event still exists. Edit the event in Salesforce instead.

---

## The app syncs once but not daily

The scheduler checks every 5 minutes. Make sure:
- The app is running in the background (it should start automatically at login)
- Your laptop was on or woke up after 9am

**Mac:** Check the LaunchAgent is registered:
```bash
launchctl list | grep calsync
```
If it's missing, re-run `Install.command`.

**Windows:** Check the Startup folder:
```
%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup\
```
`CalSync.bat` should be present. If missing, re-run `Install.bat`.

---

## How to read the log file

The log is at `calsync.log` in the same folder as `SyncApp`.

Key log prefixes:

| Prefix | Meaning |
|---|---|
| `[startup]` | App started, loading saved credentials |
| `[scheduler]` | Scheduled sync activity |
| `[sync]` | Individual event sync actions |
| `[sf]` | Salesforce API calls |
| `[colors]` | Google Calendar color loading |
| `[creds]` | Credential save/load activity |
| `[handler]` | Web dashboard requests |

---

## Still stuck?

Use the **Feedback** page inside the app (`http://localhost:5001`) to reach the ACE team on Slack.
