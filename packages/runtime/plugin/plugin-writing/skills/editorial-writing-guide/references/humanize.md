# humanize: Remove AI Writing Patterns

Rewrite AI-generated or stiff prose so it reads as a human wrote it. Goal: prose that fits its medium and reader, not prose that beats detectors.

## Checklist

- [ ] Identify the medium, audience, intended tone, and facts that must not change.
- [ ] Find repeated AI writing patterns in the text.
- [ ] Replace each pattern with a concrete statement or a simpler construction.
- [ ] Add voice only where the genre permits opinion or personal reaction.
- [ ] Ask what still makes the draft look generated, then revise once more.
- [ ] Return the final rewrite and include audit notes only when requested.

## Pattern Catalog

- **Inflated significance**: Replace `pivotal`, `testament`, `game-changer`, and claims about a broader landscape with the specific change or consequence.
- **Vague authority**: Remove `experts agree`, `studies show`, and credibility claims unless the source identifies the person, study, publication, and relevant date.
- **Empty depth**: Replace trailing `-ing` clauses, false ranges, generic challenge sections, and future-looking conclusions with a supported claim or delete them.
- **Stock vocabulary**: Use words such as `robust`, `seamless`, `holistic`, `leverage`, and `unlock` only when they state an exact meaning.
- **Indirect verbs**: Prefer `is`, `has`, and concrete actions over `serves as`, `stands as`, and `acts as`.
- **Repeated formulas**: Break automatic three-item lists, `not just X, but also Y`, repeated paragraph shapes, and identical transitions.
- **Elegant variation**: Repeat the correct ordinary term. Do not cycle through synonyms that can look like different things.
- **Decorative format**: Remove repeated bold labels, unnecessary headings, title case, emojis, and punctuation patterns when the medium does not need them.
- **Chatbot residue**: Remove praise, setup lines, offers for more help, knowledge disclaimers, excessive hedging, and repeated summaries.

## Adding Voice

- Use a clear point of view in reviews, comments, posts, and retrospectives. Stay neutral in news, documentation, and other factual genres.
- Vary sentence rhythm without adding random fragments or asides.
- Replace abstract reactions with a concrete observation or supplied experience.
- Use first person only when the medium and authorial role permit it.

## Safety Rails

- Do not invent facts, opinions, experience, typos, slang, or uncertainty.
- Do not break grammar or remove useful headings, lists, citations, or accessibility structure for style alone.
- Do not claim that the text can defeat a detector. Detection is not the goal.
- Do not ban a word or punctuation mark because it appears in generated prose. Repeated use is the signal.

## Required Checks Before Output

- **Register fit**: Format and tone match the medium and request.
- **Concrete anchors**: Substantial claims use supplied names, numbers, quotes, or observed details.
- **Fact discipline**: Dates, metrics, quotes, and causal claims remain verifiable or explicitly uncertain.
- **Voice calibration**: The text has a stance only where the genre permits it.
- **No over-correction**: The rewrite adds no fake errors, facts, asides, or fragments.

## Output Format

```text
DRAFT
<rewritten text>

AUDIT (if requested)
<one or two lines naming the remaining patterns>

FINAL
<revision after the pattern audit>
```

For in-place editing, write the final text to the source file and return a short change summary.

## Examples

### Before

> This powerful platform offers a seamless and intuitive experience, helping teams unlock their full potential.

### After

> The platform handles task tracking and reporting in one place, which cuts down on tool switching.
