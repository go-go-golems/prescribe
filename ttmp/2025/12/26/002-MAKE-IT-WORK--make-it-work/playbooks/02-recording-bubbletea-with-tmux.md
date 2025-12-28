---
Title: Recording Bubbletea apps with tmux (playbook)
Ticket: 002-MAKE-IT-WORK
Status: active
Topics:
    - tui
    - bubbletea
    - tmux
    - playbook
DocType: playbook
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "How to run, drive, and capture Bubbletea TUIs using tmux in a repeatable way (including alt-screen pitfalls)."
LastUpdated: 2025-12-27T00:00:00.000000000-05:00
WhatFor: "Repeatable, scriptable TUI recording and debugging without manual keypresses/capture-pane copy/paste."
WhenToUse: "Any time you need to validate or record a Bubbletea UI flow (screenshots as text, regression frames, debugging)."
---

# Recording Bubbletea apps with tmux (playbook)

This playbook describes a reliable workflow to **run a Bubbletea TUI inside tmux**, **drive it from scripts**, and **capture text “screenshots”** via `tmux capture-pane`.

The sharp edge: Bubbletea frequently uses the **alternate screen**. Naive `tmux capture-pane` calls can look “successful” but still capture **blank output** unless you handle timing/buffers correctly.

## Core rules (avoid pain)

- **Prefer a built binary** over `go run` when recording.
  - `go run` adds compile/startup delay and makes timing flaky.
- **Start tmux with an explicit size** when detached (`-x/-y`).
  - Otherwise Bubbletea may not get a sane initial `WindowSizeMsg`.
- **Don’t capture immediately**.
  - Always wait for non-empty pane content.
- **Alt-screen capture is tricky**.
  - `tmux capture-pane -a -p` may fail (“no alternate screen”), or succeed but still be empty.
  - Best approach: try alt capture, but fall back to normal capture if the alt output is whitespace.
- **Keep capture artifacts out of git**.
  - Store captures under `ttmp/.../scripts/` but delete them before committing (or `.gitignore` them).

## Minimal manual workflow (one-off debugging)

### 1) Start the TUI in a detached tmux session

Replace `<CMD>` with your actual command (ideally a built binary).

```bash
tmux new-session -d -x 110 -y 34 -s tui-rec -n w "<CMD>"
sleep 1.5
```

Example (prescribe):

```bash
cd /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe && \
go build -o ./dist/prescribe ./cmd/prescribe && \
tmux new-session -d -x 110 -y 34 -s tui-rec -n w "./dist/prescribe -r /tmp/pr-builder-test-repo -t main tui" && \
sleep 1.5
```

### 2) Send keys

```bash
tmux send-keys -t tui-rec:w.0 f
sleep 0.2
tmux send-keys -t tui-rec:w.0 1
sleep 0.2
tmux send-keys -t tui-rec:w.0 Escape
```

### 3) Capture a “screenshot”

Baseline capture (often works, sometimes blank if you’re on alt-screen):

```bash
tmux capture-pane -t tui-rec:w.0 -p > /tmp/tui.txt
```

If you need scrollback (last 200 lines):

```bash
tmux capture-pane -t tui-rec:w.0 -p -S -200 > /tmp/tui.txt
```

### 4) Stop session

```bash
tmux kill-session -t tui-rec
```

## Reliable capture function (recommended)

This function implements the “alt-screen may be empty” rule:

```bash
capture_tui() {
  local target="$1" out="$2"

  local alt
  alt="$(tmux capture-pane -t "${target}" -a -p 2>/dev/null || true)"

  # If alt capture is non-empty (non-whitespace), keep it.
  if echo "${alt}" | tr -d '\r\n\t ' | grep -q '.'; then
    printf "%s\n" "${alt}" > "${out}"
    return 0
  fi

  # Otherwise fall back to normal capture.
  tmux capture-pane -t "${target}" -p > "${out}" 2>/dev/null || true
}
```

## Waiting until the app is actually rendered

Instead of fixed sleeps, wait for content:

```bash
wait_for_render() {
  local target="$1"
  local retries="${2:-40}"
  local delay="${3:-0.25}"

  local i
  for i in $(seq 1 "${retries}"); do
    local buf
    buf="$(tmux capture-pane -t "${target}" -p 2>/dev/null || true)"
    if echo "${buf}" | tr -d '\r\n\t ' | grep -q '.'; then
      return 0
    fi
    sleep "${delay}"
  done
  return 1
}
```

## Waiting for a specific screen (“result is ready”)

For flows like “Generate” where latency varies, wait for a regex:

```bash
wait_for_text() {
  local target="$1" regex="$2"
  local retries="${3:-60}"
  local delay="${4:-0.25}"

  local i
  for i in $(seq 1 "${retries}"); do
    if tmux capture-pane -t "${target}" -p 2>/dev/null | grep -Eq "${regex}"; then
      return 0
    fi
    sleep "${delay}"
  done
  return 1
}
```

Example:

```bash
tmux send-keys -t tui-rec:w.0 g
wait_for_text tui-rec:w.0 '# Pull Request:' 80 0.25
capture_tui tui-rec:w.0 /tmp/result.txt
```

## Recommended repo pattern: keep a ticket-local harness script

For this ticket we use:

- `scripts/tui-tmux.sh`: start/stop + scenario runner + timestamped captures
- `reference/03-tui-screenshots-smoke-scenario.md`: embeds representative captures into a stable doc

Pattern for future tickets:

- Put the harness in the ticket `scripts/` folder.
- Keep captures timestamped and **delete them before committing**, then paste the interesting frames into a `reference/` doc.



