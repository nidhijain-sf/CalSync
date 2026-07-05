Create a new git branch following CalSync's branching conventions.

## Usage

- `/branch` — ask the user what kind of branch they need
- `/branch feature/my-feature` — create a specific branch
- `/branch 1.2.3` — create a release branch `release/1.2.3`

## Branching Conventions

| Type | Pattern | When to use |
|---|---|---|
| Release | `release/1.2.3` | New version being prepared |
| Feature | `feature/short-description` | New functionality |
| Bug fix | `fix/short-description` | Bug fix |
| Hotfix | `hotfix/short-description` | Urgent fix on top of a release |
| Docs | `docs/short-description` | Documentation only changes |
| Chore | `chore/short-description` | Scripts, CI, config changes |

## Steps

1. Parse `$ARGUMENTS`:
   - If a plain version number (e.g. `1.2.3`) → branch name is `release/1.2.3`
   - If already has a prefix (e.g. `feature/foo`) → use as-is
   - If a short description with no prefix → ask the user which type (feature / fix / hotfix / docs / chore)
   - If empty → ask the user what they want to work on and suggest a branch name

2. Check the current branch and working tree:
   - Run `git status` — if there are uncommitted changes, warn the user and ask whether to stash them or stay
   - Show the current branch

3. Confirm the new branch name with the user before creating

4. Create and switch to the branch:
   ```
   git checkout -b <branch-name>
   ```

5. Push the branch to remote and set upstream:
   ```
   git push -u origin <branch-name>
   ```

6. Report:
   - New branch name
   - Based off which branch
   - Remote tracking set up
   - Suggested next step (e.g. "Start coding, then use `/push` to commit your changes")
