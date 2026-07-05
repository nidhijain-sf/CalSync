#!/bin/bash

cd "$(dirname "$0")"

echo "======================================"
echo "  CalSync v2 — Salesforce → Google"
echo "======================================"
echo ""
echo "Starting the app..."
echo ""

# Open browser after short delay
(sleep 2 && open http://localhost:5001) &

echo "Your browser will open automatically."
echo "When you are done, press Control+C to stop."
echo ""

./SyncApp
