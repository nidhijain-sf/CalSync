# CalSync — Project Overview

## What is CalSync?

CalSync is a small background app built by the ACE team that automatically copies your **Salesforce work events** into your **Google Calendar** — so you always have one calendar that shows everything.

Specifically, it syncs events of the type **"Billable Utilization"** from Salesforce into your personal Google Calendar, keeping them up to date every day.

---

## Why does it exist?

Many people at ACE track their billable work in Salesforce, but live in Google Calendar day-to-day. Without CalSync, you'd have to manually copy events between the two systems, or switch back and forth to see your full schedule.

CalSync eliminates that entirely — once set up, you never think about it again.

---

## What does it actually do?

- Runs silently in the background on your laptop (Mac or Windows)
- Every day at **9:00 AM**, it checks your Salesforce calendar for Billable Utilization events
- Any new events get added to your Google Calendar
- Any changed events (time, title, location) get updated
- Any events deleted from Salesforce get removed from Google Calendar
- If your laptop was off at 9am, it catches up automatically the next time you open it

---

## What it does NOT do

- It does not sync events from Google Calendar back to Salesforce (one-way only)
- It does not sync all Salesforce events — only the **Billable Utilization** record type
- It does not sync past events — only events from today onwards
- It does not require you to be logged into Salesforce in your browser

---

## How do I use it?

1. Install the app (see [installation guide](./installation.md))
2. Connect your Salesforce account (username + password + security token)
3. Connect your Google Calendar account (one click, browser login)
4. Click **Sync Now** once to run your first sync
5. Done — the app runs automatically from here on

---

## Where does it run?

The app runs entirely **on your own laptop**. It does not send your data to any external server other than Salesforce and Google directly. Your credentials are stored locally on your machine.

---

## Who built it?

CalSync was built by the **ACE team**. For help or feedback, use the Feedback page inside the app to reach the team on Slack.
