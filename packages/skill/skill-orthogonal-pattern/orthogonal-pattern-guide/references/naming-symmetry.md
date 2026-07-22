# Naming and Symmetry

Name an axis for the question and each leaf for one answer. Siblings use the same grammatical form and level of abstraction.

```text
BAD  delivery/{send-email,sms-provider,push}/
GOOD channel/{email,sms,push}/
```

- Axis names are stable domain nouns such as `channel`, `format`, or `destination`.
- Variant names are bare values such as `email`, `json`, or `object-store`; do not repeat the axis.
- Avoid mechanism or catch-all siblings such as `helpers`, `services`, `manager`, or `util` beside domain axes.
- Put variant-specific options under the axis and variant namespace, such as `channel.email.sender` and `channel.sms.number`.

Parallel consumers should reuse the same axis vocabulary even when their implementations differ. A web request path and a background worker may realize `channel` differently, but silent renaming to `sender` in one and `transport` in the other hides that they implement the same decision.
