# Provider-Native References

A reference is an opaque locator interpreted by a selected provider. A caller may
also supply an optional provider-native revision, such as a commit, branch, item
version, or immutable digest. The command passes both values to the provider instead
of parsing them into a Xonovex-specific identifier.

## Resolution

A command may infer a provider only when the subject or execution context makes one
choice unambiguous. It reports the inference and its basis. When more than one
provider could interpret the locator, the caller selects the provider explicitly; the
command stops rather than guessing.

The selected provider owns:

- locator interpretation and native revision semantics;
- authentication and authorization at the provider boundary;
- reading, writing, copying, and idempotency behavior;
- any native relationship between subjects, evidence, and results.

## Results and destinations

Operations return their result inline unless the request explicitly supplies a
destination reference. A publish or copy operation resolves that destination through
the selected provider and returns the provider-native locator plus its native revision
when supported. Inline results require no persistence provider.

Source references and result destinations remain distinct. Publishing or copying a
result does not rewrite the source reference.

## Provider-shaped examples

These examples deliberately retain each provider's shape; they are not normalized:

```text
local:   plans/example.md
git:     https://example.com/team/repository.git + commit 4f21b87
git:     https://example.com/team/repository.git + branch feature/example
github:  owner/repository pull request #42
github:  owner/repository issue #17
tracker: PROJECT-123
```

The provider decides whether a revision is required and how it is represented. For
example, Git can resolve a commit or branch, while a tracker adapter can resolve its
own item key and revision model.

There is no Xonovex ID, central resolver, central reference store, provider-neutral
reference schema, or silent local fallback. A missing or ambiguous provider remains
an explicit resolution error.

## Related guides

- [Command inventory](../README.md)
- [Role lenses](role-lenses.md)
- [Invocation and execution](invocation.md)
