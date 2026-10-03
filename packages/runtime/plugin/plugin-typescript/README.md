# Typescript Plugin

Develop the TypeScript stack, its frameworks, package tooling, validation, and tests.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-typescript` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills and commands.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-typescript@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest; Claude Code slash commands are separate components.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-typescript@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [astro-guide](skills/astro-guide/SKILL.md) | Editing or scaffolding Astro sites with islands architecture. |
| [hono-guide](skills/hono-guide/SKILL.md) | Editing or scaffolding Hono 4.0+ API servers in TypeScript. |
| [hono-opinionated-guide](skills/hono-opinionated-guide/SKILL.md) | Editing Hono APIs that follow the opinionated style: inline OpenAPI handlers, explicit router selection (LinearRouter / RegExpRouter), sync handlers where possible, bodyLimit middleware. A focused overlay that covers only house-style decisions, not generic Hono usage. |
| [npm-guide](skills/npm-guide/SKILL.md) | Publishing an npm package or checking that one is ready to publish. |
| [react-guide](skills/react-guide/SKILL.md) | Building or editing React 19+ components, hooks, or app routing. |
| [threejs-guide](skills/threejs-guide/SKILL.md) | Building or editing 3D scenes in vanilla Three.js for WebGL/WebGPU. |
| [typescript-guide](skills/typescript-guide/SKILL.md) | Editing or reviewing TypeScript in Node.js ESM projects. |
| [vitest-guide](skills/vitest-guide/SKILL.md) | Writing or editing Vitest 3+ tests in TypeScript. |
| [zod-guide](skills/zod-guide/SKILL.md) | Defining or editing Zod 4.0+ schemas for runtime validation in TypeScript. |

## Validation

Run `npx moon run plugin-typescript:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
