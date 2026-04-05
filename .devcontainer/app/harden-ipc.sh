#!/bin/bash
# =============================================================================
# VS Code IPC Escape Mitigation
#
# This script is the SECOND layer of IPC hardening (after remoteEnv in
# devcontainer.json). It exists because VS Code reinjects its IPC
# variables after the remoteEnv clearing — this script catches them.
#
# Sourced at line 1 of .bashrc BEFORE the interactive guard, because
# coding agents invoke bash as non-interactive login shells where
# anything after the "[ -z "$PS1" ] && return" guard never executes.
#
# Must complete quickly — VS Code's env probe times out after 10 seconds.
#
# Attack vectors mitigated:
# - TerminalService RCE via VSCODE_IPC_HOOK_CLI
#   (execute arbitrary commands on host via VS Code's terminal service)
# - Git credential hijacking via VSCODE_GIT_IPC_HANDLE
#   (exfiltrate git credentials from host credential store)
# - GPG agent forwarding
#   (sign commits/messages using host's GPG keys without consent)
# - Electron-as-Node execution
#   (run arbitrary JS on host via ELECTRON_RUN_AS_NODE=1 + VS Code binary)
# - Host browser opening
#   (trigger actions on host via BROWSER env var)
#
# Sources:
# - The Red Guild: Leveraging VS Code Internals to Escape Containers
# - Daniel Demmel: Coding Agents in Secured VS Code Dev Containers
# =============================================================================

# VS Code IPC sockets — primary escape vector.
# These variables point to Unix sockets that allow direct communication
# with the VS Code process running on the HOST. An agent with access to
# these sockets can execute arbitrary commands outside the container.
unset VSCODE_IPC_HOOK_CLI 2>/dev/null
unset VSCODE_GIT_IPC_HANDLE 2>/dev/null

# Git credential helper variables — VS Code injects these to route
# git credential requests through its own helper, which proxies to the
# host's credential store. An agent could use this to exfiltrate tokens.
unset VSCODE_GIT_ASKPASS_NODE 2>/dev/null
unset VSCODE_GIT_ASKPASS_MAIN 2>/dev/null
unset VSCODE_GIT_ASKPASS_EXTRA_ARGS 2>/dev/null
unset GIT_ASKPASS 2>/dev/null

# Electron-as-Node — if set, VS Code's Electron binary acts as a plain
# Node.js runtime, enabling arbitrary script execution on the host.
unset ELECTRON_RUN_AS_NODE 2>/dev/null

# Host browser — prevents container processes from opening URLs in the
# host's browser (minor escape vector, but closes the surface).
unset BROWSER 2>/dev/null

# Git credential helper — VS Code injects this via git system config.
# The empty string set at build time (Dockerfile) overrides it, but
# VS Code may re-inject at the system level. This clears it again.
git config --global --unset credential.helper 2>/dev/null || true

# GPG agent — prevents use of host GPG keys for signing.
unset GPG_AGENT_INFO 2>/dev/null
unset GNUPGHOME 2>/dev/null

# One-time cleanup of existing IPC sockets left in /tmp.
# These are Unix domain sockets that VS Code creates for IPC.
# No background loop — just a single cleanup pass on shell init.
find /tmp -maxdepth 2 -name "vscode-ipc-*" -type s -delete 2>/dev/null || true
find /tmp -maxdepth 2 -name "vscode-git-*" -type s -delete 2>/dev/null || true
