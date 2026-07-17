# First-time Bitrise API setup

Zero to a first successful Bitrise API call: mint a personal access token, put it in the OS keychain, and prove it works with `GET /v0.1/me`. There is nothing to install and no login command to run. This page is only the path — [auth.md](auth.md) owns credential handling and [builds.md](builds.md) owns the operations that follow.

## There is no CLI to install

The Bitrise API has no command-line client: no `auth login`, no credential helper, no config file a tool populates for you. Every call is `curl` plus a token you attach yourself, so the whole setup is mint a token -> store it -> send it as a header.

The `bitrise` CLI that does exist is a **local `bitrise.yml` workflow runner** (`bitrise run <workflow>`, `bitrise validate`) for executing Steps on your own machine. It holds no API token and cannot list apps, read builds, or fetch logs — installing it gets you no closer to an API call.

That leaves two prerequisites, both usually already present:

```bash
curl --version      # any recent curl
python3 --version   # used to pull fields out of JSON responses
```

## Mint a personal access token

Bitrise -> **Account settings** -> **Security** -> **Personal access tokens** -> generate a token, with the shortest practical expiry rather than "never".

The value is shown **once**, at creation. Know what you are holding before you generate it: the token is **account-wide** — not repo-scoped, and the same token that reads a log can start a build — so prefer one belonging to a least-privileged automation account over a personal admin login. [auth.md](auth.md) has the scope, rotation, and revocation detail.

## Store it in the OS keychain

Capture the value straight from the creation screen into the OS keychain under the name `bitrise-token`. Do not park it in a file, a hand-typed shell variable, or the clipboard on the way.

[auth.md](auth.md) owns the storage ladder — the per-OS keychain commands (macOS `security`, Linux `secret-tool`, Windows SecretManagement), the team secret-manager reads, and the gitignored-file fallback of last resort. Take the highest rung the machine supports, then come back here.

## Verify with `GET /v0.1/me`

`/v0.1/me` is the cheapest proof the token works. Run it before any real call:

```bash
TOKEN=$(security find-generic-password -s bitrise-token -w)   # macOS; auth.md has the Linux / Windows / team reads
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: $TOKEN" https://api.bitrise.io/v0.1/me
```

`200` and setup is done. Note the header shape: Bitrise takes the **raw** token with no `Bearer` / `token ` prefix, and that prefix is the usual cause of a `401` on a token that is otherwise fine. [auth.md](auth.md) has the rule and the rest of the 401 triage.

## Find the app slug

Everything past `/me` is app-scoped (`/v0.1/apps/{app}/...`), so an app slug is the next thing to resolve. [auth.md](auth.md) covers the three ways to get one — read it off a build URL, take it from the git host's build-status entry, or list your apps over the API — and [builds.md](builds.md) shows that same slug split out of an `app.bitrise.io` URL in context.

## Next

- Token reach, rotation, a stubborn `401`, or CI and cloud storage: [auth.md](auth.md).
- First real work — map a commit to its build, read a log, triage flakiness, re-trigger: [builds.md](builds.md).
