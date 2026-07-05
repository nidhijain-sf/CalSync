Test CalSync end-to-end to verify the app is working correctly.

## Steps

1. Check the app is running: `pgrep -fl SyncApp`
   - If not running, start it: `open http://localhost:5001` won't work — tell the user to run `./scripts/run-local.sh` first
2. Open the dashboard: `open http://localhost:5001`
3. Check the following and report status for each:
   - Is the app reachable at `http://localhost:5001`?
   - Is Salesforce shown as connected?
   - Is Google Calendar shown as connected?
   - What is the Last Sync time?
   - What is the Next Sync time?
4. Read the last 50 lines of `calsync.log` and report:
   - Any ERROR lines
   - The most recent sync summary (`[sync] summary:` line)
   - Any session expiry or re-login events
5. Verify the sync map exists and is non-empty: `cat sync_map.json | python3 -m json.tool | head -20`
6. Report a pass/fail summary:
   - PASS: app running, both accounts connected, no errors in log, sync map populated
   - FAIL: list exactly what is wrong and suggest the fix from `docs/troubleshooting.md`
