# Tasks

## TODO

- [ ] Add tasks here

- [x] Create new 'create' command structure: Create prescribe/cmd/prescribe/cmds/create.go with cobra command structure, add to root command initialization, accept flags: --use-last, --yaml-file, --title, --body, --draft, --dry-run, --base
- [ ] Add --create flag to existing generate command: Modify prescribe/cmd/prescribe/cmds/generate.go, add --create flag that triggers PR creation after generation, reuse generation logic then call PR creation
- [x] Implement GitHub CLI integration (gh pr create): Create prescribe/internal/github/github.go with CreatePR function, shell out to gh pr create with appropriate flags, handle gh command execution and capture output/errors
- [x] Implement branch pushing before PR creation: Extend prescribe/internal/git/git.go with PushBranch function, call git push before creating PR, handle push errors gracefully
- [x] Implement session data reuse (--use-last): Read last generated PR data from session file (.pr-builder/session.yaml), parse GeneratedPRData from session, use this data when --use-last flag is provided
- [x] Implement YAML file input (--yaml-file): Add --yaml-file flag to create command, read and parse YAML file containing GeneratedPRData, use parsed data for PR creation
- [x] Implement title/body override flags (--title, --body): Add --title and --body flags to create command, override generated title/body when flags are provided, support both flags together or individually
- [x] Implement draft PR support (--draft): Add --draft flag to create command, pass --draft flag to gh pr create when flag is set, default to false (not draft)
- [x] Implement dry-run mode (--dry-run): Add --dry-run flag to create command, when set show what would be created without actually calling gh pr create, display title/body/base branch/draft status
- [x] Implement error handling with PR data save: On PR creation failure save generated PR data to file (e.g., .pr-builder/pr-data-<timestamp>.yaml), include clear error message indicating where data was saved, exit with appropriate error code
- [x] Implement base branch handling (--base): Add --base flag to create command, default to main (or detected default branch via git.GetDefaultBranch()), pass --base flag to gh pr create
- [ ] Wire up generate --create flow: After successful generation in generate command call PR creation logic, use generated PR data for creation, handle errors appropriately
- [x] Add tests for PR creation: Create prescribe/internal/github/github_test.go, test gh pr create command construction with various flags, mock gh command execution for unit tests
- [ ] Update documentation: Update prescribe/README.md with create command usage, document --use-last/--yaml-file/--title/--body/--draft/--dry-run/--base flags, add examples for common workflows
- [ ] Integration test: end-to-end PR creation: Test full flow (prescribe generate → prescribe create --use-last), test prescribe generate --create, test prescribe create --yaml-file <file>, verify PR is actually created (or mocked appropriately)
