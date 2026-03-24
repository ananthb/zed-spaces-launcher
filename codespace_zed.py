#!/usr/bin/env python3

from __future__ import annotations

import argparse
import json
import os
import platform
import re
import subprocess
import sys
from pathlib import Path
from typing import Any


DEFAULT_CONFIG_PATH = "codespace-zed.config.json"
SSH_INCLUDE_LINE = "Include ~/.ssh/codespaces-zed/*.conf"


def main(argv: list[str] | None = None) -> int:
    options = parse_args(argv or sys.argv[1:])
    config_path = Path(options.config).resolve()
    config = parse_launcher_config(read_text(config_path), str(config_path))
    target_name = options.target or config.get("defaultTarget")

    if not target_name:
        raise RuntimeError("No target was provided and config.defaultTarget is not set.")

    targets = config["targets"]
    if target_name not in targets:
        raise RuntimeError(f'Unknown target "{target_name}" in {config_path}.')

    target = targets[target_name]

    require_command("gh")
    require_command("zed")
    ensure_gh_auth()

    codespaces = list_codespaces(target["repository"])
    codespace = select_codespace(codespaces, target, dry_run=options.dry_run)

    if not codespace:
        if options.dry_run:
            raise RuntimeError("No matching codespace exists and --dry-run forbids creating one.")
        codespace = create_codespace(target)

    ensure_codespace_reachable(codespace["name"])

    ssh_config = get_codespace_ssh_config(codespace["name"])
    ssh_alias = parse_primary_host_alias(ssh_config)
    ssh_paths = resolve_ssh_paths()
    ssh_paths["include_dir"].mkdir(parents=True, exist_ok=True)
    ensure_ssh_config_includes_generated_configs(ssh_paths["main_config_path"])
    write_text(ssh_paths["codespace_config_path"](codespace["name"]), ssh_config)

    zed_settings_path = resolve_zed_settings_path()
    upsert_zed_connection(
        settings_path=zed_settings_path,
        connection=build_zed_connection(
            host=ssh_alias,
            workspace_path=target["workspacePath"],
            nickname=target.get("zedNickname")
            or target.get("displayName")
            or codespace.get("displayName")
            or target_name,
            upload_binary_over_ssh=target.get("uploadBinaryOverSsh"),
        ),
    )

    remote_url = f"ssh://{ssh_alias}/{trim_leading_slash(target['workspacePath'])}"
    if options.dry_run or options.no_open:
        print(
            json.dumps(
                {
                    "target": target_name,
                    "codespace": codespace["name"],
                    "sshAlias": ssh_alias,
                    "remoteUrl": remote_url,
                    "zedSettingsPath": str(zed_settings_path),
                },
                indent=2,
            )
        )
        return 0

    run_command(["zed", remote_url], capture_output=False)
    return 0


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(prog="codespace-zed")
    parser.add_argument("target", nargs="?")
    parser.add_argument("--config", default=DEFAULT_CONFIG_PATH)
    parser.add_argument("--no-open", action="store_true")
    parser.add_argument("--dry-run", action="store_true")
    return parser.parse_args(argv)


def parse_launcher_config(source: str, config_path: str) -> dict[str, Any]:
    parsed = parse_jsonc(source, config_path)
    if not isinstance(parsed, dict):
        raise RuntimeError(f"Expected an object in {config_path}.")
    targets = parsed.get("targets")
    if not isinstance(targets, dict):
        raise RuntimeError(f'Expected "targets" object in {config_path}.')
    return parsed


def parse_jsonc(source: str, label: str = "JSON") -> Any:
    without_block_comments = re.sub(r"/\*[\s\S]*?\*/", "", source)
    without_comments = re.sub(r"^\s*//.*$", "", without_block_comments, flags=re.MULTILINE)
    without_trailing_commas = re.sub(r",\s*([}\]])", r"\1", without_comments)
    try:
        return json.loads(without_trailing_commas)
    except json.JSONDecodeError as error:
        raise RuntimeError(f"Could not parse {label}: {error}") from error


def choose_codespace(codespaces: list[dict[str, Any]], target: dict[str, Any]) -> dict[str, Any] | None:
    matches = find_matching_codespaces(codespaces, target)
    if len(matches) > 1:
        names = ", ".join(item["name"] for item in matches)
        raise RuntimeError(f'Ambiguous codespace match for {target["repository"]}: {names}')
    return matches[0] if matches else None


def find_matching_codespaces(codespaces: list[dict[str, Any]], target: dict[str, Any]) -> list[dict[str, Any]]:
    return [codespace for codespace in codespaces if matches_target(codespace, target)]


def select_codespace(
    codespaces: list[dict[str, Any]],
    target: dict[str, Any],
    *,
    dry_run: bool,
    input_fn=input,
    is_tty: bool | None = None,
) -> dict[str, Any] | None:
    if not codespaces:
        return None

    if is_tty is None:
        is_tty = sys.stdin.isatty()

    if not is_tty:
        return choose_codespace(codespaces, target)

    return prompt_for_codespace_choice(codespaces, target, dry_run=dry_run, input_fn=input_fn)


def matches_target(codespace: dict[str, Any], target: dict[str, Any]) -> bool:
    repository = codespace.get("repository")
    if isinstance(repository, dict):
        repository_name = repository.get("full_name") or repository.get("nameWithOwner")
    elif isinstance(repository, str):
        repository_name = repository
    else:
        repository_name = None
    if repository_name != target["repository"]:
        return False

    if target.get("codespaceName") and codespace.get("name") != target["codespaceName"]:
        return False

    if target.get("displayName") and codespace.get("displayName") != target["displayName"]:
        return False

    if target.get("branch"):
        git_status = codespace.get("gitStatus") or {}
        ref = git_status.get("ref") or git_status.get("branch")
        if ref and ref != target["branch"]:
            return False

    return True


def prompt_for_codespace_choice(
    codespaces: list[dict[str, Any]],
    target: dict[str, Any],
    *,
    dry_run: bool,
    input_fn=input,
) -> dict[str, Any] | None:
    matches = find_matching_codespaces(codespaces, target)
    recommended_name = matches[0]["name"] if len(matches) == 1 else None

    print(f'Existing codespaces found for {target["repository"]}:')
    for index, codespace in enumerate(codespaces, start=1):
        print(f'  {index}. {describe_codespace(codespace, recommended=codespace["name"] == recommended_name)}')

    create_label = "create a new codespace"
    if dry_run:
        create_label += " (disabled by --dry-run)"
    print(f"  {len(codespaces) + 1}. {create_label}")

    while True:
        raw = input_fn(f"Select 1-{len(codespaces) + 1}: ").strip()
        if not raw.isdigit():
            print("Enter the number for the codespace you want to use.")
            continue

        selection = int(raw)
        if 1 <= selection <= len(codespaces):
            return codespaces[selection - 1]
        if selection == len(codespaces) + 1:
            if dry_run:
                raise RuntimeError("No matching codespace exists and --dry-run forbids creating one.")
            return None

        print("Selection out of range.")


def describe_codespace(codespace: dict[str, Any], *, recommended: bool) -> str:
    branch = ""
    git_status = codespace.get("gitStatus") or {}
    ref = git_status.get("ref") or git_status.get("branch")
    if ref:
        branch = f", branch={ref}"

    display_name = codespace.get("displayName")
    state = codespace.get("state", "unknown")
    label = codespace["name"]
    if display_name:
        label += f" ({display_name})"
    label += f", state={state}{branch}"
    if recommended:
        label += " [matches config]"
    return label


def build_create_args(target: dict[str, Any]) -> list[str]:
    args = ["gh", "codespace", "create", "--repo", target["repository"]]
    optional_flags = [
        ("branch", "--branch"),
        ("displayName", "--display-name"),
        ("machine", "--machine"),
        ("location", "--location"),
        ("devcontainerPath", "--devcontainer-path"),
        ("idleTimeout", "--idle-timeout"),
        ("retentionPeriod", "--retention-period"),
    ]
    for key, flag in optional_flags:
        value = target.get(key)
        if value:
            args.extend([flag, value])
    return args


def parse_primary_host_alias(ssh_config_text: str) -> str:
    match = re.search(r"^\s*Host\s+([^\s*][^\s]*)\s*$", ssh_config_text, flags=re.MULTILINE)
    if not match:
        raise RuntimeError("Could not find a concrete Host entry in gh codespace ssh --config output.")
    return match.group(1)


def ensure_include_line(config_text: str, include_line: str = SSH_INCLUDE_LINE) -> str:
    lines = [line for line in config_text.splitlines() if line.strip() != include_line]
    body = "\n".join(lines).strip("\n")
    if body:
        return f"{include_line}\n{body}\n"
    return f"{include_line}\n"


def upsert_ssh_connection(settings: dict[str, Any], connection: dict[str, Any]) -> dict[str, Any]:
    next_settings = dict(settings)
    existing_connections = list(next_settings.get("ssh_connections") or [])
    identity = f"{connection.get('username', '')}@{connection['host']}:{connection.get('port', 22)}"

    existing_index = -1
    for index, item in enumerate(existing_connections):
        item_identity = f"{item.get('username', '')}@{item['host']}:{item.get('port', 22)}"
        if item_identity == identity:
            existing_index = index
            break

    if existing_index >= 0:
        merged = dict(existing_connections[existing_index])
        merged.update(connection)
        existing_connections[existing_index] = merged
    else:
        existing_connections.append(connection)

    next_settings["ssh_connections"] = existing_connections
    return next_settings


def build_zed_connection(
    *,
    host: str,
    workspace_path: str,
    nickname: str,
    upload_binary_over_ssh: bool | None = None,
) -> dict[str, Any]:
    connection: dict[str, Any] = {
        "host": host,
        "nickname": nickname,
        "projects": [{"paths": [workspace_path]}],
    }
    if upload_binary_over_ssh is not None:
        connection["upload_binary_over_ssh"] = upload_binary_over_ssh
    return connection


def upsert_zed_connection(*, settings_path: Path, connection: dict[str, Any]) -> None:
    existing_raw = read_text(settings_path, allow_missing=True)
    current = parse_jsonc(existing_raw, str(settings_path)) if existing_raw else {}
    if not isinstance(current, dict):
        raise RuntimeError(f"Expected object settings in {settings_path}.")
    updated = upsert_ssh_connection(current, connection)
    settings_path.parent.mkdir(parents=True, exist_ok=True)
    write_text(settings_path, json.dumps(updated, indent=2) + "\n")


def ensure_ssh_config_includes_generated_configs(main_config_path: Path) -> None:
    current = read_text(main_config_path, allow_missing=True) or ""
    updated = ensure_include_line(current)
    main_config_path.parent.mkdir(parents=True, exist_ok=True)
    write_text(main_config_path, updated)


def require_command(command: str) -> None:
    run_command(["which", command])


def ensure_gh_auth() -> None:
    try:
        run_command(["gh", "auth", "status"], capture_output=True)
    except RuntimeError as error:
        raise RuntimeError("GitHub CLI is not authenticated. Run `gh auth login` first.") from error


def list_codespaces(repository: str) -> list[dict[str, Any]]:
    stdout = run_command(
        [
            "gh",
            "codespace",
            "list",
            "--repo",
            repository,
            "--json",
            "name,displayName,repository,state,gitStatus",
        ]
    )
    parsed = parse_jsonc(stdout, "gh codespace list output")
    if not isinstance(parsed, list):
        raise RuntimeError("Expected list output from gh codespace list.")
    return parsed


def create_codespace(target: dict[str, Any]) -> dict[str, Any]:
    combined = run_command(build_create_args(target), merge_streams=True)
    match = re.search(r"[A-Za-z0-9-]+-[A-Za-z0-9]{6,}", combined)
    if not match:
        raise RuntimeError(f"Codespace created but name could not be determined from gh output:\n{combined.strip()}")

    stdout = run_command(
        [
            "gh",
            "codespace",
            "view",
            "--codespace",
            match.group(0),
            "--json",
            "name,displayName,repository,state,gitStatus",
        ]
    )
    parsed = parse_jsonc(stdout, "gh codespace view output")
    if not isinstance(parsed, dict):
        raise RuntimeError("Expected object output from gh codespace view.")
    return parsed


def ensure_codespace_reachable(codespace_name: str) -> None:
    try:
        run_command(
            ["gh", "codespace", "ssh", "--codespace", codespace_name, "--", "true"],
            capture_output=True,
        )
    except RuntimeError as error:
        raise RuntimeError(
            f'Could not start or SSH into codespace "{codespace_name}". '
            f"Ensure the codespace has an SSH server. Original error: {error}"
        ) from error


def get_codespace_ssh_config(codespace_name: str) -> str:
    return run_command(["gh", "codespace", "ssh", "--codespace", codespace_name, "--config"])


def resolve_ssh_paths() -> dict[str, Any]:
    ssh_dir = Path.home() / ".ssh"
    include_dir = ssh_dir / "codespaces-zed"
    return {
        "include_dir": include_dir,
        "main_config_path": ssh_dir / "config",
        "codespace_config_path": lambda codespace_name: include_dir / f"{codespace_name}.conf",
    }


def resolve_zed_settings_path() -> Path:
    if platform.system() == "Darwin":
        return Path.home() / ".zed" / "settings.json"
    return Path.home() / ".config" / "zed" / "settings.json"


def read_text(path: Path, allow_missing: bool = False) -> str | None:
    try:
        return path.read_text(encoding="utf-8")
    except FileNotFoundError:
        if allow_missing:
            return None
        raise


def write_text(path: Path, content: str) -> None:
    path.write_text(content, encoding="utf-8")


def trim_leading_slash(value: str) -> str:
    return value.lstrip("/")


def normalize_trailing_newline(value: str) -> str:
    return value if value.endswith("\n") else value + "\n"


def run_command(
    args: list[str],
    *,
    capture_output: bool = True,
    merge_streams: bool = False,
) -> str:
    stdout = subprocess.PIPE if capture_output else None
    if not capture_output:
        stderr = None
    elif merge_streams:
        stderr = subprocess.STDOUT
    else:
        stderr = subprocess.PIPE
    result = subprocess.run(
        args,
        check=False,
        stdout=stdout,
        text=True,
        stderr=stderr,
    )
    if result.returncode != 0:
        detail_parts = []
        if result.stdout:
            detail_parts.append(result.stdout.strip())
        if not merge_streams and result.stderr:
            detail_parts.append(result.stderr.strip())
        detail = "\n".join(part for part in detail_parts if part)
        suffix = f":\n{detail}" if detail else ""
        raise RuntimeError(f'{" ".join(args)} exited with code {result.returncode}{suffix}')
    return result.stdout or ""


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except RuntimeError as error:
        print(f"Error: {error}", file=sys.stderr)
        raise SystemExit(1)
