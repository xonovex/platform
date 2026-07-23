# Provider-Native Resource Bindings

A binding is a typed envelope around either inline data or one opaque
provider-native locator. The envelope is provider-neutral; the locator and revision
semantics are not.

Each binding carries:

- semantic kind and optional schema;
- read or write intent;
- exactly one inline value or native locator;
- provider and opaque reference for native locators;
- exact observed revision for mutable reads, or an explicit assertion that the native
  reference is intrinsically immutable;
- create or update mode for writes; create must fail on collision and update requires
  an expected revision.

Named slots bind independently. A pull request can come from one provider, evidence
from another, criteria from a tracker, and a Publish destination from a document
store. Never apply one global provider or revision to all of them.

## Resolution

- [ ] Infer a provider only when one interpretation is unambiguous; record the basis.
- [ ] Require an exact revision when provider mutability could change the operation's
      meaning.
- [ ] Let the adapter interpret locators and native revision tokens.
- [ ] Keep source, evidence, continuation, result, and destination bindings distinct.
- [ ] Stop instead of parsing with an arbitrary provider or falling back locally.
- [ ] Recheck expected revisions immediately before a write.

The runtime coordinates authority, idempotency, and policy. The provider boundary
enforces native authentication, permissions, concurrency, and effects and returns
native receipts.

There is no universal workflow object identity, central resolver, normalized native
locator, or silent provider substitution. Provider-native references may be paths,
commits, branches, immutable digests, pull requests, issues, work items, database
records, or API resources.

The
[multi-provider request example](../assets/examples/multi-provider-review-request.json)
shows named evidence bindings with independent providers and revisions.
