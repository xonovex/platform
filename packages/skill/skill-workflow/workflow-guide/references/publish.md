# Publish

1. Resolve one exact subject, optional subject revision, one explicit destination,
   binding preconditions, expected destination revision, and applicable capabilities.
2. When the subject is forwarded context, validate its ID, positive version, digest,
   supersedes relationship, provenance, applicability, audience, destination
   visibility, and absence of secrets or destination-inappropriate internal detail.
3. Default to `preview` and report the exact external write.
4. Before `apply`, require explicit authority, the expected destination revision when
   an existing destination exposes one, and a stable idempotency key when the provider
   supports one.
5. Reconcile the exact destination and context marker immediately before `apply`.
   Return one exact existing ID/version/digest as a no-op, block duplicates or
   same-version divergence, and create an incremented successor for changed semantics.
6. Apply only the previewed write, then verify the observed destination reference and
   revision.
7. Report failed or unknown effects without claiming total success.

A provider-native note or pull-request comment is a valid destination for selected
context. Treat the comment as a projection unless it is the declared authoritative
record, and return its native reference so later operations can supply it through
`--context` or a Markdown handoff. Never publish all session context implicitly.

Append-only context publication supplies application-level retry safety even when the
provider has no idempotency key or conditional note update. For other subjects, name
the non-idempotent retry boundary and reconcile the exact destination before retrying
an unknown write.

Publication persists a result; it does not imply approval.
