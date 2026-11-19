# Native Plugin

Write, build, and debug portable native C99 systems with explicit memory, concurrency, and data layouts.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-native` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-native@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-native@xonovex-marketplace
```

Optional companions: `xonovex-core`, `xonovex-game-engine`.

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [applying-data-oriented-design](skills/applying-data-oriented-design/SKILL.md) | Making data layouts cache-friendly for performance-critical code, in any language. |
| [debugging-native-code](skills/debugging-native-code/SKILL.md) | Finding the cause of a crash, memory corruption, leak, or intermittent bug in native C or C++ code, or making a class of bug impossible by design. |
| [managing-memory](skills/managing-memory/SKILL.md) | Deciding how memory is allocated, owned, and freed in manual-memory code, in any language. |
| [using-cmake](skills/using-cmake/SKILL.md) | Writing or editing CMake builds for C or C++ projects on CMake 3.20+. |
| [writing-c99-opinionated](skills/writing-c99-opinionated/SKILL.md) | Writing or editing systems or embedded C99 libraries with no hidden allocation, caller-provided buffers, and handle indices. |
| [writing-c99](skills/writing-c99/SKILL.md) | Writing, editing, or reviewing plain C99 code such as libraries, CLI tools, and system utilities. |
| [writing-cross-platform-code](skills/writing-cross-platform-code/SKILL.md) | Porting native C or C++ software to another OS or to the web, or putting OS calls behind one platform layer. |
| [writing-lock-free-code](skills/writing-lock-free-code/SKILL.md) | Writing or reviewing shared-memory concurrent code with atomics, lock-free data structures, or custom synchronization. |

## Validation

Run `npx moon run plugin-native:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
