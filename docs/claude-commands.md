# Claude Code Commands

CalSync ships with project-level slash commands for Claude Code. These are available in any Claude Code session opened inside this repo.

## Available Commands

| Command | Usage | What it does |
|---|---|---|
| `/release` | `/release 1.2.3` | Build binaries, commit, tag, and push a new release |
| `/test` | `/test` | Check the app is running, accounts connected, and log is healthy |
| `/push` | `/push "my message"` | Stage, commit, and push changes safely (blocks credentials) |
| `/diff` | `/diff` or `/diff main` or `/diff v0.0.3 v0.0.4` | Human-readable summary of what changed |
| `/summary` | `/summary` | Change summary formatted as release notes, PR description, and Slack message |
| `/compare` | `/compare v0.0.3 v0.0.4` | Side-by-side comparison of two versions or branches |
| `/log` | `/log` or `/log full` | Read and summarise `calsync.log` in plain English |

---

## Examples

### Cut a new release
```
/release 0.0.5
```
Claude will build both binaries, commit, tag `v0.0.5`, confirm with you, then push and open GitHub Actions.

### Check app health before a release
```
/test
```
Claude will verify the app is running, both accounts are connected, and the log has no errors.

### Get a change summary before opening a PR
```
/summary
```
Claude will produce ready-to-paste release notes, a PR description, and a Slack announcement.

### See what changed between two releases
```
/compare v0.0.3 v0.0.4
```
Claude will diff the two tags and explain every change in plain English, including any migration steps for users upgrading.

### Push your changes
```
/push "fix scheduler not triggering after sleep"
```
Claude will check for accidental credential files, stage everything safe, and push.

---

## How They Work

Commands live in `.claude/commands/` as Markdown files. Each file is a prompt given to Claude Code when you type the slash command. They have access to all the same tools Claude Code normally uses — reading files, running bash, git, etc.
