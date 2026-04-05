# =============================================================================
# App Container — Hardened Development Environment
#
# Runs as non-root (vscode, UID 1000) with zero Linux capabilities,
# no sudo, no setuid binaries, and no-new-privileges. This is the
# container where user code, Claude Code, and development tools run.
# =============================================================================

# SECURITY: Pinned to specific digest to prevent supply chain attacks.
# Floating tags silently change — a compromised tag could inject malware.
# Update digest periodically after verifying the new image.
FROM mcr.microsoft.com/devcontainers/base:ubuntu@sha256:4bcb1b466771b1ba1ea110e2a27daea2f6093f9527fb75ee59703ec89b5561cb

ARG NODE_VERSION=20
ARG DEBIAN_FRONTEND=noninteractive

# System packages. sudo is deliberately excluded — it is purged below.
RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl wget unzip jq vim nano \
    procps build-essential python3 bash \
    && rm -rf /var/lib/apt/lists/*

# Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# SECURITY: Remove sudo entirely. The base image ships with passwordless
# sudo for the vscode user. An agent or compromised tool with sudo access
# could escalate to root, install packages, modify system files, or
# reconfigure the container. SUDO_FORCE_REMOVE=yes overrides Ubuntu's
# safety check that refuses removal without a root password set.
RUN SUDO_FORCE_REMOVE=yes apt-get purge -y sudo \
    && rm -rf /etc/sudoers /etc/sudoers.d \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/*

# Use the base image's existing vscode user (UID 1000). Creating a new
# user would get a different UID, causing permission mismatches with
# bind-mounted host files (Claude config, workspace). The base image's
# vscode user at UID 1000 matches the typical host user.
ARG USERNAME=vscode

# Persistent directories for tools that need writable home paths.
RUN mkdir -p \
    /home/$USERNAME/.npm \
    /home/$USERNAME/.claude \
    && chown -R $USERNAME:$USERNAME /home/$USERNAME

# IPC hardening script (see harden-ipc.sh for details)
COPY harden-ipc.sh /usr/local/bin/harden-ipc.sh
RUN chmod +x /usr/local/bin/harden-ipc.sh

# SECURITY: Source IPC hardening BEFORE .bashrc's interactive guard.
# Coding agents (Claude Code) invoke bash as non-interactive login shells,
# so anything after the "[ -z "$PS1" ] && return" guard in .bashrc never
# runs for agents. Prepending the source ensures it runs for ALL shells.
RUN BASHRC="/home/$USERNAME/.bashrc" \
    && TEMP=$(mktemp) \
    && echo 'source /usr/local/bin/harden-ipc.sh' > "$TEMP" \
    && cat "$BASHRC" >> "$TEMP" \
    && mv "$TEMP" "$BASHRC" \
    && chown $USERNAME:$USERNAME "$BASHRC"

# SECURITY: Pre-clear git credential helpers at build time. VS Code
# injects a credential helper via git system config that proxies to the
# host — this is a container escape vector (can exfiltrate credentials
# or execute commands on the host via the git protocol). Setting an empty
# helper at the user level overrides any system-level helper. The runtime
# harden-ipc.sh repeats this as a fallback since VS Code re-injects it.
USER $USERNAME
RUN git config --global --unset-all credential.helper 2>/dev/null || true \
    && git config --global credential.helper ""
USER root

# Claude Code (Bun-based CLI, installed as the non-root user)
USER $USERNAME
ENV PATH="/home/$USERNAME/.local/bin:${PATH}"
RUN curl -fsSL https://claude.ai/install.sh | bash
USER root

# Moon (monorepo task runner)
RUN curl -fsSL https://moonrepo.dev/install/moon.sh | bash \
    && mv /root/.moon/bin/moon /usr/local/bin/moon \
    || true

# SECURITY: Strip all setuid/setgid bits from every binary in the image.
# Defense-in-depth with no-new-privileges (docker-compose security_opt):
# - no-new-privileges: kernel ignores setuid bits at execve time
# - Stripping bits: removes them from the filesystem entirely
# - Sudo removal: eliminates the primary escalation tool
# Three independent layers ensure privilege escalation is blocked even
# if one mechanism is somehow bypassed.
RUN find / -perm /4000 -type f -exec chmod u-s {} + 2>/dev/null || true \
    && find / -perm /2000 -type f -exec chmod g-s {} + 2>/dev/null || true

USER $USERNAME
WORKDIR /workspaces/xonovex-platform
