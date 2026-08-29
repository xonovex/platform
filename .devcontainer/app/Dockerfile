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

SHELL ["/bin/bash", "-o", "pipefail", "-c"]

ARG DEBIAN_FRONTEND=noninteractive

# Use the base image's existing vscode user (UID 1000). Creating a new
# user would get a different UID, causing permission mismatches with
# bind-mounted host files (Claude config, workspace). The base image's
# vscode user at UID 1000 matches the typical host user.
ARG USERNAME=vscode

ARG NIX_VERSION=2.34.7
ARG NIX_ARCHIVE_SHA256=eafe5042404e818505e28c5ca3d0885f3ec45c31f955489a25bb38258f87560e

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
    && curl --fail --location --show-error \
      "https://releases.nixos.org/nix/nix-${NIX_VERSION}/nix-${NIX_VERSION}-x86_64-linux.tar.xz" \
      --output /tmp/nix.tar.xz \
    && printf '%s  %s\n' "$NIX_ARCHIVE_SHA256" /tmp/nix.tar.xz | sha256sum --check --strict \
    && tar -xJf /tmp/nix.tar.xz -C /tmp \
    && "/tmp/nix-${NIX_VERSION}-x86_64-linux/install" --no-daemon \
    && /root/.nix-profile/bin/nix-collect-garbage -d \
    && rm -rf /nix/var/nix/temproots/* "/tmp/nix-${NIX_VERSION}-x86_64-linux" /tmp/nix.tar.xz

# Copy the flake (package definitions + lock file) into the image.
# flake.nix defines centrally approved package groups.
# flake.lock pins exact versions for reproducibility.
# Both are committed and require PR review to change.
COPY flake.nix flake.lock /tmp/devcontainer-flake/

# Package group allowlist — determines which flake.nix groups are installed.
# Developers add groups via PR to allowed-packages.conf, same as domains
# and ports. All versions pinned via flake.lock.
COPY allowed-packages.conf /tmp/allowed-packages.conf

# Install package groups listed in allowed-packages.conf.
# Each group is a separate nix profile install so they can be independently
# managed. The flake is evaluated from the local copy, not fetched from
# the network (the lock file ensures reproducibility).
ARG NIX_BIN=/root/.nix-profile/bin
RUN while IFS= read -r line || [ -n "$line" ]; do \
      line=$(echo "$line" | sed 's/#.*//' | tr -d '[:space:]'); \
      [ -z "$line" ] && continue; \
      echo "Installing group: $line"; \
      $NIX_BIN/nix profile add "/tmp/devcontainer-flake#${line}"; \
    done < /tmp/allowed-packages.conf \
    && $NIX_BIN/nix-collect-garbage -d \
    && rm -rf /tmp/devcontainer-flake /tmp/allowed-packages.conf

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

# Shell hardening modules — each .sh file in harden.d/ mitigates a
# specific container escape vector. harden.sh sources them all.
# Every .sh file in harden.d/ is loaded during shell initialization.
COPY harden.sh /usr/local/bin/harden.sh
COPY harden.d/ /usr/local/lib/harden.d/
RUN chmod +x /usr/local/bin/harden.sh

# SECURITY: Source hardening BEFORE .bashrc's interactive guard.
# Coding agents (Claude Code) invoke bash as non-interactive login shells,
# so anything after the "[ -z "$PS1" ] && return" guard in .bashrc never
# runs for agents. Prepending the source ensures it runs for ALL shells.
RUN BASHRC="/home/$USERNAME/.bashrc" \
    && TEMP=$(mktemp) \
    && echo 'source /usr/local/bin/harden.sh' > "$TEMP" \
    && cat "$BASHRC" >> "$TEMP" \
    && mv "$TEMP" "$BASHRC" \
    && chown $USERNAME:$USERNAME "$BASHRC"

# SECURITY: Pre-clear git credential helpers at build time. VS Code
# injects a credential helper via git system config that proxies to the
# host — this is a container escape vector (can exfiltrate credentials
# or execute commands on the host via the git protocol). Setting an empty
# helper at the user level overrides any system-level helper. The runtime
# git-credentials.sh repeats this as a fallback since VS Code re-injects it.
USER $USERNAME
RUN git config --global --unset-all credential.helper 2>/dev/null || true \
    && git config --global credential.helper ""
USER root

# Add Nix profile and Claude Code to PATH for all users/shells.
# The Nix profile bin contains all installed package group binaries.
ENV PATH="/home/$USERNAME/.nix-profile/bin:/home/$USERNAME/.local/bin:${PATH}"


# Claude Code and Moon are installed from an integrity-locked npm dependency tree.
COPY package.json package-lock.json /home/$USERNAME/.local/lib/devcontainer-tools/
RUN chown -R $USERNAME:$USERNAME /home/$USERNAME/.local
USER $USERNAME
RUN npm ci --prefix /home/$USERNAME/.local/lib/devcontainer-tools --omit=dev \
    && mkdir -p /home/$USERNAME/.local/bin \
    && ln -s /home/$USERNAME/.local/lib/devcontainer-tools/node_modules/.bin/claude /home/$USERNAME/.local/bin/claude \
    && ln -s /home/$USERNAME/.local/lib/devcontainer-tools/node_modules/.bin/moon /home/$USERNAME/.local/bin/moon
USER root

# Entrypoint: background IPC socket cleanup loop + sleep infinity.
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

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

CMD ["/usr/local/bin/entrypoint.sh"]
