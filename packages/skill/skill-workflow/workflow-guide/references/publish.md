# Publish

## Goal

Publish one exact subject to one explicit provider destination and return its native locator and revision without changing the source.

## Procedure

1. Pin the subject revision and explicit destination. Infer kind or provider only when unambiguous.
2. Load only explicitly selected or unambiguous publication, domain, method, and provider capabilities.
3. Let the provider own native authentication, identifiers, revisions, idempotency, and publication effects.
4. Preview the exact effect. Apply it only when the caller authorizes that subject, destination, and effect.
5. Return the provider-native locator and revision when supported.

Publication never revises the source or implies approval, acceptance, or a gate decision. Trigger, executor, identity, and maturity do not select provider behavior.

## Error handling

- Stop on a missing destination or mutable source revision.
- Report ambiguous providers or destinations without guessing.
- Stop when an explicit capability is unavailable; never fall back.
- Return a preview and request confirmation when exact authorization is missing.
