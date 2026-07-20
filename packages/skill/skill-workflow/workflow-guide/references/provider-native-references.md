# Provider-Native References

Treat a reference as an opaque locator interpreted by the selected provider. Pass its optional native revision separately; never translate it into a universal workflow identifier.

## Resolution

- Infer a provider only when the subject or execution context makes one choice unambiguous, and report the inference and basis.
- Stop on ambiguity instead of parsing a locator with an arbitrary provider.
- Let the provider own authentication, authorization, locator interpretation, revision semantics, reading, writing, idempotency, relationships, and side effects.
- Preserve source references and result destinations as distinct values.
- Never silently replace an unavailable selected provider with a local file or another provider.

## Results

Return results inline unless the caller explicitly supplies a destination. When persisting or publishing, return the provider-native locator and native revision when supported. A successful provider operation does not imply approval, acceptance, or authorization beyond the exact selected effect.

Provider-shaped locators may be paths, commits, branches, immutable digests, pull requests, issues, work items, database records, or API resources. Their representation and guarantees remain provider-owned.
