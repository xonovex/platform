---
name: presentation-guide
description: "Use when creating slide decks, presentations, or visual codebase walkthroughs. Triggers on prompts about scaffolding Motion-based slide packages, GraphViz diagrams, narrative flow, or codebase walkthroughs, even when the user doesn't say 'presentation'. Skip ad-hoc Markdown handouts, PowerPoint/Keynote-specific exports, and live-coded demos without slide structure."
---

# Presentation Guidelines

## Core Principles

- **Explore First** - Use Task agents to gather context from codebase
- **Structure Slides** - Organize findings into clear narrative flow
- **Visual Diagrams** - Create GraphViz diagrams for architecture and flows
- **Factory Pattern** - Use factory functions for automatic maxSteps calculation
- **Theme Consistency** - Extract and apply brand colors throughout

## Workflow

- **Create Presentation** - Explore codebase and generate markdown presentation, see [references/presentation-create.md](references/presentation-create.md)
- **Scaffold Package** - Convert markdown to Motion presentation package, see [references/presentation-motion-scaffold.md](references/presentation-motion-scaffold.md)

## Gotchas

- Slide cadence beats slide density — 3-5 bullets per slide forces the speaker to elaborate, not read
- Narrative arc (problem → tension → resolution) outperforms feature checklists — restructure if a deck reads like a manual
- GraphViz diagrams render differently across engines (`dot` vs `neato` vs `fdp`) — pick one and stick with it for visual consistency
- Motion-based slide packages need deterministic timing — avoid randomness in animations to keep slide reproducibility

## Progressive Disclosure

- Read [references/presentation-create.md](references/presentation-create.md) - Load when creating a presentation document by exploring a codebase and structuring it into slides
- Read [references/presentation-motion-scaffold.md](references/presentation-motion-scaffold.md) - Load when scaffolding a React+Motion package from a presentation markdown
