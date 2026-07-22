# Bitrise API Authentication and App Discovery

Load **credential-management-guide** for provider-neutral storage, CI injection, rotation, and exposure response. This reference owns the Bitrise token's authority, unusual header, verification, and app lookup.

## Token authority

Create a personal access token under Bitrise → Account settings → Security. It is account-wide rather than repository-scoped and can access every app and organization visible to its owner. Reading logs and triggering builds use the same token, so prefer a least-privileged automation account and the shortest practical expiry.

## Exact header

Bitrise expects the raw value in `Authorization`, with no scheme prefix:

```bash
TOKEN="$(<provider-neutral-secret-store-read>)"
curl -s -H "Authorization: $TOKEN" https://api.bitrise.io/v0.1/me
```

`Authorization: Bearer ...` and `Authorization: token ...` both return `401`. An `expiring_raw_log_url` is a signed storage URL and must be fetched without the Bitrise header. Bitrise itself does not use netrc.

## Verify

```bash
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: $TOKEN" https://api.bitrise.io/v0.1/me
```

`200` confirms authentication. `401` means a bad or revoked token or an incorrect prefix.

## App slug

The app slug is the first identifier in `https://app.bitrise.io/app/<APP_SLUG>/build/<BUILD_SLUG>`. It can also be read from a git-host build-status URL or listed:

```bash
curl -s -H "Authorization: $TOKEN" "https://api.bitrise.io/v0.1/apps?limit=50" \
  | python3 -c "import sys,json;[print(a['slug'],a['title']) for a in json.load(sys.stdin)['data']]"
```

Concrete app and workflow mappings belong in project instructions. Inside Bitrise, place the token in Bitrise Secrets and restrict it to the consuming steps and trusted builds. Bitrise-to-cloud access is a separate credential and should use the provider's OIDC integration where supported.
