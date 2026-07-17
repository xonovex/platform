# Install and Connect the az CLI

Get from a bare machine to a working Azure DevOps Services CLI: install `az`, add the `azure-devops` extension, sign in, set org/project defaults, and verify. `az devops` is the command-line surface for Azure DevOps — there is no separate standalone CLI, and the GitHub `gh` CLI does not work against Azure DevOps. Credential storage and the `az login`-vs-PAT choice are in [auth.md](auth.md).

## Install az

Install the Azure CLI, then run `az version` to confirm (current release `2.87.0`). Update later with `az upgrade`, which also updates extensions (CLI 2.11.0+).

```bash
# macOS (Homebrew; requires macOS 13+)
brew update && brew install azure-cli

# Linux Debian/Ubuntu (one-command script)
curl -fsSL 'https://azurecliprod.blob.core.windows.net/$root/deb_install.sh' | sudo bash

# Linux RHEL 9 (dnf)
sudo rpm --import https://packages.microsoft.com/keys/microsoft.asc        # microsoft-2025.asc on RHEL/CentOS 10
sudo dnf install -y https://packages.microsoft.com/config/rhel/9.0/packages-microsoft-prod.rpm
sudo dnf install azure-cli

# Windows (winget; reopen the terminal afterwards)
winget install --exact --id Microsoft.AzureCLI

az version
```

- Debian/Ubuntu step-by-step (when the script is not allowed): `sudo apt-get install apt-transport-https ca-certificates curl gnupg lsb-release`, add the Microsoft key into `/etc/apt/keyrings/microsoft.gpg`, add a `/etc/apt/sources.list.d/azure-cli.sources` repo file (`Signed-by=/etc/apt/keyrings/microsoft.gpg`, `Suites=$(lsb_release -cs)`), then `sudo apt-get update && sudo apt-get install azure-cli`.
- Windows MSI alternative: download and run the latest MSI from `https://aka.ms/installazurecliwindows`, accept the UAC prompt, then close and reopen the terminal.

## Connect (sign in)

Sign in once; `az devops` then piggybacks on that Microsoft Entra credential — no PAT needed for interactive use. Check state first with `az account show -o table` (it errors with "Please run az login" when not authenticated).

```bash
az login                       # interactive browser sign-in (default; WAM broker on Windows 10+/Server 2019+)
az login --use-device-code     # headless/SSH/agent — relay the https://microsoft.com/devicelogin URL + code
az login --identity            # managed identity on an Azure VM/AKS/App Service
az logout                      # sign out
```

- Specific tenant (`.onmicrosoft.com` domain or tenant object id); on CLI 2.61.0+ disable the subscription selector first: `az config set core.login_experience_v2=off && az login --tenant <tenant-id-or-domain>` (re-enable with `=on`).
- User-assigned managed identity: `az login --identity --client-id <client_id>` (or `--object-id` / `--resource-id`).
- Service principal (CI/headless) and the `az login`-vs-PAT decision live in [auth.md](auth.md).
- Select the active subscription after login if needed: `az account set --subscription "<id or name>"`.

## Initialize the azure-devops extension and defaults

```bash
az extension add --name azure-devops                       # idempotent; auto-installs on first az devops command; CLI 2.30.0+
az extension show --name azure-devops                      # confirm it is present
az devops configure --defaults organization=https://dev.azure.com/<org> project=<project>
az devops configure --list                                 # view current config
```

- Set only the org default if preferred: `az devops configure -d organization=https://dev.azure.com/<org>`.
- The org default must be the full URL form `https://dev.azure.com/<org>` — a bare org name will not resolve.

## Verify

```bash
az devops project list -o table     # add --org https://dev.azure.com/<org> if no default is set
az repos list -o table              # optional: confirm repo access
```

## Gotchas

- The `azure-devops` extension works only with Azure DevOps Services (cloud, `dev.azure.com`); it does not drive Azure DevOps Server (on-premises). That is a limitation of this CLI tool, not of Server support overall — reach a Server deployment through the REST / adapter path in [editions-and-capabilities.md](editions-and-capabilities.md). The old `vsts-cli` is deprecated and folded into this extension.
- Never run plain `az login` in an unattended/agent context — it blocks waiting for a browser. Use `az login --use-device-code`; the device code expires in ~15 minutes, so restart the login if the user is slow.
- MFA is being enforced (rolling out from September 2025) for user identities. Username/password login (`az login -u/-p`) fails with "interactive authentication is needed" for MFA/Microsoft accounts — migrate automation to a service principal or managed identity.
- Wrong default tenant is a common failure: "Authentication failed against tenant" means restart with `az login --tenant <tenant-id-or-domain>`. On CLI 2.61.0+ the subscription selector can interfere with `--tenant`; disable it with `az config set core.login_experience_v2=off` first.
- Derive org/project from the git remote, not a guess: remotes embed them as `.../<org>/<project>/<repo>` (SSH v3 form `:v3/<org>/<project>/<repo>`); `<org>` maps to `https://dev.azure.com/<org>`.
- If `az devops project list` returns nothing or errors after login, you likely have no org default set and did not pass `--org`, or your account lacks access to that org.
- macOS Homebrew requires macOS 13+. On Windows, close and reopen the terminal after the MSI install before `az` is on PATH.
