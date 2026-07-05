# Setup Guide

After installing CalSync, your browser will open to `http://localhost:5001`. This is the CalSync dashboard running locally on your machine.

---

## Step 1 — Connect Salesforce

Click **Connect Salesforce** and fill in your details:

| Field | What to enter |
|---|---|
| **Username** | Your Salesforce login email (e.g. `jane.doe@company.com`) |
| **Password** | Your Salesforce password |
| **Security Token** | Your Salesforce security token (see below) |
| **Domain** | Leave as `login` unless your org uses a custom domain (e.g. `mycompany`) |

### How to find your Salesforce Security Token

1. Log into Salesforce in your browser
2. Click your profile picture (top right) → **Settings**
3. In the left menu, go to **My Personal Information** → **Reset My Security Token**
4. Click **Reset Security Token**
5. Check your email — Salesforce sends the token to your registered email address

> **Note:** If your company uses IP whitelisting, you may not need a security token. Leave the field blank and try connecting.

---

## Step 2 — Connect Google Calendar

Click **Connect Google Calendar**. You will be taken to Google's login page where you can:

1. Choose or log into the Google account where you want events to appear
2. Review and approve the permissions CalSync needs (calendar read/write)
3. Click **Allow**

You will be redirected back to the CalSync dashboard automatically.

> CalSync only accesses your **primary** Google Calendar. It does not read or modify any other calendars.

---

## Step 3 — Choose an Event Color (Optional)

Use the color picker in the dashboard to choose how your synced Salesforce events appear in Google Calendar. This makes them easy to identify at a glance.

---

## Step 4 — Run Your First Sync

Click **Sync Now**. CalSync will:

1. Query Salesforce for all your Billable Utilization events from today onwards
2. Add them to your Google Calendar
3. Show a summary: how many events were created, updated, or deleted

From this point, syncs happen automatically every day at **9:00 AM**.

---

## Reconnecting Accounts

If you ever need to reconnect an account (e.g. after a password change):

1. Open `http://localhost:5001` in your browser
2. Click **Disconnect** next to the account you want to reconnect
3. Follow the connection steps again

Your sync history and previously synced events are preserved.
