# Variation-Point Bridges

Axes can remain independent even when one exceptional pair has a real constraint. Localize that two-axis decision in the dependent variant instead of creating a top-level combination package.

```text
template/{plain,html}
channel/{email,sms,push}

channel/email/html-support   # email depends on an HTML-template capability
combined/html-email          # bad: a new pseudo-axis for one pairing
```

- A bridge touches exactly the variants involved in one `requires` or `excludes` rule.
- Place it in the variant that needs the other axis's capability.
- Keep both axis roots unaware of concrete leaves.
- Exchange neutral data or capabilities across the bridge; do not reach into another leaf's internals.
- Keep the full cross-product in configuration. Do not hand-code a module for every pair.

If every variant needs the same behavior, it is cross-cutting rather than a bridge. If many combinations require bespoke glue, the axes may be incorrectly split.
