# scientific-debugging: Hypothesis-Driven Debugging

## Guideline

Debug like a scientist — form one falsifiable hypothesis about the cause, change exactly one thing, predict the outcome before you run, observe, and let the result narrow the search; never edit at random hoping the symptom goes away.

## Rationale

The default move on a reproducible bug is to break into the debugger _before_ the failure and step through, comparing what the code actually does against what you believe it should do — the bug lives in that gap. Random edits ("flailing") can make a symptom vanish without removing the cause, leaving a latent bug and no understanding. A single-variable experiment with a prediction is decisive: a confirmed prediction advances you, a violated prediction is information (your model was wrong, which narrows things further). Crucially, confirm the hypothesis cheaply _before_ committing to a fix direction, so you don't spend hours down a rabbit hole built on a wrong assumption.

## How to Apply

1. State the hypothesis as something that could be false: "the `tt` pointer passed in is dangling," not "something's wrong with memory."
2. Read the evidence the crash already handed you and rank causes by likelihood. An access violation means an _unmapped address_ — not a permissions or threading problem — so a garbage/dangling pointer is the leading hypothesis; a sane-but-stale address holding a freed-memory fill pattern points to use-after-free over a random overwrite.
3. Pick the cheapest experiment that could _disprove_ the leading hypothesis. Note that a crash on the very first line of a function rules out "the bug is inside this function's logic."
4. Change one variable, predict the result, run, observe. If confirmed, walk the data flow one hop further up (where did this bad pointer come from?); if refuted, take the next hypothesis.
5. Confirm before fixing: set a breakpoint where the suspect object is destroyed and check it is the same object later dereferenced, or log the pointers, so you _know_ the cause before changing code.
6. After the fix, re-run the repro to prove the symptom is gone — a fix you didn't verify is a hypothesis, not a fix.

## Example

```text
Symptom : read access violation, tt->object_types, first line of changed_objects()
Rank    : (1) tt is a dangling pointer        <- access violation = unmapped address
          (x) bug in changed_objects() logic  <- ruled out: crashes before any logic runs
          (x) "no admin access to memory"     <- access violation is not a permissions error
          (x) "need a critical section"        <- access violation is not a threading error
Look    : watch window shows tt is mapped, but *tt is all 0xdddddddd  -> use-after-free, not overwrite
Confirm : breakpoint in destroy_truth(); the destroyed Truth == the one later dereferenced. Hypothesis holds.
Narrow  : who still holds it? trace caller -> caller -> the holder; fix at the source, then re-run repro.
```

## Gotchas

- A symptom disappearing is not proof of a fix — flailing edits can perturb timing or layout enough to mask a cause that is still present.
- Trust the error's actual meaning over its scary name: "access violation" is an unmapped-address read/write, not a security or locking issue; misreading it sends the whole investigation the wrong way.
- Confirming the obvious is cheap insurance — a 30-second breakpoint that verifies "yes, this exact object was freed" prevents hours of debugging the wrong subsystem.
- Changing two things at once destroys the experiment: if the symptom changes you won't know which edit did it.
- Stepping backward through the _data flow_ (where did this value come from?) is usually more decisive than stepping forward through control flow.

## Related

[references/reproduction-and-bisection.md](./reproduction-and-bisection.md), [references/instrumentation-and-checks.md](./instrumentation-and-checks.md), **memory-management-guide**
