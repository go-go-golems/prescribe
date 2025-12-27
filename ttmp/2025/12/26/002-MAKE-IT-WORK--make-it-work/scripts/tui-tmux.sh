#!/usr/bin/env bash
set -euo pipefail

# TUI tmux harness for ticket 002.
#
# Goals:
# - Launch prescribe TUI in tmux (detached) and control it via subcommands.
# - Provide high-level actions (open-filters, toggle, generate, back, quit).
# - Capture panes to files (for diary/debug) without manual tmux plumbing.
#
# Notes:
# - This script assumes tmux is installed.
# - It does NOT require interactive use; it is safe to run repeatedly.

SESSION="${SESSION:-prescribe-002}"
WINDOW="${WINDOW:-tui}"

REPO="${REPO:-/tmp/pr-builder-test-repo}"
TARGET="${TARGET:-main}"

ROOT_DIR="${ROOT_DIR:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe}"
CMD="${CMD:-go run cmd/prescribe/main.go}"

TICKET_SCRIPTS_DIR="${TICKET_SCRIPTS_DIR:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts}"

pane_target() {
  echo "${SESSION}:${WINDOW}.0"
}

tmux_has_session() {
  tmux has-session -t "${SESSION}" 2>/dev/null
}

start() {
  if tmux_has_session; then
    echo "tmux session already running: ${SESSION}"
    return 0
  fi

  # Start detached TUI.
  tmux new-session -d -s "${SESSION}" -n "${WINDOW}" \
    "cd ${ROOT_DIR} && ${CMD} -r ${REPO} -t ${TARGET} tui"

  # Give Bubbletea a moment to render first frame.
  sleep 0.2
  echo "started: ${SESSION} (${WINDOW})"
}

stop() {
  if tmux_has_session; then
    tmux kill-session -t "${SESSION}"
    echo "stopped: ${SESSION}"
    return 0
  fi
  echo "not running: ${SESSION}"
}

send_keys() {
  local keys="$1"
  tmux send-keys -t "$(pane_target)" "${keys}"
}

send_enter() {
  tmux send-keys -t "$(pane_target)" Enter
}

wait_short() {
  # Small sleep to let Bubbletea update; keep this low so scripts are snappy.
  sleep 0.12
}

capture() {
  local label="${1:-capture}"
  local ts
  ts="$(date +%Y%m%d-%H%M%S)"
  local out="${TICKET_SCRIPTS_DIR}/tui-${label}-${ts}.txt"

  # Capture the whole visible pane. If we want scrollback later, add -S option.
  tmux capture-pane -t "$(pane_target)" -p > "${out}"
  echo "${out}"
}

action_quit() {
  send_keys "q"
  wait_short
}

action_back() {
  tmux send-keys -t "$(pane_target)" Escape
  wait_short
}

action_open_filters() {
  send_keys "f"
  wait_short
}

action_toggle_included() {
  tmux send-keys -t "$(pane_target)" Space
  wait_short
}

action_toggle_filtered_view() {
  send_keys "v"
  wait_short
}

action_generate() {
  send_keys "g"
  wait_short
}

action_preset() {
  local n="$1"
  case "${n}" in
    1|2|3) send_keys "${n}" ;;
    *) echo "preset must be 1|2|3" >&2; exit 2 ;;
  esac
  wait_short
}

scenario_smoke() {
  start
  capture "00-start"
  action_open_filters
  capture "01-filters"
  action_preset 1
  capture "02-preset1"
  action_back
  capture "03-back-main"
  action_generate
  # generation might take a tick; capture twice
  capture "04-generating"
  sleep 0.4
  capture "05-result"
  action_back
  capture "06-back-main"
  action_quit
  # process will exit; leave session around until stop() is called (or user wants it).
}

usage() {
  cat <<EOF
Usage: $(basename "$0") <command> [args]

Env vars:
  SESSION=${SESSION}
  WINDOW=${WINDOW}
  REPO=${REPO}
  TARGET=${TARGET}
  ROOT_DIR=${ROOT_DIR}
  CMD=${CMD}
  TICKET_SCRIPTS_DIR=${TICKET_SCRIPTS_DIR}

Commands:
  start                     Start TUI in tmux (detached)
  stop                      Stop tmux session
  capture [label]           Capture current pane to a timestamped file
  send <keys>               Send raw keys (no Enter)
  enter                     Send Enter

Actions (high-level):
  quit                      Send 'q'
  back                      Send Esc
  filters                   Send 'f'
  toggle                    Send Space
  toggle-filtered           Send 'v'
  generate                  Send 'g'
  preset <1|2|3>            Send preset key

Scenarios:
  scenario smoke            Run a basic smoke scenario with captures

Examples:
  $(basename "$0") start
  $(basename "$0") filters && $(basename "$0") preset 1 && $(basename "$0") capture filters
  $(basename "$0") scenario smoke
EOF
}

cmd="${1:-}"
shift || true

case "${cmd}" in
  start) start ;;
  stop) stop ;;
  capture) capture "${1:-capture}" ;;
  send) start; send_keys "${1:?keys required}"; wait_short ;;
  enter) start; send_enter; wait_short ;;

  quit) start; action_quit ;;
  back) start; action_back ;;
  filters) start; action_open_filters ;;
  toggle) start; action_toggle_included ;;
  toggle-filtered) start; action_toggle_filtered_view ;;
  generate) start; action_generate ;;
  preset) start; action_preset "${1:?preset number required}" ;;

  scenario)
    start
    case "${1:-}" in
      smoke) scenario_smoke ;;
      *) echo "unknown scenario: ${1:-}" >&2; exit 2 ;;
    esac
    ;;

  ""|help|-h|--help) usage ;;
  *) echo "unknown command: ${cmd}" >&2; usage; exit 2 ;;
esac


