Compare two versions, branches, or commits of CalSync side by side.

## Usage

- `/compare v0.0.3 v0.0.4` — compare two release tags
- `/compare main release/0.0.4` — compare two branches
- `/compare <commit-sha> <commit-sha>` — compare two specific commits

## Steps

1. Parse `$ARGUMENTS` — expect two refs. If not provided, ask the user.

2. Run:
   ```
   git diff <ref1>...<ref2> --stat
   git log <ref1>..<ref2> --oneline
   ```

3. Produce a side-by-side comparison table:

   | Area | <ref1> | <ref2> |
   |---|---|---|
   | App version | ... | ... |
   | Go version | ... | ... |
   | Key files changed | ... | ... |
   | Commits | N | N |

4. For each changed file, describe what changed between the two refs in plain English

5. Highlight:
   - Breaking changes (credential format, port, file paths)
   - Sync logic changes that could affect existing users' calendars
   - New dependencies

6. End with a **Migration note**: what does a user upgrading from `<ref1>` to `<ref2>` need to do, if anything?
