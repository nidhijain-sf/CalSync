Create a new CalSync release.

## Steps

1. Ask the user for the version number if not provided as `$ARGUMENTS` (e.g. `1.0.0`)
2. Validate the version follows semver (MAJOR.MINOR.PATCH)
3. Check the working tree is clean — if not, list the uncommitted changes and stop
4. Run `./scripts/build.sh` to build fresh Mac and Windows binaries and sync templates
5. Stage the dist binaries:
   ```
   git add dist/mac/SyncApp dist/windows/SyncApp.exe dist/mac/templates dist/windows/templates
   ```
6. Commit with message: `Release v<version>`
7. Create git tag: `v<version>`
8. Show the user a summary of what was built and ask for confirmation before pushing
9. On confirmation, push: `git push origin <current-branch> && git push origin v<version>`
10. Open the GitHub Actions page: `open https://github.com/nidhijain-sf/CalSync/actions`
11. Tell the user the release will appear at: `https://github.com/nidhijain-sf/CalSync/releases`
