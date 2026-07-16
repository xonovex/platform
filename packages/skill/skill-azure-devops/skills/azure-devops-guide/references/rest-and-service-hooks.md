# REST APIs and Service Hooks

## REST calls

Use the deployment's advertised resource areas and a supported explicit `api-version`. Keep host/collection/organization/project coordinates explicit, percent-encode identifiers, and handle continuation tokens or pagination without dropping results. A Services API view is not proof that the same endpoint or field exists on Server.

Authenticate through the narrowest supported identity. Keep tokens out of URLs, shell history, logs, previews, fixtures, and error bodies. Report accepted scopes/permissions and effective actor, not credential contents.

Writes include expected revision or ETag where supported and a stable client idempotency/correlation key. Retry only safe reads and explicitly reconcilable writes with bounded backoff, provider retry hints, cancellation, and a maximum attempt count.

## Service hooks

Discover publisher/event availability, resource version, filters, consumer/action, endpoint, credentials, TLS, payload/data classification, retry/disable behavior, and delivery history before creating a subscription. Preview exact event scope and payload exposure.

Verify a signed or otherwise authenticated receiver where supported, reject replay/duplicate delivery idempotently, tolerate out-of-order events, and recover state from the referenced provider record. A hook is a notification, not authoritative proof that the referenced change exists; re-read the native work item, pull request, build, or release.

Return subscription and delivery identifiers as evidence. On rollback, disable/delete only the owned subscription and revoke its receiver credential while preserving audit references.
