# AI Agent
# Containerized environment for AI coding agents (Claude Code, OpenCode)
#
# Build: docker build -t ai-agent -f packages/ai/ai-agent/Dockerfile .
# Usage: docker compose -f stacks/ai-agent.yaml up ai-agent

FROM debian:trixie-slim

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    7zip \
    apt-transport-https \
    autoconf \
    automake \
    build-essential \
    bzip2 \
    ca-certificates \
    clang \
    clangd \
    cmake \
    curl \
    dnsutils \
    fd-find \
    ffmpeg \
    file \
    gdb \
    git \
    graphviz \
    git-lfs \
    gnupg \
    golang-go \
    htop \
    imagemagick \
    iputils-ping \
    jq \
    less \
    libcurl4-openssl-dev \
    libffi-dev \
    libpq-dev \
    libreadline-dev \
    libsqlite3-dev \
    libssl-dev \
    libtool \
    locales \
    lsb-release \
    lsof \
    lua5.4 \
    liblua5.4-dev \
    luajit \
    libluajit-5.1-dev \
    luarocks \
    ncurses-base \
    net-tools \
    netcat-openbsd \
    neovim \
    ninja-build \
    openssh-client \
    patch \
    pkg-config \
    postgresql-client \
    procps \
    pipx \
    python3 \
    python3-dev \
    python3-pip \
    python3-venv \
    redis-tools \
    ripgrep \
    shellcheck \
    sox \
    sqlite3 \
    strace \
    tzdata \
    unzip \
    wget \
    xz-utils \
    yq \
    zip \
    && rm -rf /var/lib/apt/lists/*

# Install Docker CLI from official repository
RUN install -m 0755 -d /etc/apt/keyrings \
    && curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc \
    && chmod a+r /etc/apt/keyrings/docker.asc \
    && printf '%s\n' \
        "Types: deb" \
        "URIs: https://download.docker.com/linux/debian" \
        "Suites: trixie" \
        "Components: stable" \
        "Signed-By: /etc/apt/keyrings/docker.asc" \
        > /etc/apt/sources.list.d/docker.sources \
    && apt-get update \
    && apt-get install -y --no-install-recommends docker-ce-cli docker-compose-plugin \
    && rm -rf /var/lib/apt/lists/*

# Install OpenTofu
RUN curl --proto '=https' --tlsv1.2 -fsSL https://get.opentofu.org/install-opentofu.sh -o install-opentofu.sh \
    && chmod +x install-opentofu.sh \
    && ./install-opentofu.sh --install-method standalone \
    && rm -f install-opentofu.sh

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" \
    && install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl \
    && rm kubectl

# Install Flux CLI
RUN curl -s https://fluxcd.io/install.sh | bash

# Install Helm
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-4 \
    && chmod 700 get_helm.sh \
    && ./get_helm.sh \
    && rm -f get_helm.sh

# Install Kustomize
RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash \
    && mv kustomize /usr/local/bin/kustomize

# Install Rust
ENV CARGO_HOME=/usr/local/cargo
ENV RUSTUP_HOME=/usr/local/rustup
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y --no-modify-path \
    && chmod -R a+rw $CARGO_HOME $RUSTUP_HOME
ENV PATH="${CARGO_HOME}/bin:${PATH}"

# Install uv (fast Python package manager)
RUN curl -LsSf https://astral.sh/uv/install.sh | sh \
    && mv /root/.local/bin/uv /usr/local/bin/uv \
    && mv /root/.local/bin/uvx /usr/local/bin/uvx

# Install Bun globally
ENV BUN_INSTALL=/usr/local
RUN curl -fsSL https://bun.sh/install | bash

# Install Node.js
RUN curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# Install Claude Code (native installer - recommended)
RUN curl -fsSL https://claude.ai/install.sh | bash \
    && ln -sf /root/.local/bin/claude /usr/local/bin/claude

# Install OpenCode (AI coding agent)
RUN curl -fsSL https://opencode.ai/install | bash \
    && ln -sf /root/.opencode/bin/opencode /usr/local/bin/opencode

# Set PATH
ENV PATH="/usr/local/bin:/usr/bin:/bin:/root/.local/bin:${PATH}"

# Verify installations
RUN claude --version && opencode --version && node --version && bun --version && tofu --version && kubectl version --client && flux --version && helm version && kustomize version && rustc --version && cargo --version && uv --version

# Set working directory
WORKDIR /workspace

# Default command (skip permissions since container is sandboxed)
CMD ["claude", "--dangerously-skip-permissions"]
