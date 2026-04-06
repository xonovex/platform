#!/bin/bash
# =============================================================================
# App Container Entrypoint — IPC Socket Cleanup
#
# VS Code creates IPC sockets at multiple lifecycle stages — a single
# cleanup pass in harden-ipc.sh misses sockets created after shell init.
# This loop runs as a child of PID 1, so it survives VS Code's
# cgroup-based process termination. 10 passes over 5 minutes catches
# late-created sockets.
# =============================================================================

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
