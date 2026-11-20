# Languages Plugin

Write Python, shell, SQL, and Lua, including TypeScript compiled to Lua.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-languages` in the selected harness. The bundle version is `6.0.0`.

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

Optional companions: `xonovex-core`.

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [writing-lua](skills/writing-lua/SKILL.md) | Writing or editing Lua 5.4+ or LuaJIT modules, scripts, or configuration, or making a hot path faster. |
| [writing-python](skills/writing-python/SKILL.md) | Writing or editing Python 3.12+ for APIs, data processing, scripts, or tooling. |
| [writing-shell-scripts](skills/writing-shell-scripts/SKILL.md) | Writing or editing a POSIX shell or Bash script for CI tasks, build helpers, or automation. |
| [writing-sql-postgresql](skills/writing-sql-postgresql/SKILL.md) | Writing or editing PostgreSQL 15+ queries, schemas, or migrations. |
| [writing-typescript-to-lua](skills/writing-typescript-to-lua/SKILL.md) | Writing or editing TypeScript that compiles to Lua with TSTL 1.24+ for game scripting, Defold, or an embedded engine. |

## Validation

Run `npx moon run plugin-languages:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
