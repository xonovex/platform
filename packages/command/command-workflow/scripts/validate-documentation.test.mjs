import assert from "node:assert/strict";
import {describe, it} from "node:test";
import {
  EXPECTED_COMMANDS,
  validateCommandInventory,
  validateDisplayedInventory,
  validateManifestDependencies,
  validateSemanticResidue,
} from "./validate-documentation.mjs";

const commandPrompt = (name) => `---
description: Perform ${name} without changing unrelated state
allowed-tools:
  - Read
  - Skill
argument-hint: >-
  <subject>
---

# /xonovex-workflow:${name} — ${name}

## Goal

Do ${name}.

## Arguments

- \`subject\` is required.

## Core Workflow

1. Resolve the subject.

## Implementation

Keep the operation bounded.

## Error Handling

Stop on ambiguity.
`;

const completeInventory = () =>
  new Map(
    [...EXPECTED_COMMANDS].map((name) => [`${name}.md`, commandPrompt(name)]),
  );

describe("validateCommandInventory", () => {
  it("accepts the exact twelve-command inventory", () => {
    const result = validateCommandInventory(completeInventory());

    assert.deepEqual(result.failures, []);
  });

  it("rejects missing and extra commands", () => {
    const commands = completeInventory();
    commands.delete("create.md");
    commands.set("approve.md", commandPrompt("approve"));

    const result = validateCommandInventory(commands);

    assert.ok(result.failures.some((failure) => failure.includes("create.md")));
    assert.ok(
      result.failures.some((failure) => failure.includes("approve.md")),
    );
  });

  it("rejects duplicate command names", () => {
    const commands = completeInventory();
    commands.set("review.md", commandPrompt("create"));

    const result = validateCommandInventory(commands);

    assert.ok(
      result.failures.some((failure) => failure.includes("duplicates: create")),
    );
  });

  it("rejects malformed frontmatter and an empty prompt", () => {
    const commands = completeInventory();
    commands.set("create.md", "---\ndescription:\n---\n");

    const result = validateCommandInventory(commands);

    assert.ok(
      result.failures.some((failure) => failure.includes("description")),
    );
    assert.ok(
      result.failures.some((failure) => failure.includes("prompt body")),
    );
  });
});

describe("validateDisplayedInventory", () => {
  const readme = `Twelve independently invocable commands
The eight core operations are siblings.
The four workspace utilities manage selected workspaces.
One of the eight command verbs.
${[...EXPECTED_COMMANDS].map((name) => `[${name}](commands/${name}.md)`).join("\n")}`;

  it("rejects a stale displayed command count", () => {
    const result = validateDisplayedInventory(
      readme.replace("Twelve", "Eleven"),
      "Compose the same eight operations.",
    );

    assert.ok(result.failures.some((failure) => failure.includes("total")));
  });

  it("accepts counts and links derived from the inventory", () => {
    const result = validateDisplayedInventory(
      readme,
      "Compose the same eight operations.",
    );

    assert.deepEqual(result.failures, []);
  });
});

describe("validateManifestDependencies", () => {
  it("requires empty matching package and harness dependencies", () => {
    const result = validateManifestDependencies(
      {dependencies: {"@xonovex/skill-plan": "7.0.0"}},
      {dependencies: ["xonovex-skill-plan"]},
      {dependencies: []},
    );

    assert.ok(result.failures.some((failure) => failure.includes("match")));
    assert.ok(result.failures.some((failure) => failure.includes("empty")));
    assert.ok(
      result.failures.some((failure) => failure.includes("no universal")),
    );
  });
});

describe("validateSemanticResidue", () => {
  it("rejects former commands, command modes, and lifecycle claims", () => {
    const files = new Map([
      [
        "packages/command/command-workflow/docs/stale.md",
        [
          "Run /xonovex-workflow:qa-run --profile A2.",
          "The package enforces an approval gate.",
          "A2 requires approval before execution.",
          "The package uses a central reference resolver.",
          "The package provides a role-specific command API.",
          "Select --executor agent.",
          "Install xonovex-workflow-runtime and create an AgentSchedule.",
        ].join("\n"),
      ],
    ]);

    const result = validateSemanticResidue(files);

    assert.ok(result.failures.some((failure) => failure.includes("former")));
    assert.ok(result.failures.some((failure) => failure.includes("--profile")));
    assert.ok(result.failures.some((failure) => failure.includes("lifecycle")));
    assert.ok(result.failures.some((failure) => failure.includes("executor")));
    assert.ok(result.failures.some((failure) => failure.includes("A1/A2/A3")));
    assert.ok(result.failures.some((failure) => failure.includes("central")));
    assert.ok(
      result.failures.some((failure) => failure.includes("role-specific")),
    );
    assert.ok(result.failures.some((failure) => failure.includes("runtime")));
    assert.ok(result.failures.some((failure) => failure.includes("schedule")));
  });

  it("accepts provider-native references and execution-security policy", () => {
    const files = new Map([
      [
        "packages/command/command-workflow/docs/references.md",
        "GitHub resolves owner/repository pull request #42 without a central schema.",
      ],
      [
        "packages/agent/agent-operator-go/AGENTS.md",
        "RequireKernelIsolation and default-deny egress protect untrusted execution.",
      ],
      [
        "packages/command/command-workflow/docs/invocation.md",
        "A1, A2, and A3 are optional executor metadata with no command semantics.",
      ],
    ]);

    const result = validateSemanticResidue(files);

    assert.deepEqual(result.failures, []);
  });
});
