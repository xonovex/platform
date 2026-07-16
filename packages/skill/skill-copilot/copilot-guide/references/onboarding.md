# GitHub Copilot Onboarding

## Discover

- Probe `copilot --version`, account/organization policy, surface, operating system, repository trust, and cloud-agent availability.
- Inspect CLI policy directories, user hook directory, repository hook files, repository/user settings, installed plugins, skills, custom agents, and `/env` or equivalent diagnostics.
- For cloud, inspect only repository hook files plus job firewall and ephemeral-environment assumptions.

## Preview

Show version-1 JSON and the exact native source: policy, user, project, inline settings, plugin, or cloud repository. Report handler type, event, matcher, bash/PowerShell/command selection, URL and allowed environment variables, timeout, permissions, files, network/firewall, tokens and content exposure, merge behavior, trust, evidence, verification, disable, and rollback.

A plugin preview pins its manifest/source/version and enumerates hooks, agents, skills, extensions, and MCP sources. A skill preview treats scripts and `allowed-tools` as executable authority, not documentation.

## Apply and verify

Apply after explicit authorization. Verify the loaded source and surface, then probe decision merge, exit code, malformed output, timeout, agent-stop continuation, repository trust, policy non-disable, cloud bash behavior, and restricted network. Do not use a CLI result as cloud evidence or vice versa.

## Lifecycle

- File-level `disableAllHooks` disables only that file on both surfaces.
- Repository settings can disable broader CLI hook sources but cannot disable policy hooks and are not loaded by cloud.
- Roll back the owned file/plugin/skill plus any external endpoint or firewall change.
- Drift includes CLI version, cloud surface behavior, schema, policy source, repository trust, plugin digest, handler type, URL/env allowlists, firewall, permissions, and data flow.
