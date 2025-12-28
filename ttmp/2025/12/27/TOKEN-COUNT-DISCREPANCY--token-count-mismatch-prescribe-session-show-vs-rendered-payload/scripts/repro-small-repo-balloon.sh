#!/usr/bin/env bash
set -euo pipefail

# Small-repo repro for "rendered prompt balloons" hypotheses.
#
# Goal: produce *tiny* artifacts that we can diff/inspect to confirm whether the default
# prompt template duplicates the main context (`.bracket` causes a second render).
#
# This script:
# - creates a small throwaway git repo via prescribe's existing smoke-test setup
# - initializes a session against master
# - exports rendered XML twice:
#   (A) default prompt (current behavior)
#   (B) modified prompt with any bracketed second `{{ template "context" . }}` removed (best-effort)
# - extracts the user CDATA text and counts duplication markers
#
# Minimal stdout; everything else goes into a log.

PRESCRIBE_ROOT="${PRESCRIBE_ROOT:-/home/manuel/workspaces/2025-12-26/prescribe-import/prescribe}"
TEST_REPO_DIR="${TEST_REPO_DIR:-/tmp/prescribe-test-repo}"
TARGET_BRANCH="${TARGET_BRANCH:-master}"
BASE="${BASE:-/tmp/prescribe-smallrepo-balloon-$(date +%Y%m%d-%H%M%S)}"

LOG="${BASE}.log"
SESSION_SHOW_JSON="${BASE}.session-show.json"
SESSION_TOKEN_JSON="${BASE}.session-token-count.json"
SESSION_NODUP_YAML="${BASE}.session.nodup.yaml"

RENDERED_DEFAULT_XML="${BASE}.rendered.default.xml"
RENDERED_NODUP_XML="${BASE}.rendered.nodup.xml"
USER_DEFAULT_TXT="${BASE}.user.default.txt"
USER_NODUP_TXT="${BASE}.user.nodup.txt"
PROMPT_DEFAULT_TXT="${BASE}.prompt.default.txt"
PROMPT_NODUP_TXT="${BASE}.prompt.nodup.txt"

run_quiet() {
  local label="$1"
  shift
  {
    echo
    echo "==> ${label}"
    "$@"
  } >>"$LOG" 2>&1
}

echo "small-repo balloon repro" >"$LOG"
echo "PRESCRIBE_ROOT=${PRESCRIBE_ROOT}" >>"$LOG"
echo "TEST_REPO_DIR=${TEST_REPO_DIR}" >>"$LOG"
echo "TARGET_BRANCH=${TARGET_BRANCH}" >>"$LOG"
echo "BASE=${BASE}" >>"$LOG"

# Build small repo (smoke-test helper). Keep noise out of stdout.
run_quiet "setup small test repo" bash "${PRESCRIBE_ROOT}/test-scripts/setup-test-repo.sh"

# Ensure we run the current prescribe source (not a potentially stale binary).
prescribe() {
  ( cd "$PRESCRIBE_ROOT" && go run ./cmd/prescribe "$@" )
}

run_quiet "session init/save" prescribe --repo "$TEST_REPO_DIR" --target "$TARGET_BRANCH" session init --save
run_quiet "session show json" bash -c "cd \"$PRESCRIBE_ROOT\" && go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" session show --output json > \"$SESSION_SHOW_JSON\""
run_quiet "session token-count json" bash -c "cd \"$PRESCRIBE_ROOT\" && go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" session token-count --output json > \"$SESSION_TOKEN_JSON\""

# Reconstruct the combined default prompt (system + prompt) exactly like internal/prompts/default.go does.
python3 - "$PRESCRIBE_ROOT" "$PROMPT_DEFAULT_TXT" "$PROMPT_NODUP_TXT" >>"$LOG" 2>&1 <<'PY'
import pathlib, sys, re, yaml

root = pathlib.Path(sys.argv[1])
yml_path = root / "internal" / "prompts" / "assets" / "create-pull-request.yaml"
out_default = pathlib.Path(sys.argv[2])
out_nodup = pathlib.Path(sys.argv[3])

data = yaml.safe_load(yml_path.read_text())
system = (data.get("system-prompt") or "").strip()
prompt = (data.get("prompt") or "").strip()
combined = system + "\n\n" + prompt
out_default.write_text(combined)

# Remove the bracketed duplication block:
# {{ if .bracket }} ... {{ template "context" . }} ... {{ end }}
nodup = combined
nodup = re.sub(
    r"\{\{\s*if\s*\.bracket\s*\}\}[\s\S]*?\{\{\s*template\s+\"context\"\s+\.\s*\}\}[\s\S]*?\{\{\s*end\s*\}\}\s*",
    "",
    nodup,
    count=1,
)
out_nodup.write_text(nodup)
PY

# Patch a copy of the session YAML to use the no-dup prompt template, so we don't have to pass a huge multi-line
# string through a CLI flag (which is fragile to quote correctly).
python3 - "$TEST_REPO_DIR" "$PROMPT_NODUP_TXT" "$SESSION_NODUP_YAML" >>"$LOG" 2>&1 <<'PY'
import pathlib, sys, yaml

repo = pathlib.Path(sys.argv[1])
prompt_txt = pathlib.Path(sys.argv[2]).read_text()
out = pathlib.Path(sys.argv[3])

session_path = repo / ".pr-builder" / "session.yaml"
data = yaml.safe_load(session_path.read_text())

data.setdefault("prompt", {})
data["prompt"]["preset"] = ""
data["prompt"]["template"] = prompt_txt

out.write_text(yaml.safe_dump(data, sort_keys=False))
PY

run_quiet "export rendered (default prompt) + print rendered token counts" bash -c "cd \"$PRESCRIBE_ROOT\" && go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" generate --export-rendered --separator xml --print-rendered-token-count --output-file \"$RENDERED_DEFAULT_XML\""
run_quiet "export rendered (no-dup prompt) + print rendered token counts" bash -c "cd \"$PRESCRIBE_ROOT\" && go run ./cmd/prescribe --repo \"$TEST_REPO_DIR\" --target \"$TARGET_BRANCH\" generate --load-session \"$SESSION_NODUP_YAML\" --export-rendered --separator xml --print-rendered-token-count --output-file \"$RENDERED_NODUP_XML\""

# Extract user CDATA to plain text and count duplication markers.
python3 - "$RENDERED_DEFAULT_XML" "$RENDERED_NODUP_XML" "$USER_DEFAULT_TXT" "$USER_NODUP_TXT" >>"$LOG" 2>&1 <<'PY'
import pathlib, sys, re

def extract_user(xml: str) -> str:
    m = re.search(r"<user><!\[CDATA\[(.*?)\]\]></user>", xml, re.S)
    if not m:
        return ""
    return m.group(1)

default_xml = pathlib.Path(sys.argv[1]).read_text()
nodup_xml = pathlib.Path(sys.argv[2]).read_text()
out_default = pathlib.Path(sys.argv[3])
out_nodup = pathlib.Path(sys.argv[4])

out_default.write_text(extract_user(default_xml))
out_nodup.write_text(extract_user(nodup_xml))
PY

python3 - "$SESSION_SHOW_JSON" "$SESSION_TOKEN_JSON" "$USER_DEFAULT_TXT" "$USER_NODUP_TXT" "$LOG" <<'PY'
import json, pathlib, sys, re

session_show = json.loads(pathlib.Path(sys.argv[1]).read_text())
session_token = json.loads(pathlib.Path(sys.argv[2]).read_text())
user_default = pathlib.Path(sys.argv[3]).read_text()
user_nodup = pathlib.Path(sys.argv[4]).read_text()
log_text = pathlib.Path(sys.argv[5]).read_text()

def rows(x):
    if isinstance(x, list):
        return x
    if isinstance(x, dict) and isinstance(x.get("rows"), list):
        return x["rows"]
    return []

def first_row_with(rs, **conds):
    for r in rs:
        if all(r.get(k) == v for k, v in conds.items()):
            return r
    return {}

ss = rows(session_show)
st = rows(session_token)
ss0 = ss[0] if ss else {}
total = first_row_with(st, kind="total")

marker = "The description of the pull request is:"
marker_default = user_default.count(marker)
marker_nodup = user_nodup.count(marker)

counts_lines = [l.strip() for l in log_text.splitlines() if l.startswith("Rendered payload token counts")]
export_lines = [l.strip() for l in log_text.splitlines() if l.startswith("Rendered payload export token count")]
default_counts = counts_lines[0] if len(counts_lines) > 0 else ""
nodup_counts = counts_lines[1] if len(counts_lines) > 1 else ""
default_export = export_lines[0] if len(export_lines) > 0 else ""
nodup_export = export_lines[1] if len(export_lines) > 1 else ""

print("=== small-repo balloon summary ===")
print(f"session_show.token_count={ss0.get('token_count')}")
print(f"session_token_count.stored_total={total.get('stored_total')} effective_total={total.get('effective_total')} delta={total.get('delta')}")
print(f"dup_marker=\"{marker}\" default_count={marker_default} nodup_count={marker_nodup}")
print(default_counts)
print(default_export)
print(nodup_counts)
print(nodup_export)
print(f"artifacts: {sys.argv[1]} {sys.argv[2]} {sys.argv[3]} {sys.argv[4]}")
print(f"log: {sys.argv[5]}")
PY

echo
echo "done"


