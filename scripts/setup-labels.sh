#!/bin/bash
# Create GitHub labels for the CalSync repo.
# Requires: gh CLI authenticated (run `gh auth login` first)

REPO="nidhijain-sf/CalSync"

create_label() {
  local name="$1"
  local color="$2"
  local desc="$3"
  gh label create "$name" --color "$color" --description "$desc" --repo "$REPO" --force
}

echo "Creating labels for $REPO..."

create_label "bug"              "d73a4a" "Something isn't working"
create_label "enhancement"      "a2eeef" "New feature or improvement"
create_label "documentation"    "0075ca" "Improvements or additions to docs"
create_label "needs-triage"     "e4e669" "Needs investigation or prioritisation"
create_label "good first issue" "7057ff" "Good for newcomers"
create_label "help wanted"      "008672" "Extra attention needed"
create_label "wontfix"          "ffffff" "This will not be worked on"
create_label "duplicate"        "cfd3d7" "This issue or PR already exists"
create_label "release"          "0e8a16" "Release related"
create_label "mac"              "1d76db" "Mac-specific issue"
create_label "windows"          "0052cc" "Windows-specific issue"
create_label "salesforce"       "00a1e0" "Salesforce integration"
create_label "google-calendar"  "fbbc04" "Google Calendar integration"
create_label "security"         "e11d48" "Security related"

echo ""
echo "Done. Labels created at https://github.com/$REPO/labels"
