#!/bin/bash
# =============================================================================
# App Container Entrypoint
#
# 1. Copies read-only ~/.claude.json to a writable location.
# 2. Runs background IPC socket cleanup loop.
# =============================================================================

# ---------------------------------------------------------------------------
# Claude config: read-only host mount → writable ephemeral copy
# ---------------------------------------------------------------------------
# ~/.claude.json contains MCP server configs, per-project trust decisions,
# and OAuth tokens. It is bind-mounted read-only as .claude.json.ro to
# prevent the agent from persisting malicious MCP servers or trust changes
# to the host.
#
# Claude Code crashes if ~/.claude.json is directly read-only — it needs
# to write session state (OAuth refresh, UI preferences, metrics) during
# normal operation. The copy is writable but ephemeral: changes are lost
# on container restart and never reach the host.
CLAUDE_JSON_RO="/home/vscode/.claude.json.ro"
CLAUDE_JSON="/home/vscode/.claude.json"

if [ -f "$CLAUDE_JSON_RO" ]; then
  cp "$CLAUDE_JSON_RO" "$CLAUDE_JSON"
fi

# ---------------------------------------------------------------------------
# IPC socket cleanup loop
# ---------------------------------------------------------------------------
# VS Code creates IPC sockets at multiple lifecycle stages — a single
# cleanup pass in harden.sh misses sockets created after shell init.
# This loop runs as a child of PID 1, so it survives VS Code's
# cgroup-based process termination. 10 passes over 5 minutes catches
# late-created sockets.
(for i in $(seq 1 10); do
  find /tmp -maxdepth 2 \( \
    -name "vscode-ipc-*" \
    -o -name "vscode-git-*" \
    -o -name "vscode-ssh-auth-*" \
    -o -name "vscode-remote-containers-*" \
  \) -type s -delete 2>/dev/null
  sleep 30
done) &

exec sleep infinity
