# Game Engine Plugin

Build game engines, renderers, audio systems, editors, asset pipelines, and multiplayer networking.

This plugin requires [xonovex-native](../plugin-native/README.md).

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-game-engine` in the selected harness. The bundle version is `7.0.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-game-engine@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-game-engine@xonovex-marketplace
```

Plugin dependencies: `xonovex-native`.

Optional companions: `xonovex-core`.

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [building-editor-viewports](skills/building-editor-viewports/SKILL.md) | Building click-to-select, selection outlines, or move/rotate/scale gizmos in a 3D editor viewport. |
| [designing-asset-pipelines](skills/designing-asset-pipelines/SKILL.md) | Building the asset pipeline of a game engine or tool: importing FBX, glTF, PNG, WAV or shader sources and cooking them into runtime data. |
| [designing-data-models](skills/designing-data-models/SKILL.md) | Designing the in-memory object model of an editor, tool, or engine: typed objects with properties, references between them, undo/redo, and save/load. |
| [designing-ecs](skills/designing-ecs/SKILL.md) | Designing or implementing an Entity-Component-System for a game or simulation. |
| [designing-game-networking](skills/designing-game-networking/SKILL.md) | Adding real-time multiplayer networking to a game or simulation engine. |
| [designing-gpu-renderers](skills/designing-gpu-renderers/SKILL.md) | Designing the architecture of a GPU renderer on an explicit API such as Vulkan, D3D12, Metal or WebGPU. |
| [designing-imgui](skills/designing-imgui/SKILL.md) | Building an immediate-mode GUI or ImGui-style widgets. |
| [designing-node-graphs](skills/designing-node-graphs/SKILL.md) | Building a visual node editor or data-flow graph for content authoring or procedural content. |
| [implementing-vulkan-renderers](skills/implementing-vulkan-renderers/SKILL.md) | Writing Vulkan code for a renderer: devices and queues, memory and staging, barriers and layout transitions, descriptors, pipelines, synchronization, and the swapchain. |
| [mixing-audio](skills/mixing-audio/SKILL.md) | Building a low-level real-time audio engine that mixes its own voices. |
| [writing-c99-game-opinionated](skills/writing-c99-game-opinionated/SKILL.md) | Writing or editing C99 game-engine math, physics, mesh, or geometry code that writes into caller-owned buffers. |

## Validation

Run `npx moon run plugin-game-engine:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
