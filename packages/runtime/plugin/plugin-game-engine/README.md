# Game Engine Plugin

Build game engines, renderers, audio systems, editors, asset pipelines, and multiplayer networking.

This plugin requires [xonovex-native](../plugin-native/README.md).

<!-- xonovex:installation:start -->

## Install

Add the Xonovex marketplace once, then install `xonovex-game-engine` in the selected harness. The bundle version is `5.3.0`.

### Claude Code

Install this bundle in Claude Code. The harness discovers its bundled skills and commands.

```bash
claude plugin marketplace add xonovex/platform
claude plugin install xonovex-game-engine@xonovex-marketplace
```

### Codex

Install this bundle in Codex. Codex loads the skills declared in its plugin manifest; Claude Code slash commands are separate components.

```bash
codex plugin marketplace add xonovex/platform
codex plugin add xonovex-game-engine@xonovex-marketplace
```

Plugin dependencies: `xonovex-native`.

<!-- xonovex:installation:end -->

## Skills

Choose the skill that owns the task. The harness loads its instructions when the request matches its routing description.

| Skill | Use when |
| --- | --- |
| [asset-pipeline-guide](skills/asset-pipeline-guide/SKILL.md) | Designing the asset pipeline of a tool or game engine, turning authored sources (FBX, glTF, PNG, WAV, shaders) into runtime-ready data: importers/compilers per asset type, a deterministic compile/cook step, content-addressed caching keyed by a hash of inputs and settings, dependency tracking so an edit reimports only what changed, platform-specific output, and live hot-reloading. |
| [audio-guide](skills/audio-guide/SKILL.md) | Building a low-level, real-time audio/sound system that does its own mixing: the OS audio callback / render thread, summing voices into the output buffer, sample-rate conversion and per-voice pitch, voice pools and stealing, and handing play/stop/parameter changes from the game thread to the audio thread. |
| [c99-game-opinionated-guide](skills/c99-game-opinionated-guide/SKILL.md) | Editing C99 game-engine or runtime code in projects that follow the opinionated caller-owns-memory, SoA, builder-pattern style. A focused overlay that covers only game/engine house-style decisions, not generic C99 idioms. |
| [data-model-guide](skills/data-model-guide/SKILL.md) | Designing a central in-memory data model / object database for a tool, editor, or engine: typed objects with properties, stable cross-references and sub-object ownership, change notification, undo/redo, and serialization. |
| [ecs-guide](skills/ecs-guide/SKILL.md) | Designing or implementing a data-oriented Entity-Component-System: archetype/bitmask storage, contiguous per-type component arrays, filter-and-batch systems, change tracking, and syncing ECS state into stateful external systems (renderer, physics). |
| [editor-viewport-guide](skills/editor-viewport-guide/SKILL.md) | Building the interactive 3D viewport that bridges a real-time renderer and an editor: GPU id-buffer object picking, selection outline/highlight rendering, and move/rotate/scale gizmos that work on any object plus the linear algebra behind them. |
| [game-networking-guide](skills/game-networking-guide/SKILL.md) | Architecting real-time multiplayer networking for a game or simulation engine: client/server vs peer topology and per-object authority, replicating a typed world/component state across nodes, snapshot-vs-delta updates with baselines and acks, and reliable/unreliable/ordered channels layered over UDP. |
| [gpu-rendering-guide](skills/gpu-rendering-guide/SKILL.md) | Designing the architecture of a low-level GPU renderer on any explicit API (Vulkan/D3D12/Metal/WebGPU): render/frame graphs, shader/permutation systems, the descriptor/binding model, explicit GPU↔CPU synchronization, command recording and frames-in-flight, and GPU memory strategy. |
| [gpu-rendering-vulkan-guide](skills/gpu-rendering-vulkan-guide/SKILL.md) | Implementing a Vulkan renderer: the concrete Vulkan API for device/queues, device memory + memory types + staging, images/buffers + pipeline barriers and layout transitions, descriptor sets/layouts + bindless, pipelines + pipeline cache + dynamic rendering, timeline semaphores + fences, command pools/buffers, and the swapchain. |
| [imgui-guide](skills/imgui-guide/SKILL.md) | Designing or implementing an immediate-mode GUI (IMGUI): batching the whole UI into one draw call with compact primitive buffers, keying controls by stable IDs, resolving ordering with frame-delayed state, keyboard focus / responder chains / event trickling, drag-and-drop, per-monitor DPI scaling, string localization, and screen-reader accessibility. |
| [node-graph-guide](skills/node-graph-guide/SKILL.md) | Designing a visual node-based / data-flow graph for content authoring: typed input/output pins, wires between nodes, the graph stored as plain data, compiling/lowering to executable form vs interpreting it, topological evaluation with caching, reusable subgraphs, function graphs with declared inputs/outputs, and parametric/procedural content. |

## Validation

Run `npx moon run plugin-game-engine:ci-check --force` from the repository root to check the plugin manifests, formatting, and owned skills.
