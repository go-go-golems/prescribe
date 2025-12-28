---
Title: Smoke testing with `tui-tmux.sh` (playbook)
Ticket: 002-MAKE-IT-WORK
Status: active
Topics:
    - tui
    - testing
    - tmux
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "How to use the ticket tmux harness script for repeatable Bubbletea smoke tests and captures."
LastUpdated: 2025-12-27T00:00:00.000000000-05:00
WhatFor: "Fast regression checks of the TUI during refactors."
WhenToUse: "Before/after UI changes; when you want repeatable 'screenshots' without manual keypress/capture-pane."
---

# Smoke testing with `tui-tmux.sh`

This playbook is for **future you**: how to run the tmux harness we built for ticket 002.

Script:

```bash
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh
```

## Quick Start (recommended)

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe

# Build latest binary (script defaults to ./dist/prescribe)
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh build

# Run smoke scenario (creates timestamped capture files under ticket scripts/)
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh scenario smoke

# Stop tmux session (cleanup)
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh stop
```

## Quick smoke test (non-interactive, no tmux)

When you just want to verify the TUI **starts, renders, and exits cleanly** (e.g. in CI or during refactors), run it under a pseudo-tty using `script` and kill it after a short timeout:

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && \
go build -o ./dist/prescribe ./cmd/prescribe && \
timeout 2s script -q -e -c "./dist/prescribe --repo /tmp/pr-builder-test-repo tui" /dev/null
```

Notes:
- `script` provides a pseudo-tty (Bubbletea often won’t render correctly without one).
- `timeout` avoids hanging if the program waits for input.
- Use `--repo` pointing at a prepared test repo (see ticket’s CLI testing playbook).

### Where the “screenshots” go

The script writes **text captures** like:

```
ttmp/.../scripts/tui-00-start-YYYYMMDD-HHMMSS.txt
ttmp/.../scripts/tui-01-filters-YYYYMMDD-HHMMSS.txt
...
```

They are meant to be pasted into:
- `reference/03-tui-screenshots-smoke-scenario.md`

### Important: don’t commit captures

Before committing:

```bash
rm -f ./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-*.txt
```

## Common operations (manual driving)

Start the TUI (detached tmux session):

```bash
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh start
```

Send actions:

```bash
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh filters
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh preset 1
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh back
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh generate
```

Capture a frame:

```bash
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh capture after-generate
```

Quit and stop:

```bash
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh quit
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh stop
```

## Tuning (when it’s flaky)

The harness is intentionally configurable via env vars:

- **Session/window**
  - `SESSION` (default `prescribe-002`)
  - `WINDOW` (default `tui`)
- **Repo/target**
  - `REPO` (default `/tmp/pr-builder-test-repo`)
  - `TARGET` (default `main`)
- **Command**
  - `CMD` (default `./dist/prescribe`)
- **Timing**
  - `START_WAIT` (default `1.2`)
  - `ACTION_WAIT` (default `0.20`)
  - `GENERATE_WAIT` (default `3.5`)
  - `CAPTURE_WAIT` (default `0.25`)
  - `CAPTURE_RETRIES` (default `20`)
- **Detached tmux size**
  - `TMUX_COLS` (default `110`)
  - `TMUX_ROWS` (default `34`)

Example: slow machine / slow generate

```bash
START_WAIT=2.0 ACTION_WAIT=0.35 GENERATE_WAIT=6.0 \
./ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts/tui-tmux.sh scenario smoke
```


