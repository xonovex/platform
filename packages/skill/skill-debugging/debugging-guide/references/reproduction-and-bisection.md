# reproduction-and-bisection: Minimal Deterministic Repro and Bisection

Get a reliable reproduction, aggressively shrink it to the smallest deterministic steps that still trigger the bug, then — once it is ~100% reproducible — bisect the version history to the introducing commit.

## How to Apply

1. Confirm the bug reproduces, but treat "100% reproducible" as a working assumption, not proven fact — revise it if the bug later "disappears."
2. Simplify the trigger: fewest manual steps that still crash (e.g. "start the app and pick _New Project_ twice", not a whole packaging task). Strip inputs and disable subsystems that don't change the outcome.
3. Can't simplify steps? Simplify the _data_: shrink the input file/scene until removing anything more makes the bug vanish.
4. With a deterministic repro, bisect: mark known-good and known-bad revisions; each step compiles and runs the repro to label GOOD/BAD.
5. Read the offending commit with the bug in mind — a large diff points at _which subsystem_ changed, then go back to hypothesis-driven stepping.

## Example

```bash
# A deterministic repro is the prerequisite for everything below.
git bisect start
git bisect bad                 # current revision crashes on the repro
git bisect good v2021.2        # last release that did not
# bisect checks out a midpoint; build + run the repro, then label it:
cmake --build build && ./app --new-project --new-project   # the shrunk repro
git bisect bad     # (or: git bisect good) per the result — repeat until it lands
git bisect reset
```

### Gotchas

- A flaky repro corrupts bisection — one mislabeled GOOD/BAD points it at the wrong commit; only bisect a deterministic repro.
- Incremental builds during a bisect can pick up stale objects; force a clean rebuild per step when results look inconsistent.
- Bisection is _slow_ (build + test per step) — try a direct hypothesis first, bisect when the cause is genuinely unclear.
- The offending commit names _where_ the change landed, not necessarily the buggy line — a 400-line diff still needs stepping to localize.
- For a truly random bug you can't shrink, raise the reproduction rate by stressing the suspect system — see determinism-and-replay.md.

### Related

[references/scientific-debugging.md](./scientific-debugging.md), [references/determinism-and-replay.md](./determinism-and-replay.md), **git-guide**
