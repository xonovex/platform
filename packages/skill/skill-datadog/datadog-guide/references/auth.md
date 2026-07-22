# Datadog Authentication

Load **credential-management-guide** for provider-neutral keychain or secret-manager storage, CI injection, rotation, and exposure response. This reference owns Datadog key selection, scopes, transport, site, and verification.

## Choose the credential

- An API key identifies the organization and lets agents or integrations submit telemetry.
- An application key plus the organization API key authorizes programmatic API operations. Prefer a service-account-owned application key with only the required scopes; unscoped keys inherit their owner’s permissions.
- A client token is the only key type intended for browser, mobile, or other end-user clients. Never expose API or application keys there.
- Use separate named API keys per deployment method or integration so one consumer can be rotated or revoked independently.

## Transport and verification

Send API keys in `DD-API-KEY` and application keys in `DD-APPLICATION-KEY` to the API host for the selected Datadog site. Do not place secrets in URLs, command arguments, logs, previews, fixtures, or telemetry.

Validate an API key with `/api/v1/validate`; validate an API/application-key pair with `/api/v2/validate_keys`. Creation and revocation are eventually consistent, so use short bounded backoff before treating an initial `401` or `403` as definitive.

Application-key secrets may be visible only at creation time. Store the value immediately through the mechanism selected by **credential-management-guide**, verify the intended scopes and denial of an unauthorized operation, and revoke an exposed key before replacing consumers.
