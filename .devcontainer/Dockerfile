FROM mcr.microsoft.com/devcontainers/base:ubuntu

ARG NODE_VERSION=20
ARG DEBIAN_FRONTEND=noninteractive

# System packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    git curl wget unzip jq vim nano \
    procps build-essential python3 sudo bash \
    && rm -rf /var/lib/apt/lists/*

# Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

# Use the base image's existing vscode user (UID 1000) to match host file ownership
ARG USERNAME=vscode

# Persistent directories
RUN mkdir -p \
    /home/$USERNAME/.npm \
    /home/$USERNAME/.claude \
    && chown -R $USERNAME:$USERNAME /home/$USERNAME

# IPC hardening script
COPY harden-ipc.sh /usr/local/bin/harden-ipc.sh
RUN chmod +x /usr/local/bin/harden-ipc.sh

# Source IPC hardening in .bashrc BEFORE the interactive guard
RUN BASHRC="/home/$USERNAME/.bashrc" \
    && TEMP=$(mktemp) \
    && echo 'source /usr/local/bin/harden-ipc.sh' > "$TEMP" \
    && cat "$BASHRC" >> "$TEMP" \
    && mv "$TEMP" "$BASHRC" \
    && chown $USERNAME:$USERNAME "$BASHRC"

# Claude Code
USER $USERNAME
ENV PATH="/home/$USERNAME/.local/bin:${PATH}"
RUN curl -fsSL https://claude.ai/install.sh | bash
USER root

# Moon (monorepo task runner)
RUN curl -fsSL https://moonrepo.dev/install/moon.sh | bash \
    && mv /root/.moon/bin/moon /usr/local/bin/moon \
    || true

USER $USERNAME
WORKDIR /workspaces/xonovex-platform
