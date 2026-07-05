#!/bin/bash

echo "======================================"
echo "   CalSync v2 — Uninstall"
echo "======================================"
echo ""

if ! launchctl list | grep -q "com.ace.calsync"; then
  echo "ℹ️  CalSync is not currently installed."
  echo ""
  echo "You can close this window."
  exit 0
fi

echo "🗑️  Stopping and removing CalSync..."
echo ""

launchctl unload ~/Library/LaunchAgents/com.ace.calsync.plist
rm -f ~/Library/LaunchAgents/com.ace.calsync.plist

echo "✅ CalSync has been uninstalled."
echo ""
echo "   • The app will no longer start on login"
echo "   • Scheduled syncs have been stopped"
echo "   • Your calendar data and sync history are untouched"
echo ""
echo "To reinstall later, double-click Install.command."
echo ""
echo "You can close this window."
