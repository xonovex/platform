# Local Credential Storage

## Preferred stores

- **macOS** - Use the login Keychain through a provider CLI or `security`. Store with a hidden prompt; retrieve by stable service and account attributes.
- **Linux desktop** - Use a Secret Service implementation such as GNOME Keyring or KWallet through the provider CLI or `secret-tool`. Confirm a usable session collection exists in headless environments.
- **Windows** - Prefer the provider's Credential Manager integration or a reviewed PowerShell SecretManagement vault that supports retrieval. `cmdkey` writes credentials but is not a general secret-read interface.
- **WSL and headless machines** - Use a supported Linux secret service, bridge deliberately to a Windows vault, or fetch from a remote secret manager. Do not silently fall back to plaintext.

Let a provider CLI own its native keyring entry when it already does so. Avoid copying the same credential into a second dotfile or vault.

## Retrieval

Retrieve into the narrowest process scope immediately before the command. Prefer stdin, a documented environment variable, or an authorization header assembled in memory according to the provider guide. Avoid literal command arguments, URLs, shell history, clipboard workflows, tracing, and debug dumps.

## Plaintext fallback

A gitignored `.env` or `0600` file is a local-development fallback, not a secret store. Add the ignore rule before creating the file, commit only placeholders, restrict permissions, and document replacement with a keychain or manager. `.gitignore` does not protect an already committed value or other processes running as the same user.
