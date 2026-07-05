# How the Sync Works

This page explains what CalSync does under the hood — useful if you want to understand why events appear (or don't appear) in your calendar.

---

## What gets synced

CalSync only syncs Salesforce events that meet **all** of these conditions:

- The event's **Record Type** is exactly `Billable Utilization`
- The event's **start date is today or in the future** (past events are not synced)
- The event is **owned by you** (your Salesforce user ID)
- The event has both a start time and an end time

Everything else is ignored.

---

## Sync schedule

CalSync checks every **5 minutes** whether a daily sync is due.

A sync runs when:
- It is **past 9:00 AM** today, AND
- No sync has been completed today yet

This means:
- If your laptop is on at 9am → sync runs at 9am
- If your laptop was off at 9am and you open it at 11am → sync runs shortly after you log in
- If you already ran a manual sync today → the scheduled sync is skipped

---

## What happens during a sync

For each Salesforce event found:

```
Is this event already in Google Calendar?
├── YES → Has it changed (title, time, location, color)?
│   ├── YES → Update the Google Calendar event
│   └── NO  → Skip (no change needed)
└── NO  → Create a new Google Calendar event
```

Then, for events previously synced but no longer in Salesforce:

```
Is the Salesforce event gone?
└── YES → Delete the corresponding Google Calendar event
```

---

## The sync map

CalSync keeps a file called `sync_map.json` that tracks the link between each Salesforce event ID and its corresponding Google Calendar event ID.

This is how it knows which Google event to update or delete when a Salesforce event changes.

---

## Rate limiting

Google's Calendar API has usage limits. CalSync handles this by:

- Waiting **250ms between each API call** to avoid hitting limits
- Automatically **retrying up to 5 times** with increasing delays (1s, 2s, 4s, up to 30s) if a rate limit error is returned

---

## Automatic session recovery

If your Salesforce session expires between syncs (which happens after a period of inactivity), CalSync automatically logs back in using the saved credentials before retrying the sync. You won't see any error in most cases.

---

## Notifications

After each sync (scheduled or manual), CalSync sends a desktop notification showing:

- How many events were **created or updated**
- How many events were **deleted**
- How many **errors** occurred (if any)

---

## One-way sync

The sync is **one-way only**: Salesforce → Google Calendar.

Changes you make directly in Google Calendar (editing or deleting a synced event) will be **overwritten** the next time CalSync runs, if the original event still exists in Salesforce.
