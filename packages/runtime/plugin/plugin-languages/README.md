# Languages Plugin

Write Python, shell, SQL, and Lua, including TypeScript compiled to Lua.

## Install

Add the [Xonovex marketplace](../../../../README.md#agent-plugins) once, then install this plugin in the selected harness.

```bash
# Claude Code
claude plugin install xonovex-languages@xonovex-marketplace

# Codex
codex plugin add xonovex-languages@xonovex-marketplace
```

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [lua-guide](skills/lua-guide/SKILL.md) | Editing general-purpose Lua 5.4+: modules, scripts, configuration. |
| [lua-opinionated-guide](skills/lua-opinionated-guide/SKILL.md) | Tuning performance-critical Lua hot paths: the tunings especially benefit LuaJIT, and the principles apply to vanilla Lua 5.4 too. A focused overlay that covers only hot-path performance, not Lua fundamentals. |
| [python-guide](skills/python-guide/SKILL.md) | Writing or editing Python 3.12+ for APIs, data processing, scripting, or tooling. |
| [shell-scripting-guide](skills/shell-scripting-guide/SKILL.md) | Writing or editing POSIX shell or Bash automation. |
| [sql-postgresql-guide](skills/sql-postgresql-guide/SKILL.md) | Editing PostgreSQL 15+ queries, schemas, or migrations. |
| [typescript-to-lua-guide](skills/typescript-to-lua-guide/SKILL.md) | Editing TypeScript that compiles to Lua via TSTL 1.24+ (game scripting, Defold, embedded engines). |

## Validation

Run `npx moon run plugin-languages:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
