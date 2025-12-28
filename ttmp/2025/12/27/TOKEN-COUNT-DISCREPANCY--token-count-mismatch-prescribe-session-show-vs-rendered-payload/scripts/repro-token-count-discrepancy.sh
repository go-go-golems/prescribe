#!/usr/bin/env bash
set -euo pipefail

# Repro script for TOKEN-COUNT-DISCREPANCY (minimal output).
#
# What it does:
# - (Re)initializes a deterministic session state
# - Exports context + rendered payload (XML-ish)
# - Captures prescribe's own breakdowns (session token-count, count-xml, rendered payload counts)
# - Optionally compares against pinocchio's token counter on the exact same bytes
#
# Environment overrides:
# - REPO: git repo to analyze (default: /home/manuel/workspaces/2025-12-26/prescribe-import/prescribe)
# - TARGET: target branch/ref (default: origin/main)
# - PRESCRIBE_BIN: prescribe binary (default: $REPO/prescribe or go run fallback)
# - PINOCCHIO_BIN: pinocchio binary (default: pinocchio on PATH, else go run fallback)
# - BASE: output file prefix (default: /tmp/prescribe-token-discrepancy-<timestamp>)

REPO="${REPO:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe}"
TARGET="${TARGET:-origin/main}"
BASE="${BASE:-/tmp/prescribe-token-discrepancy-$(date +%Y%m%d-%H%M%S)}"

SESSION_SHOW_JSON="${BASE}.session-show.json"
SESSION_TOKEN_JSON="${BASE}.session-token-count.json"
CONTEXT_XML="${BASE}.context.xml"
RENDERED_XML="${BASE}.rendered.xml"
COUNTXML_CONTEXT_JSON="${BASE}.countxml-context.json"
COUNTXML_RENDERED_JSON="${BASE}.countxml-rendered.json"
LOG="${BASE}.log"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

die_tail() {
  echo "FAILED. See log: ${LOG}" >&2
  tail -200 "$LOG" >&2 || true
  exit 1
}

resolve_prescribe() {
  if [[ -n "${PRESCRIBE_BIN:-}" ]]; then
    echo "$PRESCRIBE_BIN"
    return 0
  fi
  if [[ -x "${REPO}/prescribe" ]]; then
    # The repo may contain a stale prebuilt binary. Verify it supports `session`.
    if "${REPO}/prescribe" session --help >/tmp/prescribe-repro-session-help.$$ 2>&1; then
      rm -f /tmp/prescribe-repro-session-help.$$ || true
      echo "${REPO}/prescribe"
      return 0
    fi
    if grep -q "unknown command \"session\"" /tmp/prescribe-repro-session-help.$$ 2>/dev/null; then
      # stale/old binary; fall back to go run
      rm -f /tmp/prescribe-repro-session-help.$$ || true
    else
      rm -f /tmp/prescribe-repro-session-help.$$ || true
      echo "${REPO}/prescribe"
      return 0
    fi
  fi
  # Fallback: go run (slower, but always matches current source tree)
  echo "go run ./cmd/prescribe"
}

resolve_pinocchio() {
  if [[ -n "${PINOCCHIO_BIN:-}" ]]; then
    echo "$PINOCCHIO_BIN"
    return 0
  fi
  if command -v pinocchio >/dev/null 2>&1; then
    echo "pinocchio"
    return 0
  fi
  # Fallback: go run from workspace checkout
  echo "go run ./cmd/pinocchio"
}

PRESCRIBE="$(resolve_prescribe)"
PINOCCHIO="$(resolve_pinocchio)"

echo "TOKEN-COUNT-DISCREPANCY repro (minimal output)" >"$LOG"
echo "REPO=${REPO}" >>"$LOG"
echo "TARGET=${TARGET}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"
echo "PRESCRIBE=${PRESCRIBE}" >>"$LOG"
echo "PINOCCHIO=${PINOCCHIO}" >>"$LOG"

run_quiet "git fetch" git -C "$REPO" fetch --all --prune

if [[ "$PRESCRIBE" == "go run ./cmd/prescribe" ]]; then
  run_quiet "session init (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" session init --save"
  run_quiet "filter clear (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" filter clear"
  run_quiet "filter add excludes (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" filter add --name \"Trim huge docs\" --exclude 'ttmp/**' --exclude 'TUI-SCREENSHOTS*' --exclude 'FILTER-*.md' --exclude 'PLAYBOOK-*.md' --exclude 'dev-diary.md' --exclude 'PROJECT-SUMMARY.md' --exclude 'TUI-DEMO.md' --exclude '*.pdf'"
  run_quiet "context add (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" context add PROJECT-SUMMARY.md && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" context add README.md && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" context add ttmp/2025/12/27/008-GENERATE--generate-pr-descriptions-using-ai-inference/reference/01-diary.md && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" context add ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md"

  run_quiet "session show json (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" session show --output json > \"$SESSION_SHOW_JSON\""
  run_quiet "session token-count json (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" session token-count --output json > \"$SESSION_TOKEN_JSON\""

  run_quiet "export context xml (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" generate --export-context --separator xml --output-file \"$CONTEXT_XML\""
  run_quiet "export rendered xml + print rendered counts (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE -r \"$REPO\" -t \"$TARGET\" generate --export-rendered --separator xml --print-rendered-token-count --output-file \"$RENDERED_XML\""

  run_quiet "count-xml context (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE tokens count-xml --file \"$CONTEXT_XML\" --output json > \"$COUNTXML_CONTEXT_JSON\""
  run_quiet "count-xml rendered (go run)" bash -c "cd \"$REPO\" && $PRESCRIBE tokens count-xml --file \"$RENDERED_XML\" --output json > \"$COUNTXML_RENDERED_JSON\""
else
  run_quiet "session init" "$PRESCRIBE" -r "$REPO" -t "$TARGET" session init --save
  run_quiet "filter clear" "$PRESCRIBE" -r "$REPO" -t "$TARGET" filter clear
  run_quiet "filter add excludes" "$PRESCRIBE" -r "$REPO" -t "$TARGET" filter add --name "Trim huge docs" \
    --exclude 'ttmp/**' \
    --exclude 'TUI-SCREENSHOTS*' \
    --exclude 'FILTER-*.md' \
    --exclude 'PLAYBOOK-*.md' \
    --exclude 'dev-diary.md' \
    --exclude 'PROJECT-SUMMARY.md' \
    --exclude 'TUI-DEMO.md' \
    --exclude '*.pdf'

  run_quiet "context add" "$PRESCRIBE" -r "$REPO" -t "$TARGET" context add PROJECT-SUMMARY.md
  run_quiet "context add" "$PRESCRIBE" -r "$REPO" -t "$TARGET" context add README.md
  run_quiet "context add" "$PRESCRIBE" -r "$REPO" -t "$TARGET" context add ttmp/2025/12/27/008-GENERATE--generate-pr-descriptions-using-ai-inference/reference/01-diary.md
  run_quiet "context add" "$PRESCRIBE" -r "$REPO" -t "$TARGET" context add ttmp/2025/12/26/002-MAKE-IT-WORK--make-it-work/reference/02-diary.md

  run_quiet "session show json" bash -lc "\"$PRESCRIBE\" -r \"$REPO\" -t \"$TARGET\" session show --output json > \"$SESSION_SHOW_JSON\""
  run_quiet "session token-count json" bash -lc "\"$PRESCRIBE\" -r \"$REPO\" -t \"$TARGET\" session token-count --output json > \"$SESSION_TOKEN_JSON\""

  run_quiet "export context xml" "$PRESCRIBE" -r "$REPO" -t "$TARGET" generate --export-context --separator xml --output-file "$CONTEXT_XML"
  run_quiet "export rendered xml + print rendered counts" "$PRESCRIBE" -r "$REPO" -t "$TARGET" generate --export-rendered --separator xml --print-rendered-token-count --output-file "$RENDERED_XML"

  run_quiet "count-xml context" bash -lc "\"$PRESCRIBE\" tokens count-xml --file \"$CONTEXT_XML\" --output json > \"$COUNTXML_CONTEXT_JSON\""
  run_quiet "count-xml rendered" bash -lc "\"$PRESCRIBE\" tokens count-xml --file \"$RENDERED_XML\" --output json > \"$COUNTXML_RENDERED_JSON\""
fi

# Extract a few key numbers (keep stdout minimal).
python3 - "$SESSION_SHOW_JSON" "$SESSION_TOKEN_JSON" "$COUNTXML_CONTEXT_JSON" "$COUNTXML_RENDERED_JSON" "$LOG" <<'PY' || die_tail
import json, sys, re, pathlib

session_show = json.loads(pathlib.Path(sys.argv[1]).read_text())
session_token = json.loads(pathlib.Path(sys.argv[2]).read_text())
countxml_ctx = json.loads(pathlib.Path(sys.argv[3]).read_text())
countxml_r = json.loads(pathlib.Path(sys.argv[4]).read_text())
log_text = pathlib.Path(sys.argv[5]).read_text()

def rows(x):
    if isinstance(x, list):
        return x
    if isinstance(x, dict):
        # glazed sometimes uses {"rows":[...]}
        if "rows" in x and isinstance(x["rows"], list):
            return x["rows"]
    return []

def first_row_with(rs, **conds):
    for r in rs:
        ok = True
        for k, v in conds.items():
            if r.get(k) != v:
                ok = False
                break
        if ok:
            return r
    return None

ss = rows(session_show)
st = rows(session_token)
ctx = rows(countxml_ctx)
rr = rows(countxml_r)

ss0 = ss[0] if ss else {}
total_row = first_row_with(st, kind="total") or {}
ctx_doc = first_row_with(ctx, kind="document") or {}
r_doc = first_row_with(rr, kind="document") or {}
r_llm = first_row_with(rr, kind="section", section_tag="llm_payload") or {}
r_sys = first_row_with(rr, kind="cdata", tag="system") or {}
r_user = first_row_with(rr, kind="cdata", tag="user") or {}

# Pull rendered token count lines printed by generate flag from log.
rendered_line = ""
export_line = ""
for line in log_text.splitlines():
    if line.startswith("Rendered payload token counts"):
        rendered_line = line.strip()
    if line.startswith("Rendered payload export token count"):
        export_line = line.strip()

encoding = (total_row.get("encoding") or ss0.get("encoding") or "").strip()

print("=== TOKEN-COUNT-DISCREPANCY summary ===")
print(f"encoding(prescribe)={encoding or 'unknown'}")
print(f"session_show.token_count={ss0.get('token_count')}")
print(f"session_token_count.stored_total={total_row.get('stored_total')} effective_total={total_row.get('effective_total')} delta={total_row.get('delta')}")
print(f"context_xml.tokens={ctx_doc.get('tokens')} bytes={ctx_doc.get('bytes')}")
print(f"rendered_xml.tokens={r_doc.get('tokens')} bytes={r_doc.get('bytes')}")
print(f"rendered_xml.llm_payload.tokens={r_llm.get('tokens')} system_cdata.tokens={r_sys.get('tokens')} user_cdata.tokens={r_user.get('tokens')}")
if rendered_line:
    print(rendered_line)
if export_line:
    print(export_line)
print(f"artifacts: {sys.argv[1]} {sys.argv[2]} {sys.argv[3]} {sys.argv[4]}")
print(f"log: {sys.argv[5]}")
PY

# Optional: compare pinocchio token count on the exact same bytes (rendered.xml).
# We keep this quiet; failures won't abort the script.
if [[ "$PINOCCHIO" == "go run ./cmd/pinocchio" ]]; then
  {
    echo
    echo "==> pinocchio tokens count (go run) on rendered.xml"
    cd "/home/manuel/workspaces/2025-12-26/prescribe-import/pinocchio" && cat "$RENDERED_XML" | go run ./cmd/pinocchio tokens count --model '' --codec cl100k_base - || true
  } >>"$LOG" 2>&1
else
  {
    echo
    echo "==> pinocchio tokens count on rendered.xml"
    cat "$RENDERED_XML" | "$PINOCCHIO" tokens count --model '' --codec cl100k_base - || true
  } >>"$LOG" 2>&1
fi

echo
echo "pinocchio comparison written to log: ${LOG}"
echo "done"


