#!/bin/bash

DIR="$(cd "$(dirname "$0")" && pwd)"

echo "======================================"
echo "   CalSync v0.0.7 — Installation"
echo "======================================"
echo ""

# Remove quarantine from all app files
xattr -dr com.apple.quarantine "$DIR" 2>/dev/null

# Make binary executable
chmod +x "$DIR/SyncApp"

# Stop and remove any previous install cleanly
launchctl unload ~/Library/LaunchAgents/com.ace.calsync.plist 2>/dev/null
rm -f ~/Library/LaunchAgents/com.ace.calsync.plist

echo "📦 Installing CalSync..."
echo ""

# Write plist with correct path for this machine
sed "s|INSTALL_DIR|$DIR|g" "$DIR/com.ace.calsync.plist" > ~/Library/LaunchAgents/com.ace.calsync.plist

# Verify the plist was written correctly
if ! grep -q "$DIR" ~/Library/LaunchAgents/com.ace.calsync.plist; then
  echo "❌ Installation failed — could not write launch configuration."
  echo "   Please contact the ACE team."
  exit 1
fi

# Load the LaunchAgent
launchctl load ~/Library/LaunchAgents/com.ace.calsync.plist

echo "✅ CalSync installed successfully!"
echo ""
echo "🔁 The app will now:"
echo "   • Start automatically every time you log in"
echo "   • Sync your calendar every Monday at 9am"
echo "   • Catch up automatically if your laptop was off"
echo ""
echo "🌐 Opening the app in your browser..."
sleep 3
open http://localhost:5001

echo ""
echo "👉 Connect your Salesforce and Google accounts in the browser."
echo "   Then click Sync Now to run your first sync."
echo ""
echo "You can close this window."
