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

ARG DEBIAN_FRONTEND=noninteractive

# Use the base image's existing vscode user (UID 1000). Creating a new
# user would get a different UID, causing permission mismatches with
# bind-mounted host files (Claude config, workspace). The base image's
# vscode user at UID 1000 matches the typical host user.
ARG USERNAME=vscode

# Minimal apt packages — only what Nix and the base image need.
# All development tools come from Nix (see flake.nix).
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl xz-utils \
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

# ---------------------------------------------------------------------------
# Nix — reproducible, declarative package management
# ---------------------------------------------------------------------------
# Install Nix in single-user mode (no daemon, no root needed at runtime).
# The install runs as root during build, then Nix is used as vscode user.
# Single-user mode is simpler and sufficient — multi-user mode requires
# a daemon and build users, which add attack surface for no benefit in
# a single-user container.
RUN mkdir -m 0755 /nix \
    && mkdir -p /etc/nix \
    && printf 'build-users-group =\nexperimental-features = nix-command flakes\n' > /etc/nix/nix.conf \
    && curl -L https://nixos.org/nix/install | bash -s -- --no-daemon \
    && /root/.nix-profile/bin/nix-collect-garbage -d \
    && rm -rf /nix/var/nix/temproots/*

# Copy the flake (package definitions + lock file) into the image.
# flake.nix defines centrally approved package groups.
# flake.lock pins exact versions for reproducibility.
# Both are committed and require PR review to change.
# COPY from the additional build context "flake" (defined in docker-compose.yml).
# This keeps the app build context clean while accessing flake files from
# the parent .devcontainer/ directory.
COPY --from=flake flake.nix flake.lock /tmp/devcontainer-flake/

# Optional package groups, activated via .env (gitignored):
#   PACKAGE_GROUPS=android,python,go
# Only groups defined in flake.nix can be installed. The "base" group
# is always included. All versions pinned via flake.lock.
ARG PACKAGE_GROUPS=""

# Install base + selected package groups via Nix.
# Each group is a separate nix profile install so they can be independently
# managed. The flake is evaluated from the local copy, not fetched from
# the network (the lock file ensures reproducibility).
ENV NIX_BIN=/root/.nix-profile/bin
RUN echo "Installing Nix package groups: base ${PACKAGE_GROUPS}" \
    && $NIX_BIN/nix profile install "/tmp/devcontainer-flake#base" \
    && for group in $(echo "${PACKAGE_GROUPS}" | tr ',' ' '); do \
         [ -z "$group" ] && continue; \
         echo "Installing group: $group"; \
         $NIX_BIN/nix profile install "/tmp/devcontainer-flake#${group}"; \
       done \
    && $NIX_BIN/nix-collect-garbage -d \
    && rm -rf /tmp/devcontainer-flake

# Make Nix store available to the vscode user. Single-user Nix installs
# to /nix owned by the installing user. Change ownership so vscode can
# use nix commands at runtime (e.g., nix shell for ad-hoc packages from
# the flake, nix profile list to see installed groups).
RUN chown -R $USERNAME:$USERNAME /nix /root/.nix-profile \
    && mv /root/.nix-profile /home/$USERNAME/.nix-profile \
    && mv /root/.nix-defexpr /home/$USERNAME/.nix-defexpr 2>/dev/null || true \
    && mv /root/.nix-channels /home/$USERNAME/.nix-channels 2>/dev/null || true

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

# Add Nix profile and Claude Code to PATH for all users/shells.
# The Nix profile bin contains all installed package group binaries.
ENV PATH="/home/$USERNAME/.nix-profile/bin:/home/$USERNAME/.local/bin:${PATH}"

# Claude Code (Bun-based CLI, installed as the non-root user)
USER $USERNAME
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
