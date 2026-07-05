Stage, commit, and push all current changes to the remote branch.

## Steps

1. Run `git status` and show the user what files are changed, added, or deleted
2. Run `git diff --stat` to summarise what changed
3. Check for any sensitive files about to be committed:
   - Warn and stop if any of these are staged: `google_credentials.json`, `sf_credentials.json`, `google_token.json`, `*.log`
4. Ask the user for a commit message if not provided as `$ARGUMENTS`
   - If provided, use it directly
   - If not provided, suggest one based on the diff and ask for confirmation
5. Stage all safe changes: `git add -A`
6. Commit with the agreed message
7. Push to the current branch: `git push`
8. Report the commit hash, branch, and link to the branch on GitHub:
   `https://github.com/nidhijain-sf/CalSync/tree/<branch>`
