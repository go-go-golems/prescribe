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
```


