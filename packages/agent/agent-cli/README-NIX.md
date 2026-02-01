# Nix + Bubblewrap Sandbox for AI Agents

This sandbox method builds per-agent environments using Nix on the host and runs agents inside bubblewrap sandboxes with `/nix/store` mounted read-only.

## Gentoo Setup

### 1. Install Bubblewrap

```bash
emerge -av sys-apps/bubblewrap
```

### 2. Install Nix

**Option A: Official installer (recommended)**

```bash
# Single-user install (simpler, runs as your user)
sh <(curl -L https://nixos.org/nix/install) --no-daemon

# Or multi-user install (uses nix-daemon)
sh <(curl -L https://nixos.org/nix/install) --daemon
```

**Option B: Via Gentoo overlay** (if available)

```bash
emerge -av sys-apps/nix
```

After install, source the Nix profile:

```bash
. ~/.nix-profile/etc/profile.d/nix.sh
```

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
if [ -e ~/.nix-profile/etc/profile.d/nix.sh ]; then
  . ~/.nix-profile/etc/profile.d/nix.sh
fi
```

### 3. Enable Unprivileged User Namespaces

Check current setting:

```bash
sysctl kernel.unprivileged_userns_clone
```

If it returns `0`, enable it:

**Temporary (until reboot):**

```bash
sudo sysctl -w kernel.unprivileged_userns_clone=1
```

**Permanent:**

```bash
echo 'kernel.unprivileged_userns_clone = 1' | sudo tee /etc/sysctl.d/userns.conf
sudo sysctl --system
```

### 4. Allow Unfree Packages (for claude-code)

```bash
mkdir -p ~/.config/nixpkgs
echo '{ allowUnfree = true; }' > ~/.config/nixpkgs/config.nix
```

### 5. Verify Setup

```bash
# Check nix-build is available
which nix-build

# Check bubblewrap is available
which bwrap

# Check /nix/store exists
ls /nix/store

# Test bwrap can create namespaces
bwrap --ro-bind / / --dev /dev --proc /proc echo "bwrap works"
```

## Usage

```bash
# Build the package
npx moon run script-agent:build

# Run with nix sandbox (uses default Claude Code preset)
agent-cli run -s nix /path/to/project

# With verbose output
agent-cli run -s nix -v

# With debug output (shows bwrap command)
agent-cli run -s nix -v -d

# Dry run (show config without executing)
agent-cli run -s nix -n
```

### Nix-Specific Options

```bash
# Use the default claude preset (nodejs_24, claude-code, git, ripgrep, fd, fzf, jq, curl, coreutils, bash)
agent-cli run -s nix --nix-preset claude

# Custom package sets
agent-cli run -s nix --nix-sets nodejs,python,kubernetes

# Custom packages via image spec
agent-cli run -s nix -I 'nix:{"packages":["nodejs_24","python312","rust"]}'
```

### Other Options

```bash
# Enable/disable network
agent-cli run -s nix --network      # enabled (default)
agent-cli run -s nix --no-network   # disabled

# With a different agent
agent-cli run -s nix -a opencode

# With a provider
agent-cli run -s nix -p gemini
```

## How It Works

1. **Environment Build Phase**: Given an `EnvSpec` (list of packages), generates a pinned Nix expression and runs `nix-build` to realize dependencies into `/nix/store`.

2. **Sandbox Run Phase**: Creates per-agent directories and launches `bwrap` with:
   - `/nix/store` mounted read-only
   - Environment output mounted to `/env`
   - Writable work/tmp/home directories
   - `PATH` set to `/env/bin:/usr/bin:/bin`

## File Locations

- Nix specs: `~/.local/share/agent-nix/specs/<envId>.nix`
- Environment symlinks: `~/.local/share/agent-nix/envs/<envId>`
- Per-agent runtime: `~/.local/share/agent-nix/agents/<agentId>/{work,tmp,home}`

## Default Packages (Claude Preset)

- `nodejs_24` - Node.js runtime
- `claude-code` - Claude Code CLI
- `git` - Version control
- `ripgrep` - Fast search
- `fd` - Fast find
- `fzf` - Fuzzy finder
- `jq` - JSON processor
- `curl` - HTTP client
- `coreutils` - Basic utilities
- `bash` - Shell

## Troubleshooting

### "nix-build is not available"

Source the Nix profile or check your PATH:

```bash
. ~/.nix-profile/etc/profile.d/nix.sh
```

### "Failed to create namespace"

Enable unprivileged user namespaces:

```bash
sudo sysctl -w kernel.unprivileged_userns_clone=1
```

### Build takes too long

First build downloads and builds packages. Subsequent runs with the same packages use cached environments. Check cache status:

```bash
ls -la ~/.local/share/agent-nix/envs/
```

### "allowUnfree" errors

The `claude-code` package requires unfree packages. Ensure config exists:

```bash
cat ~/.config/nixpkgs/config.nix
# Should contain: { allowUnfree = true; }
```
