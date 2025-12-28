#!/usr/bin/env python3
"""
Extract a small human-readable snippet for a given parameter from Glazed's
`--print-parsed-parameters` output (YAML-ish).

Why this exists:
- We want reusable debugging helpers under the ticket's scripts/ folder.
- It avoids scattering inline python one-liners across terminal history and bash scripts.

Usage:
  ./extract-parsed-parameter-snippet.py --file /tmp/out.txt --param separator
  ./extract-parsed-parameter-snippet.py --file /tmp/out.txt --param separator --require-source profiles
"""

from __future__ import annotations

import argparse
import sys
from pathlib import Path


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--file", required=True, help="Path to --print-parsed-parameters output file")
    ap.add_argument(
        "--param",
        required=True,
        help="Parameter key to search for (e.g. 'separator'). Matches lines like 'separator:'",
    )
    ap.add_argument(
        "--context",
        type=int,
        default=10,
        help="How many lines of context to show before/after the match (default: 10)",
    )
    ap.add_argument(
        "--require-source",
        default="",
        help="If set, require this parse-step source to appear within the param block (e.g. 'profiles')",
    )
    args = ap.parse_args()

    p = Path(args.file)
    if not p.exists():
        print(f"ERROR: file not found: {p}", file=sys.stderr)
        return 2

    lines = p.read_text().splitlines()
    needle = f"{args.param}:"
    idxs = [i for i, l in enumerate(lines) if l.strip() == needle]

    if not idxs:
        print(f"ERROR: could not find param '{args.param}' (line '{needle}') in {p}", file=sys.stderr)
        return 1

    # Print snippet around the first match.
    i = idxs[0]
    lo = max(0, i - args.context)
    hi = min(len(lines), i + args.context + 1)

    snippet = lines[lo:hi]
    print("\n".join(snippet))

    if args.require_source:
        # Heuristic: scan forward while indentation is >= the param line indentation.
        base_indent = len(lines[i]) - len(lines[i].lstrip(" "))
        found = False
        for j in range(i + 1, len(lines)):
            indent = len(lines[j]) - len(lines[j].lstrip(" "))
            if lines[j].strip() == "":
                continue
            if indent <= base_indent:
                break
            if lines[j].strip() == f"source: {args.require_source}":
                found = True
                break
        if not found:
            print(
                f"\nERROR: did not find required source '{args.require_source}' inside '{args.param}' block",
                file=sys.stderr,
            )
            return 1

    return 0


if __name__ == "__main__":
    raise SystemExit(main())


