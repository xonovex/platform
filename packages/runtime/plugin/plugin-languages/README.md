# Languages Plugin

Write Python, shell, SQL, and Lua, including TypeScript compiled to Lua.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-languages` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-languages@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-languages@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [writing-lua-opinionated](skills/writing-lua-opinionated/SKILL.md) | Tuning performance-critical Lua hot paths: the tunings especially benefit LuaJIT, and the principles apply to vanilla Lua 5.4 too. A focused overlay that covers only hot-path performance, not Lua fundamentals. |
| [writing-lua](skills/writing-lua/SKILL.md) | Editing general-purpose Lua 5.4+: modules, scripts, configuration. |
| [writing-python](skills/writing-python/SKILL.md) | Writing or editing Python 3.12+ for APIs, data processing, scripting, or tooling. |
| [writing-shell-scripts](skills/writing-shell-scripts/SKILL.md) | Writing or editing POSIX shell or Bash automation. |
| [writing-sql-postgresql](skills/writing-sql-postgresql/SKILL.md) | Editing PostgreSQL 15+ queries, schemas, or migrations. |
| [writing-typescript-to-lua](skills/writing-typescript-to-lua/SKILL.md) | Editing TypeScript that compiles to Lua via TSTL 1.24+ (game scripting, Defold, embedded engines). |

## Validation

Run `npx moon run plugin-languages:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
