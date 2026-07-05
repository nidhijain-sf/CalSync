#!/bin/bash
# Run the app locally for development.
# Requires google_credentials.json in the repo root.

set -e

cd "$(dirname "$0")/.."

if [ ! -f "google_credentials.json" ]; then
  echo "ERROR: google_credentials.json not found in repo root."
  echo "Copy your OAuth credentials file here before running."
  exit 1
fi

echo "Starting CalSync at http://localhost:5001 ..."
go run main.go
