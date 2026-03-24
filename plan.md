
## 2026-03-23 - GitHub Codespaces Zed Launcher Tool

### Goal
- Add a local developer tool that can:
- Start an existing predefined GitHub Codespace or create a new one.
- Update Zed configuration so the Codespace is available as a remote target.
- Launch Zed connected to that Codespace.

### Assumptions
- The tool will be intended for the local developer machine, not the deployed app.
- GitHub CLI (`gh`) and Zed CLI (`zed`) are the preferred integrations.
- The user is already authenticated with GitHub CLI and has Zed installed locally.
- Codespace selection can be driven by a small local config file or predefined defaults committed in the repo.

### Proposed Implementation
1. Build a standalone tool in this directory
2. Define a small configuration format for named Codespace targets:
- repository
- branch or ref
- machine/region/devcontainer options when creating a new Codespace
- workspace path or display name used for Zed integration
3. Implement the script flow:
- read configured target or CLI arguments
- detect matching existing Codespaces with `gh codespace list`
- if a matching Codespace exists, start it when needed
- otherwise create a new Codespace with `gh codespace create`
- fetch connection details needed by Zed
- update Zed remote/collaboration config idempotently
- invoke `zed` to open the remote workspace
4. Add guardrails and clear errors for:
- missing `gh` or `zed`
- missing GitHub auth
- ambiguous Codespace matches
- Zed config format mismatch
5. Document usage in `README.md` or a dedicated tooling doc.

### Validation Plan
- Verify the script resolves an existing Codespace path without creating duplicates.
- Verify the script can create a new Codespace from configuration.
- Verify repeated runs are idempotent with respect to Zed config updates.
- Smoke-test that the final command launches Zed with the expected remote target.

### Deliverables
- Tooling script and any supporting config/types.
- Usage documentation.
- One dedicated commit for this feature, per repository instructions.
- Add parser tests with saved HTML fixtures for each adapter.
