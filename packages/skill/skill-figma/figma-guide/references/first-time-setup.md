# first-time-setup: Get from zero to a first Figma REST call

There is **no Figma CLI to install** — the API is REST plus a token, nothing else. Every operation in this skill is an HTTPS request to `https://api.figma.com` carrying a personal access token, so a machine with `curl` and a JSON parser (`python3` or `jq`) is already provisioned. Setup is therefore one task: get a token, store it, prove it works. Credential handling proper — scopes, the storage ladder, CI/CD, rotation, hygiene — belongs to [auth.md](./auth.md); this file is only the getting-started path through it.

## Create a personal access token

Figma → Settings → Security → Personal access tokens → generate a new token. Name it after the integration so it can be revoked in isolation, set the shortest practical expiry, and grant only the scopes your endpoints need — read-only design-to-code work starts at `file_content:read`. Figma shows the value **once** at creation, so capture it straight into the keychain rather than parking it in a scratch file or the clipboard. Scope choices and the plan-gated variables scope are in [auth.md](./auth.md).

## Store it in the keychain

Store the token once, read it back at call time — never inline it. On macOS:

```bash
security add-generic-password -a "$USER" -s figma-token -U -w   # prompts on a hidden line
```

The Linux (`secret-tool`) and Windows (PowerShell SecretManagement) equivalents, shared-automation vaults, and the gitignored-`.env` fallback are the ladder in [auth.md](./auth.md) — every rung ends at the same header.

## Verify with GET /v1/me

`GET /v1/me` is the cheapest proof the credential works: it takes no file key, so it isolates auth from URL parsing, permissions, and plan gating.

```bash
TOKEN=$(security find-generic-password -a "$USER" -s figma-token -w)
curl -s -o /dev/null -w '%{http_code}\n' -H "X-Figma-Token: $TOKEN" https://api.figma.com/v1/me
```

`200` means the token is live and setup is done. Anything else is an auth problem, not a setup problem — the header goes in raw under `X-Figma-Token` and not as a `Bearer` credential, which is the usual cause of a `401`; [auth.md](./auth.md) owns that rule and reads the other status codes.

## Make the first real call

The token was the only prerequisite. A first real call is two steps from here: resolve the `figma.com` link into a file key and an API-ready node id — the id needs converting, so do not send the URL's form verbatim ([url-parsing.md](./url-parsing.md)) — then fetch that node's JSON ([read-nodes.md](./read-nodes.md)).

**Related:** [auth.md](./auth.md), [url-parsing.md](./url-parsing.md), [read-nodes.md](./read-nodes.md)
