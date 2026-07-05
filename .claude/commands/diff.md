Show a human-readable summary of what has changed — in the working tree, since the last commit, or between two branches/tags.

## Usage

- `/diff` — changes in the working tree vs last commit
- `/diff main` — changes on current branch vs main
- `/diff v0.0.3 v0.0.4` — changes between two tags or commits

## Steps

1. Parse `$ARGUMENTS`:
   - No args → `git diff HEAD`
   - One arg → `git diff <arg>...HEAD`
   - Two args → `git diff <arg1>...<arg2>`

2. Run `git diff --stat` for the chosen range to get a file-level overview

3. For each changed file, summarise in plain English what changed:
   - `main.go` — describe the functional change (e.g. "added retry logic for Google API calls")
   - `templates/index.html` — describe the UI change
   - `docs/` files — note what documentation was updated
   - `scripts/` files — note what script changed
   - Config / workflow files — note the intent

4. Highlight anything that looks risky:
   - Changes to OAuth or credential handling
   - Changes to the sync logic
   - New dependencies in `go.mod`

5. End with a one-paragraph **Change Summary** suitable for pasting into a release note or PR description
