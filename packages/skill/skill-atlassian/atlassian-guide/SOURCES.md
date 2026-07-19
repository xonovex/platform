# Sources

## Atlassian CLI (acli) — installed binary, verified in-session

- **Provenance:** Installed `acli` binary help output and observed local command behavior
- **Version:** `acli 1.3.22-stable` (Homebrew tap `atlassian/acli`)
- **Last reviewed:** 2026-07-06
- **Used for:** the command tree and flags captured from `acli --help` and `--help` on `auth`, `jira`, `jira auth (login/status/switch/logout)`, and `jira workitem (search/view/create/transition/comment)`; the `--token` reads-from-stdin behaviour; the two auth surfaces (`acli auth` global OAuth vs `acli jira auth` per-product/API-token); and the on-disk layout — token in the macOS Keychain (service `acli`, account `jira:<cloud_id>:<account_id>`), non-secret profile in `~/.config/acli/jira_config.yaml`.
- **References:** references/auth.md, references/first-time-setup.md, references/jira.md

## Atlassian CLI documentation (Atlassian)

- **URL:** https://developer.atlassian.com/cloud/acli/
- **Last reviewed:** 2026-07-06
- **Used for:** Cloud-only scope (no Server / Data Center), the API-token vs web-OAuth login paths, non-macOS install channels (Linux deb/rpm/tarball, Windows msi/winget), and API-token creation at `id.atlassian.com/manage-profile/security/api-tokens`.
- **References:** references/auth.md, references/first-time-setup.md

## Homebrew tap-trust behaviour, verified in-session

- **Provenance:** Observed Homebrew CLI behavior in a controlled local session
- **Last reviewed:** 2026-07-06
- **Used for:** the `brew tap atlassian/acli` → `Refusing to load formula ... from untrusted tap` → `brew trust atlassian/acli` → `brew install acli` flow observed on a current Homebrew.
- **References:** references/first-time-setup.md

## Guide-level synthesis

- **Provenance:** Repository-original integration of the source blocks above; these references combine multiple inputs or maintained conventions rather than one exclusive upstream
- **References:** references/auth.md, references/first-time-setup.md, references/jira.md
- **Last reviewed:** 2026-07-06

## Refresh Workflow

1. Re-run `acli --version` and `acli jira workitem <cmd> --help` for the covered commands; diff flags against `references/jira.md` and the SKILL.md essentials.
2. Re-check the Atlassian CLI docs for new products/subcommands or changed auth paths.
3. Re-verify `acli jira auth status` still returns `✓ Authenticated` and the config/keychain layout still holds.
4. Bump the **Last reviewed** dates above.
