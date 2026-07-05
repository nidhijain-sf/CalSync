#!/bin/bash

echo "======================================"
echo "   CalSync — Uninstall"
echo "======================================"
echo ""

launchctl unload ~/Library/LaunchAgents/com.ace.calsync.plist 2>/dev/null
pkill -f "SyncApp" 2>/dev/null
rm -f ~/Library/LaunchAgents/com.ace.calsync.plist

echo "CalSync has been uninstalled."
echo ""
echo "   - The app will no longer start on login"
echo "   - Scheduled syncs have been stopped"
echo "   - Your calendar data and sync history are untouched"
echo ""
echo "To reinstall later, double-click Install.command."
echo ""
echo "You can close this window."
