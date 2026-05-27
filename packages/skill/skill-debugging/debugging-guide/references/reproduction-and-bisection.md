# reproduction-and-bisection: Minimal Deterministic Repro and Bisection

**Guideline:** Before debugging, get a reliable reproduction, then aggressively shrink it to the smallest deterministic steps that still trigger the bug; once it is ~100% reproducible, bisect the version history to the commit that introduced it.

**Rationale:** A reproducible bug is dramatically easier than a random one, because reproducibility is exactly what lets you break in _before_ the failure and step through the cause — with a random bug you can't even know in advance which run will fail. Shrinking the repro pays for itself many times over: a smaller repro iterates faster and, by stripping away everything that _doesn't_ matter, usually reveals the shape of the root cause. Once the repro is solid, bisection turns "somewhere in N commits" into a logarithmic search and hands you the exact change — but it is only as trustworthy as the repro and the GOOD/BAD labels feeding it, so it is a precision tool, not the first reflex.

## Contents

- Establish reproducibility and how confident to be in it
- Shrink the repro: fewer steps, smaller inputs, drop irrelevant systems
- Bisect history to the offending commit, and bisect's failure modes
- What to do once you have the commit

**How to Apply:**

1. Confirm the bug reproduces. Running the failing command again and hitting the same spot is good evidence — but treat "100% reproducible" as a working assumption, not a proven fact, and revise it if the bug later "disappears."
2. Simplify the trigger. Replace the slow end-to-end path with the fewest manual steps that still crash (e.g. instead of a whole packaging task, just "start the app and pick _New Project_ twice"). Strip inputs and disable subsystems that don't change the outcome.
3. If you can't simplify the steps, simplify the _data_: shrink the input file/scene until removing anything more makes the bug vanish.
4. With a deterministic repro, bisect the history. Mark a known-good and known-bad revision and follow the bisect protocol; each step compiles and runs the repro to label that revision GOOD or BAD.
5. Read the offending commit with the bug in mind. A large diff (hundreds of lines) may not show the cause on a skim, especially in unfamiliar code — use it to point at _which subsystem_ changed, then go back to hypothesis-driven stepping.
6. Decide whether to fix it yourself or hand it off: you already have the repro loaded and the context in working memory, which often beats the original author's familiarity for an "easy" bug.

**Example:**

```bash
# A deterministic repro is the prerequisite for everything below.
git bisect start
git bisect bad                 # current revision crashes on the repro
git bisect good v2021.2        # last release that did not
# bisect checks out a midpoint; build + run the repro, then label it:
cmake --build build && ./app --new-project --new-project   # the shrunk repro
git bisect bad     # (or: git bisect good) per the result — repeat until it lands
# ...
# bisect prints: <hash> "Made components provide internally stored graph instances."
git bisect reset
```

**Gotchas:**

- A flaky repro corrupts bisection — one mislabeled GOOD/BAD points it at the wrong commit; only bisect a deterministic repro.
- Incremental builds during a bisect can pick up stale objects (or hit compiler/project-setup quirks); force a clean rebuild per step when results look inconsistent.
- Bisection is _slow_ (a build + test per step) and is not the first tool to reach for — try a direct hypothesis first and bisect when the cause is genuinely unclear.
- The offending commit names _where_ the change landed, not necessarily the buggy line — a 400-line diff in an unfamiliar system still needs stepping to localize.
- For a truly random bug you can't shrink, increase the reproduction rate by stressing the suspect system (e.g. open/close 200 windows per frame) to turn a rare crash into a fast one — see determinism-and-replay.md.

**Related:** [references/scientific-debugging.md](./scientific-debugging.md), [references/determinism-and-replay.md](./determinism-and-replay.md), **git-guide**
