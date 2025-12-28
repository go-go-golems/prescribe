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
# Default to the built binary for fast startup. You can override to `go run ...` if needed.
CMD="${CMD:-./dist/prescribe}"

TICKET_SCRIPTS_DIR="${TICKET_SCRIPTS_DIR:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe/ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/scripts}"

# tmux/window sizing + pacing
TMUX_COLS="${TMUX_COLS:-110}"
TMUX_ROWS="${TMUX_ROWS:-34}"
START_WAIT="${START_WAIT:-1.2}"
ACTION_WAIT="${ACTION_WAIT:-0.20}"
CAPTURE_WAIT="${CAPTURE_WAIT:-0.25}"
CAPTURE_RETRIES="${CAPTURE_RETRIES:-20}"
GENERATE_WAIT="${GENERATE_WAIT:-3.5}"

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

  if [ "${CMD}" = "./dist/prescribe" ] && [ ! -x "${ROOT_DIR}/dist/prescribe" ]; then
    echo "ERROR: ${ROOT_DIR}/dist/prescribe is missing. Run:" >&2
    echo "  cd ${ROOT_DIR} && go build -o ./dist/prescribe ./cmd/prescribe" >&2
    exit 1
  fi

  # Start detached TUI.
  #
  # NOTE: Bubbletea relies on terminal size events. When a tmux session is started detached,
  # Bubbletea may never receive an initial window size unless we create the session with a size.
  tmux new-session -d -x "${TMUX_COLS}" -y "${TMUX_ROWS}" -s "${SESSION}" -n "${WINDOW}" \
    "cd ${ROOT_DIR} && ${CMD} -r ${REPO} -t ${TARGET} tui"

  # Give Bubbletea a moment to render first frame.
  sleep "${START_WAIT}"
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
  sleep "${ACTION_WAIT}"
}

capture_buffer() {
  local target
  target="$(pane_target)"

  # Bubbletea draws on the alternate screen, but tmux may report success with an empty alt buffer.
  # So: try alt-screen capture, but fall back to normal capture if alt output is blank.
  local alt
  alt="$(tmux capture-pane -t "${target}" -a -p 2>/dev/null || true)"
  if echo "${alt}" | tr -d '\r\n\t ' | grep -q '.'; then
    printf "%s\n" "${alt}"
    return 0
  fi

  tmux capture-pane -t "${target}" -p 2>/dev/null || true
}

wait_for_render() {
  # Wait until the pane contains some non-whitespace output.
  local i
  for i in $(seq 1 "${CAPTURE_RETRIES}"); do
    local buf
    buf="$(capture_buffer)"
    if echo "${buf}" | tr -d '\r\n\t ' | grep -q '.'; then
      return 0
    fi
    sleep "${CAPTURE_WAIT}"
  done
  return 1
}

capture() {
  local label="${1:-capture}"
  local ts
  ts="$(date +%Y%m%d-%H%M%S)"
  local out="${TICKET_SCRIPTS_DIR}/tui-${label}-${ts}.txt"

  # Wait for Bubbletea to actually render something before capturing.
  wait_for_render >/dev/null 2>&1 || true

  capture_buffer > "${out}"
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
  sleep "${GENERATE_WAIT}"
  capture "05-after-generate"
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
  build                     Build ./dist/prescribe (recommended before scenario runs)
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
  $(basename "$0") build && $(basename "$0") scenario smoke
  $(basename "$0") start
  $(basename "$0") filters && $(basename "$0") preset 1 && $(basename "$0") capture filters
  $(basename "$0") scenario smoke
EOF
}

cmd="${1:-}"
shift || true

case "${cmd}" in
  start) start ;;
  build) cd "${ROOT_DIR}" && go build -o ./dist/prescribe ./cmd/prescribe ;;
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
    case "${1:-}" in
      smoke) scenario_smoke ;;
      *) echo "unknown scenario: ${1:-}" >&2; exit 2 ;;
    esac
    ;;

  ""|help|-h|--help) usage ;;
  *) echo "unknown command: ${cmd}" >&2; usage; exit 2 ;;
esac


