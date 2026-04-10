import unittest

from codespace_zed import (
    build_display_name,
    build_create_args,
    build_zed_connection,
    choose_codespace,
    describe_codespace,
    ensure_include_line,
    matches_target,
    parse_args,
    parse_jsonc,
    parse_primary_host_alias,
    prepare_target_for_creation,
    prompt_for_codespace_choice,
    select_codespace,
    slugify_work_label,
    upsert_ssh_connection,
)


class CodespaceZedTests(unittest.TestCase):
    def test_parse_args_reads_target_and_flags(self) -> None:
        actual = parse_args(["demo", "--config", "custom.json", "--no-open", "--dry-run"])
        self.assertEqual(actual.config, "custom.json")
        self.assertTrue(actual.no_open)
        self.assertTrue(actual.dry_run)
        self.assertEqual(actual.target, "demo")

    def test_parse_jsonc_accepts_comments_and_trailing_commas(self) -> None:
        actual = parse_jsonc(
            """
            {
              // comment
              "name": "demo",
              "nested": {
                "enabled": true,
              },
            }
            """
        )
        self.assertEqual(actual, {"name": "demo", "nested": {"enabled": True}})

    def test_matches_target_compares_repo_branch_and_display_name(self) -> None:
        codespace = {
            "displayName": "demo-main",
            "gitStatus": {"ref": "main"},
            "name": "demo-abc123",
            "repository": {"full_name": "acme/demo"},
        }
        self.assertTrue(
            matches_target(
                codespace,
                {"branch": "main", "displayName": "demo-main", "repository": "acme/demo"},
            )
        )
        self.assertFalse(matches_target(codespace, {"branch": "develop", "repository": "acme/demo"}))

    def test_matches_target_accepts_string_repository_shape(self) -> None:
        codespace = {
            "displayName": "demo-main",
            "gitStatus": {"ref": "main"},
            "name": "demo-abc123",
            "repository": "acme/demo",
        }
        self.assertTrue(matches_target(codespace, {"branch": "main", "repository": "acme/demo"}))

    def test_select_codespace_uses_prompt_when_interactive(self) -> None:
        codespaces = [
            {"name": "demo-a", "displayName": "one", "repository": "acme/demo", "state": "Available"},
            {"name": "demo-b", "displayName": "two", "repository": "acme/demo", "state": "Shutdown"},
        ]
        selected = select_codespace(
            codespaces,
            {"repository": "acme/demo"},
            dry_run=False,
            input_fn=lambda _: "2",
            is_tty=True,
        )
        self.assertEqual(selected["name"], "demo-b")

    def test_prompt_for_codespace_choice_can_create_new(self) -> None:
        codespaces = [{"name": "demo-a", "repository": "acme/demo", "state": "Available"}]
        selected = prompt_for_codespace_choice(
            codespaces,
            {"repository": "acme/demo"},
            dry_run=False,
            input_fn=lambda _: "2",
        )
        self.assertIsNone(selected)

    def test_describe_codespace_marks_matching_entry(self) -> None:
        description = describe_codespace(
            {
                "name": "demo-a",
                "displayName": "rpcpool-main",
                "gitStatus": {"ref": "main"},
                "state": "Available",
            },
            recommended=True,
        )
        self.assertIn("[matches config]", description)
        self.assertIn("branch=main", description)

    def test_choose_codespace_throws_on_ambiguous_matches(self) -> None:
        codespaces = [
            {"name": "demo-a", "repository": {"full_name": "acme/demo"}},
            {"name": "demo-b", "repository": {"full_name": "acme/demo"}},
        ]
        with self.assertRaisesRegex(RuntimeError, "Ambiguous"):
            choose_codespace(codespaces, {"repository": "acme/demo"})

    def test_build_create_args_maps_optional_fields(self) -> None:
        self.assertEqual(
            build_create_args(
                {
                    "branch": "main",
                    "devcontainerPath": ".devcontainer/devcontainer.json",
                    "displayName": "demo-main",
                    "idleTimeout": "30m",
                    "location": "EastUs",
                    "machine": "standardLinux32gb",
                    "repository": "acme/demo",
                    "retentionPeriod": "72h",
                }
            ),
            [
                "gh",
                "codespace",
                "create",
                "--repo",
                "acme/demo",
                "--branch",
                "main",
                "--display-name",
                "demo-main",
                "--machine",
                "standardLinux32gb",
                "--location",
                "EastUs",
                "--devcontainer-path",
                ".devcontainer/devcontainer.json",
                "--idle-timeout",
                "30m",
                "--retention-period",
                "72h",
            ],
        )

    def test_prepare_target_for_creation_updates_display_name(self) -> None:
        target = {
            "repository": "rpcpool/rpcpool",
            "branch": "main",
            "displayName": "rpcpool-main",
        }
        prepared = prepare_target_for_creation(
            target,
            input_fn=lambda _: "indexer cleanup",
            is_tty=True,
        )
        self.assertEqual(prepared["displayName"], "rpcpool-main-indexer-cleanup")
        self.assertEqual(target["displayName"], "rpcpool-main")

    def test_build_display_name_truncates_to_codespaces_limit(self) -> None:
        display_name = build_display_name(
            repository="rpcpool/rpcpool",
            branch="main",
            work_label="a" * 80,
        )
        self.assertLessEqual(len(display_name), 48)
        self.assertTrue(display_name.startswith("rpcpool-main-"))

    def test_slugify_work_label_normalizes_text(self) -> None:
        self.assertEqual(slugify_work_label("Fix RPC / Health Checks"), "fix-rpc-health-checks")

    def test_parse_primary_host_alias_returns_first_concrete_host(self) -> None:
        ssh_config = """
Host cs-demo
  HostName github.com
  User git
"""
        self.assertEqual(parse_primary_host_alias(ssh_config), "cs-demo")

    def test_ensure_include_line_is_idempotent(self) -> None:
        once = ensure_include_line("Host example\n  HostName example.com\n")
        twice = ensure_include_line(once)
        self.assertEqual(once, twice)
        self.assertTrue(once.startswith("Include ~/.ssh/codespaces-zed/*.conf\n"))

    def test_ensure_include_line_moves_existing_include_to_top(self) -> None:
        config = "Host *\n  IdentityAgent test\nInclude ~/.ssh/codespaces-zed/*.conf\n"
        updated = ensure_include_line(config)
        self.assertTrue(updated.startswith("Include ~/.ssh/codespaces-zed/*.conf\n"))
        self.assertEqual(updated.count("Include ~/.ssh/codespaces-zed/*.conf"), 1)

    def test_upsert_ssh_connection_merges_by_host_identity(self) -> None:
        updated = upsert_ssh_connection(
            {
                "theme": "One Dark",
                "ssh_connections": [
                    {"host": "cs-demo", "nickname": "Old", "projects": [{"paths": ["/old"]}]}
                ],
            },
            build_zed_connection(host="cs-demo", nickname="New", workspace_path="/workspaces/demo"),
        )
        self.assertEqual(updated["theme"], "One Dark")
        self.assertEqual(len(updated["ssh_connections"]), 1)
        self.assertEqual(updated["ssh_connections"][0]["nickname"], "New")
        self.assertEqual(updated["ssh_connections"][0]["projects"], [{"paths": ["/workspaces/demo"]}])


if __name__ == "__main__":
    unittest.main()
