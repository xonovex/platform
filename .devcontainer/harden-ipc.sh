#!/bin/bash
# =============================================================================
# VS Code IPC Escape Mitigation
#
# Mitigates The Red Guild's VS Code container escape vectors:
# - TerminalService RCE via VSCODE_IPC_HOOK_CLI
# - Git credential helper hijacking via VSCODE_GIT_IPC_HANDLE
# - GPG agent forwarding
#
# This script is sourced at line 1 of .bashrc BEFORE the interactive guard.
# It must complete quickly — VS Code's env probe times out after 10 seconds.
#
# Sources:
# - The Red Guild: Leveraging VS Code Internals to Escape Containers
# - Daniel Demmel: Coding Agents in Secured VS Code Dev Containers
# =============================================================================

# Unset VS Code IPC variables that enable container-to-host communication.
# VS Code reinjects these despite remoteEnv clearing — this is the fallback.
unset VSCODE_IPC_HOOK_CLI 2>/dev/null
unset VSCODE_GIT_IPC_HANDLE 2>/dev/null
unset VSCODE_GIT_ASKPASS_NODE 2>/dev/null
unset VSCODE_GIT_ASKPASS_MAIN 2>/dev/null
unset VSCODE_GIT_ASKPASS_EXTRA_ARGS 2>/dev/null
unset GIT_ASKPASS 2>/dev/null

# Remove VS Code credential helpers that could be abused
git config --global --unset credential.helper 2>/dev/null || true

# Kill any forwarded GPG agent sockets
unset GPG_AGENT_INFO 2>/dev/null
unset GNUPGHOME 2>/dev/null

# One-time cleanup of any existing IPC sockets (no background loop)
find /tmp -maxdepth 2 -name "vscode-ipc-*" -type s -delete 2>/dev/null || true
find /tmp -maxdepth 2 -name "vscode-git-*" -type s -delete 2>/dev/null || true
