# Sources

## Xonovex change-workflow conventions

- **Provenance:** Repository-original rules for delivering a work package by delegation, built on the contract of **recording-handoff** and on the steps **implementing-work-item**, **validating-work-item** and **preparing-delivery**: the supervisor role, the dependency order and the skip rules, the state of an item read from its files, the model and the effort of each step agent, the brief, the independent verification of each agent's work, the local commit bound to the verdict's tree hash, and the log entry of each item
- **Last reviewed:** 2026-11-19
- **References:** references/item-states.md, references/supervising-agents.md, references/commit-and-record.md
- **Aspects extracted:** The package run starts with the entry step of **recording-handoff**; one item and one step run at a time because each step regenerates the shared indexes; each step runs in its own agent whose model and effort are set in the call; the supervisor verifies each agent from the files and sends a finding to the same agent; a commit holds exactly the tree that a verdict with only `pass` results covers; a stopped item skips its dependents and not the package; nothing is pushed, and **validating-work-package** is offered and not started.
