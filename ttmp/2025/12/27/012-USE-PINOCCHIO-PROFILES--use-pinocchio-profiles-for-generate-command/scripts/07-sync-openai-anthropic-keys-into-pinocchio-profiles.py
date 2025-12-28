#!/usr/bin/env python3
"""
Sync OpenAI + Anthropic API keys from a local config YAML into Pinocchio profiles.yaml.

This script is intentionally "reference-safe":
- It does NOT print secret values.
- It only reports presence + key lengths.

Default source:
  /home/manuel/code/mento/moments/config/app/local.yaml

Default destination (XDG):
  ~/.config/pinocchio/profiles.yaml

It can:
- append missing profiles (o4-mini, sonnet-4.5)
- or update existing profile keys if present

Usage:
  python3 scripts/07-sync-openai-anthropic-keys-into-pinocchio-profiles.py
"""

from __future__ import annotations

import argparse
import os
import stat
from pathlib import Path
from typing import Any, Optional

import yaml


def _find_first_string_key(obj: Any, candidates: set[str]) -> Optional[str]:
    if isinstance(obj, dict):
        for k, v in obj.items():
            lk = str(k).lower()
            if lk in candidates and isinstance(v, str) and v.strip():
                return v.strip()
        for v in obj.values():
            r = _find_first_string_key(v, candidates)
            if r:
                return r
    if isinstance(obj, list):
        for it in obj:
            r = _find_first_string_key(it, candidates)
            if r:
                return r
    return None


def _load_yaml(path: Path) -> Any:
    return yaml.safe_load(path.read_text())


def _ensure_profile(doc: dict, profile: str) -> dict:
    if profile not in doc or not isinstance(doc.get(profile), dict):
        doc[profile] = {}
    return doc[profile]


def _ensure_section(profile_doc: dict, section: str) -> dict:
    if section not in profile_doc or not isinstance(profile_doc.get(section), dict):
        profile_doc[section] = {}
    return profile_doc[section]


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument(
        "--local",
        default="/home/manuel/code/mento/moments/config/app/local.yaml",
        help="Path to local.yaml that contains OpenAI/Anthropic keys",
    )
    ap.add_argument(
        "--profiles",
        default=str(Path.home() / ".config" / "pinocchio" / "profiles.yaml"),
        help="Path to pinocchio profiles.yaml to update",
    )
    ap.add_argument(
        "--append-missing-profiles",
        action="store_true",
        help="Append o4-mini and sonnet-4.5 profiles if missing",
    )
    args = ap.parse_args()

    local_path = Path(args.local).expanduser()
    profiles_path = Path(args.profiles).expanduser()

    if not local_path.exists():
        raise SystemExit(f"missing local config: {local_path}")
    if not profiles_path.exists():
        raise SystemExit(f"missing profiles.yaml: {profiles_path}")

    local = _load_yaml(local_path)
    openai_key = _find_first_string_key(
        local,
        {
            "openai_api_key",
            "openai-api-key",
            "openaiapikey",
            "openai_key",
            "openai-key",
        },
    )
    anthropic_key = _find_first_string_key(
        local,
        {
            "anthropic_api_key",
            "anthropic-api-key",
            "anthropicapikey",
            "anthropic_key",
            "anthropic-key",
            "claude_api_key",
            "claude-api-key",
            "claudeapikey",
        },
    )

    if not openai_key:
        raise SystemExit("could not find OpenAI API key in local.yaml (looked for common names)")
    if not anthropic_key:
        raise SystemExit("could not find Anthropic/Claude API key in local.yaml (looked for common names)")

    # Preserve file mode
    mode = stat.S_IMODE(os.stat(profiles_path).st_mode)

    doc = _load_yaml(profiles_path)
    if not isinstance(doc, dict):
        raise SystemExit("profiles.yaml did not parse as a mapping")

    # Update existing profiles (if present) or append if requested.
    updated = []

    def set_openai(profile_name: str) -> None:
        pd = _ensure_profile(doc, profile_name)
        oc = _ensure_section(pd, "openai-chat")
        oc["openai-api-key"] = openai_key
        ai = _ensure_section(pd, "ai-chat")
        ai.setdefault("ai-api-type", "openai")
        ai.setdefault("ai-engine", profile_name)
        updated.append(profile_name)

    def set_claude(profile_name: str) -> None:
        pd = _ensure_profile(doc, profile_name)
        cc = _ensure_section(pd, "claude-chat")
        cc["claude-api-key"] = anthropic_key
        ai = _ensure_section(pd, "ai-chat")
        ai.setdefault("ai-api-type", "claude")
        ai.setdefault("ai-engine", profile_name)
        updated.append(profile_name)

    # Always update if present.
    if "o4-mini" in doc:
        set_openai("o4-mini")
    if "sonnet-4.5" in doc:
        set_claude("sonnet-4.5")

    if args.append_missing_profiles:
        if "o4-mini" not in doc:
            set_openai("o4-mini")
        if "sonnet-4.5" not in doc:
            set_claude("sonnet-4.5")

    if not updated:
        print("No matching profiles updated (use --append-missing-profiles if you want to add o4-mini/sonnet-4.5).")
        return 0

    # Backup next to file (once).
    bak = profiles_path.with_suffix(profiles_path.suffix + ".bak-sync-keys")
    if not bak.exists():
        bak.write_text(profiles_path.read_text())
        os.chmod(bak, mode)

    profiles_path.write_text(yaml.safe_dump(doc, sort_keys=False))
    os.chmod(profiles_path, mode)

    print("Updated profiles:", ", ".join(sorted(set(updated))))
    print(f"profiles_path={profiles_path}")
    print(f"backup={bak}")
    print(f"openai_key_len={len(openai_key)} anthropic_key_len={len(anthropic_key)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())


