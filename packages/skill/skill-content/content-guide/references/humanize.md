# humanize: Remove AI Writing Patterns

Rewrite AI-generated or stiff prose so it reads as a human wrote it. Goal: prose that fits its medium and reader, not prose that beats detectors.

## Contents

- [Workflow](#workflow)
- [Pattern Catalog](#pattern-catalog)
- [Adding Voice](#adding-voice)
- [Safety Rails](#safety-rails)
- [Required Checks Before Output](#required-checks-before-output)
- [Output Format](#output-format)
- [Examples](#examples)
- [Gotchas](#gotchas)

## Workflow

1. Read the input (file or inline). Identify medium, audience, intended tone.
2. Scan for AI patterns (see Pattern Catalog).
3. Rewrite problematic sections. Preserve meaning, match tone, add specificity.
4. Add voice where the genre allows it (opinion, rhythm variation, concrete reactions).
5. Anti-AI pass: ask "what makes this obviously AI-generated?" — answer briefly, then revise.
6. Output draft → audit notes (if audit was requested) → final rewrite.

## Pattern Catalog

### Content patterns

- **Inflated significance** — Cut `pivotal`, `testament`, `stands as`, `reflects broader`, `evolving landscape`. Replace with a concrete change or decision.
- **Credibility signaling** — Drop `featured in major publications`, `leading expert`. Name the publication, the date, or the specific work.
- **Participle fake-depth** — Avoid trailing `-ing` clauses (`creating`, `reinforcing`, `highlighting`, `ensuring`). Make it a separate sentence with a real claim.
- **Marketing tone** — Strip `vibrant`, `seamless`, `breathtaking`, `unlock potential`. Say what the thing actually does.
- **Vague attributions** — Replace `experts argue`, `observers note`, `research suggests` with a named source or cut the claim.
- **Generic filler sections** — Delete "challenges and future prospects" wrap-ups. Use a specific event or constraint instead.

### Language patterns

- **Overused vocabulary** — `crucial`, `pivotal`, `delve`, `leverage`, `foster`, `holistic`, `robust`, `seamless`, `tapestry`, `realm`, `underscore`, `showcase`, `boast`. Use only when plainly correct.
- **Copula avoidance** — Prefer `is` and `has` over `serves as`, `stands as`, `acts as`.
- **Negative parallelism** — Rewrite `not just X, but also Y` and `not X, it's Y`.
- **Rule of three** — Break reflexive triadic lists (`clearer, faster, cheaper`).
- **Elegant variation** — Repeat the ordinary word. Don't swap `app` → `application` → `platform` mid-paragraph.
- **False ranges** — Avoid `from small startups to large enterprises`. Name what you actually know.

### Style patterns

- **Em dash overuse** — Use commas, colons, or full stops by default. Reserve em dashes for genuine breaks.
- **Boldface decoration** — Bold only what a scanner needs. Don't bold every product name.
- **Inline-header lists** — Collapse `**Speed:** Faster load times` bullets into a sentence when the list isn't really list-shaped.
- **Title case headings** — Use sentence case unless the publication style requires otherwise.
- **Emojis** — Remove them unless the medium expects them.
- **Curly quotes / typographic ellipses** — Normalize to straight quotes in plain-text contexts (chat, comments, code).

### Communication patterns

- **Chatbot artifacts** — Cut `Here is a breakdown`, `Let me know if you need more details`, `I hope this helps`.
- **Knowledge-cutoff disclaimers** — Cut `details are limited`, `appears to have been introduced recently`. State the fact or leave it out.
- **Sycophancy** — Cut `Great question`, `That's a really insightful observation`.
- **Filler phrases** — `In order to` → `To`. `Has the ability to` → `can`. `It is important to note that` → cut.
- **Excessive hedging** — `might potentially` → `may`. Don't stack `maybe... sort of... I think`.
- **Generic conclusions** — Cut `Overall, the future looks promising`. End where the content ends.

## Adding Voice

Removing AI tells is half the job. Sterile, voiceless prose is just as obvious as slop.

- **Have an opinion** when the genre allows: review, comment, post, retrospective. Stay neutral when the genre demands it: docs, news summary.
- **Vary rhythm.** Mix short punchy sentences with longer ones that take their time.
- **Acknowledge complexity.** "It works, but feels like a workaround more than a real solution."
- **Use first person** when it fits the genre.
- **Be specific about feelings.** Not "this is concerning" but a concrete thing that bothers you.
- **Let some mess in.** Tangents and asides are human — when the medium allows.

## Safety Rails

- Don't invent typos, slang, or fake uncertainty to simulate humanity.
- Don't break grammar on purpose.
- Don't strip useful structure (headings, lists, citations, accessibility) for style alone.
- Don't claim detector safety. Detectors are probabilistic, not proof.
- Em dashes, semicolons, and `however` are not AI tells on their own. **Regularity is.**
- Don't invent facts to add specificity. If you can't verify a number, name, quote, or causal claim, attribute it, soften it, or cut it.

## Required Checks Before Output

1. **Register fit** — Format and tone match the medium and the user's request.
2. **Concrete anchors** — Each substantial paragraph has a name, number, quote, or observed detail.
3. **Fact discipline** — Fragile claims (dates, quotes, metrics, causal links) are verifiable or softened.
4. **Regularity tripwire** — The most repeated visible move appears fewer than 3 times.
5. **Voice calibration** — Stance present where the genre expects it, neutral where it doesn't.
6. **No over-correction** — No fake typos, forced asides, or random fragments added just to break a pattern.

## Output Format

```
DRAFT
<rewritten text>

AUDIT (if audit was requested)
<one or two lines naming the remaining AI tells>

FINAL
<revised text after the anti-AI pass>
```

If the user asked for in-place editing, write the FINAL to the source file and print only a short summary of changes.

## Examples

### Before (AI-tell heavy)

> The new feature increased user engagement by 32%. Users interacted more frequently with the dashboard, reflecting broader shifts in how teams approach analytics. Feedback has been generally positive, although some concerns remain.

### After (human voice, casual review)

> Engagement jumped 32% after the new feature shipped. Talking to a few users though, it sounds like they click more because they have to, not because they want to. The number looks great on a slide; the experience underneath is messier.

### Before (marketing puffery)

> This powerful platform offers a seamless and intuitive experience, helping teams unlock their full potential.

### After (observable consequence)

> The platform handles task tracking and reporting in one place, which cuts down on tool switching.

## Gotchas

- Removing every em dash is over-correction — regularity is the tell, not the punctuation
- Sterile, voiceless prose looks just as AI-generated as marketing puffery — add a register-appropriate stance
- Inventing facts (made-up sources, fabricated numbers) to "add specificity" is worse than the AI prose you started with — cut or soften unverifiable claims instead
