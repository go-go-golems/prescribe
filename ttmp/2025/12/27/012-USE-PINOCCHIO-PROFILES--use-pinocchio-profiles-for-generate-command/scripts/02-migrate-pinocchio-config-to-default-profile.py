#!/usr/bin/env python3
"""
Migrate legacy Pinocchio config.yaml (layer-slug -> param map) into a profiles.yaml "default" profile.

Why this exists:
- We want repeatable, editable scripts in the ticket's scripts/ folder (no inline one-liners).
- This script is designed to be safe around secrets: it never prints values, only structure summaries.

What it does:
- Reads ~/.pinocchio/config.yaml (or an explicit path)
- Reads ~/.config/pinocchio/profiles.yaml (or an explicit path)
- Takes all top-level mapping entries from config.yaml (treat them as layer slugs)
- Merges them into profiles.yaml under profile "default" (or a specified profile name)
  - Existing keys in the target profile are overwritten by config.yaml values (migration wins)
- Writes back profiles.yaml after creating a timestamped backup

Important:
- profiles.yaml is used by Glazed profile middleware (profile -> layer -> param -> value).
- Non-mapping top-level keys in config.yaml (like "repositories: [...]") are skipped, because they
  do not match the layer format and would not apply via Glazed layers anyway.
"""

from __future__ import annotations

import argparse
import datetime as dt
import os
import sys
from pathlib import Path
from typing import Any, Dict, Tuple

import yaml


def load_yaml(path: Path) -> Any:
    if not path.exists():
        return {}
    text = path.read_text()
    if text.strip() == "":
        return {}
    return yaml.safe_load(text)


def ensure_mapping(x: Any, what: str) -> Dict[str, Any]:
    if x is None:
        return {}
    if not isinstance(x, dict):
        raise ValueError(f"{what} must be a YAML mapping at the top level, got {type(x).__name__}")
    return x


def extract_layer_maps(config: Dict[str, Any]) -> Tuple[Dict[str, Dict[str, Any]], Dict[str, str]]:
    layers: Dict[str, Dict[str, Any]] = {}
    skipped: Dict[str, str] = {}
    for k, v in config.items():
        if isinstance(v, dict):
            layers[k] = v
        else:
            skipped[k] = type(v).__name__
    return layers, skipped


def deep_merge(dst: Dict[str, Any], src: Dict[str, Any]) -> None:
    """
    Deep-merge src into dst (recursive for dicts). src wins on conflicts.
    """
    for k, v in src.items():
        if isinstance(v, dict) and isinstance(dst.get(k), dict):
            deep_merge(dst[k], v)  # type: ignore[index]
        else:
            dst[k] = v


def atomic_write(path: Path, data: str) -> None:
    tmp = path.with_suffix(path.suffix + ".tmp")
    tmp.write_text(data)
    os.replace(tmp, path)


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--config", default=str(Path.home() / ".pinocchio" / "config.yaml"))
    ap.add_argument("--profiles", default=str(Path.home() / ".config" / "pinocchio" / "profiles.yaml"))
    ap.add_argument("--profile-name", default="default")
    ap.add_argument("--dry-run", action="store_true", help="Show what would change, but do not write")
    ap.add_argument("--no-backup", action="store_true", help="Do not create a backup (not recommended)")
    args = ap.parse_args()

    config_path = Path(args.config).expanduser()
    profiles_path = Path(args.profiles).expanduser()

    cfg_raw = load_yaml(config_path)
    profiles_raw = load_yaml(profiles_path)

    cfg = ensure_mapping(cfg_raw, f"config file {config_path}")
    profiles = ensure_mapping(profiles_raw, f"profiles file {profiles_path}")

    layers_to_merge, skipped = extract_layer_maps(cfg)

    target_profile_name = args.profile_name
    existing_profile = profiles.get(target_profile_name, {})
    existing_profile = ensure_mapping(existing_profile, f"profiles[{target_profile_name!r}]")

    # Prepare merged copy
    merged_profile: Dict[str, Any] = dict(existing_profile)
    for layer_slug, layer_map in layers_to_merge.items():
        existing_layer = merged_profile.get(layer_slug, {})
        existing_layer = ensure_mapping(existing_layer, f"profiles[{target_profile_name!r}][{layer_slug!r}]")
        layer_copy = dict(existing_layer)
        deep_merge(layer_copy, layer_map)
        merged_profile[layer_slug] = layer_copy

    # Summaries (no secrets)
    print("=== migrate pinocchio config -> profiles default ===")
    print(f"config:   {config_path}")
    print(f"profiles: {profiles_path}")
    print(f"profile:  {target_profile_name}")
    print(f"layers_to_merge: {len(layers_to_merge)}")
    if skipped:
        print("skipped_top_level_keys (non-mapping):")
        for k, tname in skipped.items():
            print(f"  - {k}: {tname}")
    else:
        print("skipped_top_level_keys (non-mapping): none")

    if args.dry_run:
        print("\nDRY RUN: not writing anything")
        # Print a compact per-layer key count (no values)
        for layer_slug, layer_map in layers_to_merge.items():
            keys = list(layer_map.keys())
            print(f"- layer {layer_slug}: {len(keys)} keys")
        return 0

    # Backup + write
    original_mode = None
    if profiles_path.exists():
        st = profiles_path.stat()
        original_mode = st.st_mode & 0o777

    if not args.no_backup and profiles_path.exists():
        ts = dt.datetime.now().strftime("%Y%m%d-%H%M%S")
        backup_path = profiles_path.with_suffix(profiles_path.suffix + f".bak-{ts}")
        backup_path.write_bytes(profiles_path.read_bytes())
        if original_mode is not None:
            os.chmod(backup_path, original_mode)
        print(f"backup_written: {backup_path}")

    profiles[target_profile_name] = merged_profile
    out = yaml.safe_dump(profiles, sort_keys=False)
    atomic_write(profiles_path, out)
    if original_mode is not None:
        os.chmod(profiles_path, original_mode)

    print("write: ok")
    print("note: values were not printed (may contain secrets)")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())


