# Typescript Plugin

Develop the TypeScript stack, its frameworks, package tooling, validation, and tests.

## Install

Add the [Xonovex marketplace](../../../../README.md#agent-plugins) once, then install this plugin in the selected harness.

```bash
# Claude Code
claude plugin install xonovex-typescript@xonovex-marketplace

# Codex
codex plugin add xonovex-typescript@xonovex-marketplace
```

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
