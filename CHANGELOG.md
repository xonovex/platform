# Changelog

Release notes for the plugin catalog line: the `packages/runtime/plugin/*` plugins versioned in lockstep with `.claude-plugin/marketplace.json` and `.agents/plugins/marketplace.json`. The agent CLI keeps its own notes in `packages/tooling/cli/*/CHANGELOG.md`.

## 6.0.0 - 2026-11-20

10 grouped plugins with 86 skills. Previous catalog version: 5.3.0.

### Breaking changes

- The 77 individual `xonovex-skill-*` and `xonovex-command-*` plugins are removed. Ten grouped plugins hold the skills: `xonovex-agentic`, `xonovex-core`, `xonovex-game-engine`, `xonovex-integrations`, `xonovex-languages`, `xonovex-native`, `xonovex-platform`, `xonovex-typescript`, `xonovex-workflow` and `xonovex-writing`.
- The slash commands are removed. Each skill takes the request directly.
- The skills have gerund names, for example `using-git`, `reviewing-code` and `writing-typescript`; the `-guide` names are removed.
- The workflow plugin is replaced by a port of the ns workflow. It keeps each change in a directory of files that each of its twelve steps reads and writes, from `researching-change` to `validating-work-package`.
- `delegating-plan` and `distilling-plan` are removed.
- `reflecting-on-sessions` is removed. `authoring-instructions` and `authoring-skills` turn the lessons of a session into rules and skills.
- `debugging-native-code` is in `xonovex-native`.
- `writing-lua-opinionated` is part of `writing-lua`.

### Added

- `evaluating-skills` runs the trigger and output evaluations of a skill, separate from `authoring-skills`.
- The skill evaluations use Claude Haiku 5.5 to generate and Claude Sonnet 5.5 to judge.

### Changed

- Each description leads with the task in the words of a user. The catalog descriptions total 34,593 characters, down from 41,044.
- The skill validator fails a description above 820 characters.

## 5.3.0 - 2026-08-29

77 skill and command plugins. Previous catalog version: 5.2.0.

### Added

#### Skills

- `xonovex-skill-writing` owns shared clarity, terminology, evidence, and concision rules.
- `xonovex-skill-editorial-writing` owns editorial voice, register, readability, multilingual consistency, and humanization.
- `xonovex-skill-news-writing` owns current-news research, verification, multilingual files, and publication metadata.
- `xonovex-skill-travel-writing` owns destination and venue research, practical facts, itineraries, and multilingual guide files.
- `xonovex-skill-technical-writing` owns outcome-first technical prose, decisive caveats, and exact technical meaning.

### Changed

- The content utility commands keep their public names and now delegate to the matching editorial, news, or travel writing skill.
- `xonovex-skill-code-quality` adds a comment inventory that distinguishes removable narration from licenses, shebangs, formatter directives, lint suppressions, coverage markers, and tool annotations.

### Removed

- `xonovex-skill-content` is replaced by the focused editorial, news, and travel writing plugins. Installed copies keep working, but the plugin no longer resolves from either marketplace.

## 5.2.0 - 2026-08-23

73 skill and command plugins. Previous catalog version: 5.1.0.

### Added

#### Skills

- skill-plan gains the followup operation, which closes out a completed, paused,
  or handed-over plan with an inline record of evidenced status, gates and
  blockers, an open-decision register, residue, follow-up plan seeds, a review
  brief, and a retrospective.
- skill-plan gains the distill operation, which turns a completed plan's branch,
  record, and execution logs into a self-contained replayable skill suite, with
  provenance quarantine and mechanical suite verification.

#### Commands

- /xonovex-workflow:plan-followup and /xonovex-workflow:plan-distill, the entry
  points for the two new plan-guide operations.

## 5.1.0 - 2026-08-06

73 skill and command plugins. Previous catalog version: 5.0.0.

### Removed

These are breaking changes, released under a minor by decision. Installed
copies keep working; none of these resolve from either marketplace any more.

#### Commands

- /xonovex-workflow:acceptance-formalize, acceptance-validate, story-refine,
  pr-create, pr-review-analyze, pr-review-refine, pr-review-post and
  pr-review-resolve. Entry triage now goes straight to plan-research, and
  merge and delivery run into the Definition-of-Done gate. The procedures
  are unaffected and still live in the pull-request-guide, code-review-guide,
  bdd-guide, user-stories-guide, testing-guide, github-guide and
  gitlab-guide skills.
- /xonovex-utility:version-bump. versioning-guide still carries the procedure.
- /xonovex-utility:skill-ablate, folded into /xonovex-utility:skill-optimize
  as its verify phase. skill-optimize gains --all for catalog scale, --model
  and --report-only.

#### Skill plugins

- xonovex-skill-caveman, a terse writing-style overlay, and
  xonovex-skill-fable, a voice-imitation prompt overlay. No package depended
  on either.
- xonovex-skill-adr, -android-analytics, -android-wcag, -expressjs, -fdd,
  -motion-react, -presentation, -remotion and -strudel, retired when the
  catalog closed on its consolidated state.

### Added

- /xonovex-workflow:plan-delegate, supervising a roadmap by delegation.
- Harness skills xonovex-skill-claude-code, -codex, -copilot, -kiro,
  -opencode and -pi.
- xonovex-skill-accessibility and xonovex-skill-credential-management.
