# comment-inventory: Comment Inventory

Produce an evidence-linked, read-only inventory. Do not edit, delete, rewrite, or reformat source during the audit.

## Scope

- Read repository instructions and formatter, linter, type-checker, coverage, and documentation configuration before applying generic rules.
- State included roots, file types, and excluded path classes.
- Exclude generated, vendored, minified, lock, snapshot, golden, and third-party content unless the request or policy makes it owned source.
- Prefer language-aware parsing. Text search discovers candidates but cannot distinguish comments and docstrings from strings, runtime messages, examples, queries, or data.

## Preserve Functional Text and Rationale

- Treat shebangs, licenses, formatter controls, suppressions, coverage markers, generation markers, build annotations, and other tool-consumed text as functional. A shebang is executable metadata.
- Preserve functional text. Prefer a tool's unused-suppression diagnostic over inference.
- Preserve an unknown annotation and report it as ambiguous. The current environment can lack the tool that consumes it.
- Preserve concise, current explanations of a non-obvious reason, invariant, constraint, workaround, caveat, protocol, compatibility boundary, or measured tradeoff.
- Preserve public API documentation when it states behavior, errors, ownership, units, ordering, side effects, or constraints that the signature cannot express.

## Flag Prose Debt

- Flag comments that contradict current behavior or state a false invariant.
- Flag comments that restate a well-named declaration or narrate the next statement.
- Flag plan, task, agent, document, phase, porting, migration, and future-work narration that does not describe current behavior.
- Flag commented-out code, decorative section banners, and walkthroughs whose structure and names already provide the same information.
- Flag docstrings that only repeat declaration names, parameter names, types, or return types.
- Route `TODO`, `FIXME`, and project-defined work markers to the TODO and FIXME inventory. Do not count a marker twice.

## Establish Evidence

- Read surrounding code to decide whether a comment explains why or repeats what. Check declarations and callers for behavioral claims.
- Use history only when current source cannot establish whether provenance remains necessary.
- Keep references to external contracts on which the implementation depends.
- Group repeated instances with one cause into one finding and cite every location. Finding count does not determine severity.

## Grade and Recommend

- **High**: The comment contradicts behavior or states a false requirement that can misdirect a change.
- **Medium**: The comment narrates a plan, task, agent, document, phase, migration, or obsolete implementation; contains dead code; or mixes useful rationale with stale narration.
- **Low**: The comment only restates clear code, repeats a signature, adds a decorative heading, or uses more words than its rationale needs.
- **Ambiguous**: The text may be functional, generated, policy-controlled, or contract-bearing. Preserve it and name missing evidence.

Recommend **preserve**, **delete**, **rewrite**, or **investigate**. A rewrite keeps only current rationale in present tense. Recommendations remain read-only unless the user separately authorizes implementation.

## Report and Check Coverage

- Lead with scope and counts. List findings from high to low with location, classification, evidence, and action.
- Then list preserved and ambiguous classes, exclusions, coverage gaps, tool limits, and missing evidence.
- Rescan for forbidden markers and finding patterns. Confirm that functional or ambiguous text is not actionable without contrary evidence.

## Examples

Flag restatement and project narration, but preserve the invariant and unknown annotation:

```ts
// Load the current configuration.
const config = loadCurrentConfiguration();

// Phase 2 follows roadmap item R3.
initializeCache(config);

// Hash bytes because the remote signature does not cover Unicode code points.
const digest = hash(encoder.encode(value));

// tool-that-may-not-be-installed: preserve-next
runCompatibilityCase();
```
