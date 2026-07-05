# Scripts & Automation

## Scripts (for developers)

All scripts live in `scripts/`. Run them from the repo root.

```bash
chmod +x scripts/*.sh   # make executable (first time only)
```

### `scripts/run-local.sh` — Run locally for development

```bash
./scripts/run-local.sh
```

Starts the app at `http://localhost:5001`. Requires `google_credentials.json` in the repo root (not committed — copy it in manually).

---

### `scripts/build.sh` — Build both binaries + sync templates

```bash
./scripts/build.sh
```

Builds `SyncApp` (Mac) and `SyncApp.exe` (Windows) and places them in `dist/mac/` and `dist/windows/`. Also copies `templates/index.html` to both dist folders.

---

### `scripts/release.sh` — Tag and prepare a release

```bash
./scripts/release.sh 1.2.3
```

1. Checks the working tree is clean
2. Runs `build.sh` to compile fresh binaries
3. Stages and commits the dist changes
4. Creates a git tag `v1.2.3`

Then push manually:
```bash
git push origin main && git push origin v1.2.3
```

Pushing the tag triggers the GitHub Actions release workflow (see below).

---

### `scripts/setup-labels.sh` — Create GitHub labels

```bash
./scripts/setup-labels.sh
```

Creates all standard labels on the GitHub repo. Requires the `gh` CLI to be installed and authenticated:
```bash
gh auth login
```

---

## GitHub Actions Workflows

### Build Check (`.github/workflows/build-check.yml`)

Runs automatically on every push to `main` or `release/**` branches, and on every pull request to `main`.

**What it does:**
- Compiles the Go code for both Mac and Windows
- Checks that no credentials or secrets are accidentally committed

### Release (`.github/workflows/release.yml`)

Runs automatically when a version tag (`v*.*.*`) is pushed to GitHub.

**What it does:**
1. Builds fresh Mac and Windows binaries
2. Creates zip packages (`CalSync-mac-v1.2.3.zip`, `CalSync-windows-v1.2.3.zip`)
3. Creates a GitHub Release with auto-generated release notes and the zips attached

**To trigger a release:**
```bash
./scripts/release.sh 1.2.3
git push origin main && git push origin v1.2.3
```

---

## GitHub Labels

Labels are defined in `.github/labels.yml` and can be applied via `scripts/setup-labels.sh`.

| Label | Colour | Purpose |
|---|---|---|
| `bug` | Red | Something isn't working |
| `enhancement` | Cyan | New feature or improvement |
| `documentation` | Blue | Docs changes |
| `needs-triage` | Yellow | Needs investigation |
| `good first issue` | Purple | Good for newcomers |
| `help wanted` | Green | Extra attention needed |
| `release` | Dark green | Release related |
| `mac` | Blue | Mac-specific |
| `windows` | Dark blue | Windows-specific |
| `salesforce` | SF Blue | Salesforce integration |
| `google-calendar` | Yellow | Google Calendar integration |
| `security` | Pink-red | Security related |
| `wontfix` | White | Will not be addressed |
| `duplicate` | Grey | Already exists |

---

## Versioning

CalSync follows [Semantic Versioning](https://semver.org/):

```
v MAJOR . MINOR . PATCH
```

| Part | When to bump |
|---|---|
| `MAJOR` | Breaking change (e.g. removed feature, changed credentials format) |
| `MINOR` | New feature added (e.g. new sync option, UI addition) |
| `PATCH` | Bug fix or small improvement |

Examples: `v1.0.0` → `v1.0.1` (bug fix) → `v1.1.0` (new feature) → `v2.0.0` (breaking change)
