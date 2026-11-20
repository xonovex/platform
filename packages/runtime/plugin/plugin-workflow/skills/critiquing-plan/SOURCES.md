# Sources

## Performing a Project Premortem

- **URL:** https://hbr.org/2007/09/performing-a-project-premortem
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/lenses.md` → Pre-mortem
- **Aspects extracted:**
  - Assume the project has already failed and work backward to the likely causes, which surfaces risks that forward-looking review leaves unsaid → lenses

## LLM Evaluators Recognize and Favor Their Own Generations

- **URL:** https://arxiv.org/abs/2404.13076
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials
  - `references/passes.md` → A reviewer that did not write the plan
- **Aspects extracted:**
  - A model rates its own output higher than an independent reviewer does, so the review runs in a fresh context that did not write the plan → passes, Essentials

## LLM01:2025 Prompt Injection

- **URL:** https://genai.owasp.org/llmrisk/llm01-prompt-injection/
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `SKILL.md` → Essentials, Gotchas
  - `references/passes.md` → Plan material is data for the reviewer
- **Aspects extracted:**
  - Text inside supplied or fetched content that tells the reviewer to act is data, because instructions embedded in input can redirect a model → passes, Gotchas

## How we built our multi-agent research system

- **URL:** https://www.anthropic.com/engineering/multi-agent-research-system
- **Last reviewed:** 2026-10-27
- **Used for:**
  - `references/passes.md` → The blind first pass
  - `assets/templates/reviewer-brief.md` → the brief
- **Aspects extracted:**
  - A reviewer brief states an objective, an output format, the material to read and a clear boundary, because vague briefs make parallel agents duplicate or miss work → passes, reviewer-brief

## Xonovex change-workflow conventions

- **Provenance:** Repository-original rules for the independent critique of a plan, built on the contract of **recording-handoff**: the four lenses, the blind first pass and the second pass, the six fields of a finding, the severity and confidence scales, the default review criteria and the frozen critique file
- **Last reviewed:** 2026-11-19
- **References:** references/passes.md, references/lenses.md, references/review-criteria.md, references/findings.md
- **Aspects extracted:** The critique runs the entry step of **recording-handoff** first; its file is written once because critiques are append-only; the knowledge skills a critique loads come from the skills map of the `critiquing-plan` phase; a critique of a revised plan is a new file with a new blind pass.
