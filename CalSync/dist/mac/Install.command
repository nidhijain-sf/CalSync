#!/bin/bash

DIR="$(cd "$(dirname "$0")" && pwd)"

echo "======================================"
echo "   CalSync — Installation"
echo "======================================"
echo ""

# Remove quarantine from entire folder recursively
xattr -dr com.apple.quarantine "$DIR" 2>/dev/null
xattr -cr "$DIR" 2>/dev/null

# Make binary executable and sign it
chmod +x "$DIR/SyncApp"
codesign --force --deep --sign - "$DIR/SyncApp" 2>/dev/null

# Stop and remove any previous install cleanly
launchctl unload ~/Library/LaunchAgents/com.ace.calsync.plist 2>/dev/null
pkill -f "SyncApp" 2>/dev/null
lsof -ti :5001 | xargs kill -9 2>/dev/null
sleep 2
rm -f ~/Library/LaunchAgents/com.ace.calsync.plist

echo "Installing CalSync..."
echo ""

cat > ~/Library/LaunchAgents/com.ace.calsync.plist << PLIST
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.ace.calsync</string>
  <key>ProgramArguments</key>
  <array>
    <string>${DIR}/SyncApp</string>
  </array>
  <key>WorkingDirectory</key>
  <string>${DIR}</string>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>${DIR}/calsync.log</string>
  <key>StandardErrorPath</key>
  <string>${DIR}/calsync.log</string>
</dict>
</plist>
PLIST

if ! grep -q "$DIR" ~/Library/LaunchAgents/com.ace.calsync.plist; then
  echo "Installation failed — could not write launch configuration."
  echo "Please contact the ACE team."
  exit 1
fi

launchctl load ~/Library/LaunchAgents/com.ace.calsync.plist

echo "CalSync installed successfully!"
echo ""
echo "The app will now:"
echo "   - Start automatically every time you log in"
echo "   - Sync your calendar every day at 9am"
echo "   - Catch up automatically if your laptop was off"
echo ""

sleep 3
APP_PID=$(pgrep -f "SyncApp" | head -1)
if ! pgrep -f "SyncApp" > /dev/null; then
  echo "The app did not start. Check the log for details:"
  echo "   ${DIR}/calsync.log"
  echo ""
  tail -20 "$DIR/calsync.log" 2>/dev/null
  echo ""
  echo "Please contact the ACE team."
else
  echo "Opening the app in your browser..."
  open http://localhost:5001
  echo ""
  echo "Connect your Salesforce and Google accounts in the browser."
  echo "Then click Sync Now to run your first sync."
fi

echo ""
echo "You can close this window."
