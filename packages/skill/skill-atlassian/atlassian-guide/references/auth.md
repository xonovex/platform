# Atlassian Cloud CLI Authentication

Load **credential-management-guide** for provider-neutral storage, CI injection, rotation, and exposure response. This reference owns Atlassian credential choices, `acli` transport, profiles, and verification.

## Two auth surfaces

- `acli auth login` uses a global browser OAuth flow and authenticates Jira and Confluence sites. Tokens refresh natively; this path requires a browser and permissive organization policy.
- `acli jira auth login` is per-product and is the only path that accepts an API token. Use it for headless or OAuth-restricted environments.

`acli auth status` reports global OAuth accounts. `acli jira auth status` reports the active Jira profile and whether it uses `oauth` or `api_token`.

## API token

Create, expire, rotate, and revoke a token at `https://id.atlassian.com/manage-profile/security/api-tokens`. A basic token inherits its owner's Jira permissions; prefer the narrowest supported scoped token. If organization policy blocks both OAuth and token creation, escalate to an administrator rather than inventing another login path.

The `--token` flag reads stdin:

```bash
IFS= read -r -s -p "Atlassian API token: " ATLASSIAN_API_TOKEN
printf '\n'
printf '%s' "$ATLASSIAN_API_TOKEN" | acli jira auth login \
  --site <site>.atlassian.net --email you@example.com --token
unset ATLASSIAN_API_TOKEN
```

In CI, inject `ATLASSIAN_API_TOKEN` from the runner's secret store immediately before login. Keep shell tracing disabled. After login, `acli` stores the credential in the operating-system secret store. Its `~/.config/acli/*.yaml` files contain profile metadata, not the token. Do not create a second plaintext copy.

## Verify and switch

```bash
acli jira auth status
acli jira project list >/dev/null
acli jira auth switch --site <site>.atlassian.net --email you@example.com
```

The status command confirms a stored profile; the list call verifies it against Jira. A `401` may mean an expired or revoked token, wrong site, or email that does not own the token.

## Automation

Browser OAuth is not a CI login. Use a dedicated, scoped API token when `acli` automation has no federated alternative, inject it through the CI secret store, and pipe it to login at job start. **credential-management-guide** owns the cross-platform store, protected-job checks, rotation, and revoke-first response.
