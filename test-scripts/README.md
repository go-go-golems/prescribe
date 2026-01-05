# test-scripts

These are **manual smoke-test scripts** for `prescribe`. They create a temporary git repo under `/tmp` and then run a sequence of `prescribe` CLI commands against it.

## Environment variables

- `TEST_REPO_DIR`: path for the temporary test repo (default: `/tmp/prescribe-test-repo`)
- `TARGET_BRANCH`: target branch for the diff (default: `master`)

## Usage

```bash
cd prescribe

# Full smoke suite
bash test-scripts/test-all.sh

# Minimal CLI sanity check
bash test-scripts/test-cli.sh

# Filter-specific checks
bash test-scripts/test-filters.sh

# Session-centric checks
bash test-scripts/test-session-cli.sh

# PR creation integration (safe: local git remote + fake gh)
#
# NOTE: The smoke scripts use `generate --export-context` / `--export-rendered` so they
# don't require AI profiles/API keys. The PR creation integration test still requires
# real AI step settings unless you set SKIP_GENERATE=1.
bash test-scripts/08-integration-test-pr-creation.sh
```

