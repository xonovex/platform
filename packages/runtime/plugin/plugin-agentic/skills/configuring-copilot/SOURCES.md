# Sources

Product-specific facts are pinned to the documentation snapshot below. The local runtime probe for this implementation returned `not-installed`; documentation conformance is not runtime conformance.

## Copilot hooks

- **URLs:**
  - https://docs.github.com/en/copilot/reference/hooks-reference
  - https://docs.github.com/en/copilot/concepts/agents/hooks
  - https://docs.github.com/en/copilot/tutorials/copilot-cli-hooks
- **Last reviewed:** 2026-07-16
- **Used for:** `capabilities.md`, `onboarding.md`, `patterns.md`
- **References:** references/capabilities.md, references/onboarding.md, references/patterns.md
- **Aspects extracted:** CLI/cloud events, scopes, policy hooks, command/HTTP/prompt handlers, decision output, exit behavior, cloud sandbox and firewall, disable behavior, and security guidance.

## Copilot plugins and skills

- **URLs:**
  - https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-plugin-reference
  - https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/plugins-creating
  - https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills
  - https://docs.github.com/en/copilot/concepts/agents/about-agent-skills
- **Last reviewed:** 2026-07-16
- **Used for:** `onboarding.md`, `patterns.md`
- **References:** references/onboarding.md, references/patterns.md
- **Aspects extracted:** Plugin manifests, packaged hooks/skills/agents/extensions, load precedence, install/update/remove commands, skill locations, scripts, allowed tools, and trust warnings.

## Refresh workflow

1. Re-fetch the official pages above and compare event, handler, configuration, trust, and lifecycle behavior.
2. Run the documented version probe and replace `not-installed` only with the observed version.
3. Execute every named deterministic probe on the observed surface; keep unsupported and experimental results explicit.
4. Update the matrix version and review date when behavior changes.

## Guide-level synthesis

- **Provenance:** Repository-original integration of the source blocks above; these references combine multiple inputs or maintained conventions rather than one exclusive upstream
- **References:** references/capabilities.md, references/onboarding.md, references/patterns.md
- **Last reviewed:** 2026-07-16
