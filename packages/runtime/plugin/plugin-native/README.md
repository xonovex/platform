# Native Plugin

Write C99 and build portable native systems with explicit memory, concurrency, and data layouts.

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-native` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills and commands.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-native@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest; Claude Code slash commands are separate components.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-native@xonovex-marketplace
```

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [c99-guide](skills/c99-guide/SKILL.md) | Editing or reviewing general-purpose C99: libraries, CLI tools, system code without strong opinions on memory ownership or layout. |
| [c99-opinionated-guide](skills/c99-opinionated-guide/SKILL.md) | Editing systems or embedded C99 code in projects that follow the opinionated caller-owns-memory, data-oriented style. A focused overlay that covers only house-style decisions, not generic C99 idioms. |
| [cmake-guide](skills/cmake-guide/SKILL.md) | Editing CMake build files for C/C++ projects on CMake 3.20+. |
| [cross-platform-guide](skills/cross-platform-guide/SKILL.md) | Making native C/C++ software portable across operating systems and targets: isolating all OS/windowing/input/audio calls behind one platform-abstraction interface, ordering a port to a new OS, building to the web via Emscripten/WebAssembly (cooperative main loop, async file/network, GL ES/WebGPU mapping, 32-bit pointers, memory growth), and reading input devices like gamepads. |
| [data-oriented-design-guide](skills/data-oriented-design-guide/SKILL.md) | Designing or refactoring performance-critical data layouts for cache efficiency, in any language. |
| [lock-free-guide](skills/lock-free-guide/SKILL.md) | Writing or reviewing shared-memory concurrent code: atomics, lock-free/wait-free data structures, or scalable synchronization. |
| [memory-management-guide](skills/memory-management-guide/SKILL.md) | Deciding how memory is allocated, owned, and freed in manual-memory or buffer-passing code, in any language. |

## Validation

Run `npx moon run plugin-native:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
