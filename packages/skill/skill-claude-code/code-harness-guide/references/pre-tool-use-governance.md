# Governance PreToolUse Hook

Register the deterministic governance handler in project scope by copying `assets/pre-tool-use-settings.json` to `.claude/settings.json` and `scripts/governance-pre-tool-use.sh` to `.claude/hooks/governance-pre-tool-use.sh`. Review any existing settings and merge only the owned `PreToolUse` group instead of replacing unrelated hooks.

The handler maps `Edit` and `Write` events to the versioned governance decision-service contract. It sends only the session/tool correlation identifiers, tool name, and target path to the pod-local or explicitly configured decision endpoint. It then records the minimized enforcement outcome using the same correlation identifier; successful calls do not carry a failure code. A deny exits `2`, so Claude Code blocks the tool call before permission evaluation. An unavailable mandatory decision or enforcement recorder also exits `2`; advisory mode reports the outage and exits `0`.

Run `scripts/refresh-pre-tool-use-probe.sh` from a credentialed maintainer environment. The probe creates an isolated temporary project, starts the decision service, registers the real project hook, runs pinned Claude Code `2.1.211`, proves a protected write is absent and an ordinary write is present, verifies both decisions in the shared evidence JSONL, and requires the paired verdict/enforcement telemetry records. Set `PROBE_OUTPUT_DIR` to retain the raw CLI outputs, service evidence, and telemetry for review.

Remove only the owned `Edit|Write` hook group and handler to roll back. Confirm the group disappears from `/hooks`; other project, user, plugin, or managed handlers remain untouched.
