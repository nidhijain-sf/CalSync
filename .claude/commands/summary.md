Generate a change summary for the current branch — what changed, why it matters, and what to tell users.

## Steps

1. Determine the base: compare current branch against `main`
   ```
   git log main..HEAD --oneline
   git diff main...HEAD --stat
   ```

2. List all commits on this branch with their messages

3. Group changes by category:
   - **Bug fixes** — what was broken and is now fixed
   - **New features** — what new behaviour was added
   - **UI changes** — any changes to `templates/index.html`
   - **Docs** — documentation added or updated
   - **Infrastructure** — scripts, workflows, CI changes
   - **Dependency updates** — changes to `go.mod` / `go.sum`

4. For each changed file read the actual diff and describe the change in plain English (not just the filename)

5. Produce three outputs:

   ### Release Notes (for GitHub Release)
   Short bullet list of what end users care about — no technical jargon

   ### PR Description (for pull request body)
   More detailed summary with technical context, suitable for a code reviewer

   ### Slack Message (for team announcement)
   One or two sentences max — what shipped and why it matters
