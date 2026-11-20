# Typescript Plugin

Develop the TypeScript stack, its frameworks, package tooling, validation, and tests.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-typescript` in the selected harness. The bundle version is `6.0.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-typescript@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-typescript@xonovex-marketplace
```

Optional companions: `xonovex-core`.

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [publishing-npm-packages](skills/publishing-npm-packages/SKILL.md) | Publishing an npm package or checking that one is ready to publish. |
| [using-astro](skills/using-astro/SKILL.md) | Building or editing an Astro site. |
| [using-hono-opinionated](skills/using-hono-opinionated/SKILL.md) | Writing Hono APIs with inline OpenAPI route schemas, an explicit router choice, sync handlers, and request body limits. |
| [using-hono](skills/using-hono/SKILL.md) | Building or editing a Hono 4.0+ HTTP API in TypeScript. |
| [using-react](skills/using-react/SKILL.md) | Building or editing React 19+ components, hooks, or app routes. |
| [using-threejs](skills/using-threejs/SKILL.md) | Building or editing a 3D scene in Three.js for WebGL or WebGPU, or making one render faster. |
| [using-vitest](skills/using-vitest/SKILL.md) | Writing or editing Vitest 3+ tests in TypeScript. |
| [using-zod](skills/using-zod/SKILL.md) | Writing or editing Zod 4.0+ schemas that validate data at runtime in TypeScript. |
| [writing-typescript](skills/writing-typescript/SKILL.md) | Writing, editing, or reviewing TypeScript in a Node.js ESM project. |

## Validation

Run `npx moon run plugin-typescript:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
