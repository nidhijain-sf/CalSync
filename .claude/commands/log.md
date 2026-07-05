Read and summarise the CalSync application log in plain English.

## Steps

1. Find the log file:
   - Default location: `calsync.log` in the repo root
   - If not found, check `dist/mac/calsync.log`
   - If still not found, tell the user where to find it

2. Read the last 100 lines (or all lines if `$ARGUMENTS` is `full`)

3. Summarise:
   - **Last sync**: when it ran, how many events were synced / deleted / skipped / errored
   - **Errors**: list any ERROR lines with a plain-English explanation
   - **Auth events**: any Salesforce re-logins or Google token refreshes
   - **Scheduler events**: when the next sync is scheduled

4. Flag anything that needs attention:
   - Repeated errors
   - Failed re-login attempts
   - "accounts not connected" warnings

5. End with a one-line health status:
   - ✅ All good — last sync completed successfully
   - ⚠️ Warning — [describe issue]
   - ❌ Error — [describe issue and point to troubleshooting doc]
