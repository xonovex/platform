import {existsSync, readFileSync} from "node:fs";
import {resolve} from "node:path";
import {describe, expect, it} from "vitest";

const repositoryRoot = resolve(import.meta.dirname, "../../../..");
const readRepositoryFile = (path: string): string =>
  readFileSync(resolve(repositoryRoot, path), "utf8");

const coreWorkflowCommands = [
  "create",
  "review",
  "revise",
  "decide",
  "execute",
  "validate",
  "publish",
  "abandon",
] as const;

const planningReferences = [
  "plan-research",
  "plan-create",
  "plan-critique",
  "plan-revise",
  "plan-subplans-create",
  "plan-continue",
  "plan-update",
  "plan-validate",
] as const;

const workspaceCommands = [
  "workspace-create",
  "workspace-merge",
  "workspace-abandon",
  "workspace-cleanup",
] as const;

const retryProtectedCommands = [
  "execute",
  "publish",
  "workspace-create",
  "workspace-merge",
  "workspace-cleanup",
] as const;

// The safety core from the architecture contract. Growing this set is a decision,
// not an accident, so the count is asserted rather than the membership alone.
const SAFETY_CORE_FLAGS = [
  "--context",
  "--destination",
  "--effect",
  "--expected-revision",
  "--idempotency-key",
  "--request",
  "--revision",
  "--source",
] as const;

const workflowCommandSources = (): readonly string[] =>
  [...coreWorkflowCommands, ...workspaceCommands].map((command) =>
    readRepositoryFile(
      `packages/command/command-workflow/commands/${command}.md`,
    ),
  );

describe("workflow composition contracts", () => {
  it("keeps the command flag surface at the safety core", () => {
    const used = new Set<string>();
    for (const source of workflowCommandSources()) {
      for (const match of source.matchAll(/(?<![`\w])--[a-z][a-z-]*/gu)) {
        used.add(match[0]);
      }
    }

    expect([...used].toSorted()).toEqual([...SAFETY_CORE_FLAGS]);
  });

  it("keeps core command subjects revision-addressable", () => {
    for (const command of coreWorkflowCommands) {
      const source = readRepositoryFile(
        `packages/command/command-workflow/commands/${command}.md`,
      );
      expect(source).toContain("`subject`");
    }
  });

  it("defines a minimal cold-boundary handoff anchored to code", () => {
    const handoffs = readRepositoryFile(
      "packages/skill/skill-workflow/workflow-guide/references/handoffs.md",
    );

    for (const heading of [
      "## Subject",
      "## What was done",
      "## Decisions",
      "## References and links",
      "## Open issues",
    ]) {
      expect(handoffs).toContain(heading);
    }
    expect(handoffs).toContain("file:line");
    expect(handoffs).toMatch(/only at a cold boundary/u);
    // The 16-field context protocol the contract replaced must not return.
    expect(handoffs).not.toMatch(
      /Context digest|Context version|Audience:|Visibility:/u,
    );
  });

  it("exposes revision and retry protection for mutating provider operations", () => {
    for (const command of retryProtectedCommands) {
      const source = readRepositoryFile(
        `packages/command/command-workflow/commands/${command}.md`,
      );
      expect(source).toContain("[--idempotency-key <key>]");
      expect(source).toContain("`--idempotency-key`");
    }

    const merge = readRepositoryFile(
      "packages/command/command-workflow/commands/workspace-merge.md",
    );
    expect(merge).toContain("[--expected-revision <revision>]");

    const effects = readRepositoryFile(
      "packages/skill/skill-workflow/workflow-guide/references/effects.md",
    );
    expect(effects).toMatch(
      /requires the exact source or\s+destination revision when that provider exposes one/u,
    );
    expect(effects).toMatch(
      /requires a stable idempotency key when the\s+provider supports one/u,
    );
  });

  it("states governance once and points every other file at it", () => {
    const governance = readRepositoryFile(
      "packages/skill/skill-workflow/workflow-guide/references/governance.md",
    );
    expect(governance).toMatch(/never instructs it/u);
    expect(governance).toMatch(/declared team convention|Declared Convention/u);

    // Only governance.md may state the untrusted-data invariant normatively.
    const restating = [
      "SKILL.md",
      "references/handoffs.md",
      "references/context-forwarding.md",
      "references/sdlc.md",
      "references/execute.md",
      "references/decide.md",
      "references/publish.md",
    ].filter((file) =>
      /untrusted data/u.test(
        readRepositoryFile(
          `packages/skill/skill-workflow/workflow-guide/${file}`,
        ),
      ),
    );

    expect(restating).toEqual([]);
  });

  it("defines Execute positively and expects an antecedent", () => {
    const execute = readRepositoryFile(
      "packages/skill/skill-workflow/workflow-guide/references/execute.md",
    );

    expect(execute).toMatch(/carries out work that was already specified/u);
    expect(execute).toMatch(/expects an antecedent/u);
    expect(execute).toMatch(/freeform session/u);
  });

  it("scopes SDLC composition to the frozen scenario families", () => {
    const sdlc = readRepositoryFile(
      "packages/skill/skill-workflow/workflow-guide/references/sdlc.md",
    );

    for (const family of [
      "Xonovex platform",
      "Drodan and CruiseReviews",
      "Native and game engine",
      "Infrastructure and operations",
    ]) {
      expect(sdlc).toContain(family);
    }
    // The 22-phase role matrix the rewrite deleted must not return.
    expect(sdlc).not.toContain("Incident commander");
    expect(sdlc).not.toContain("| Phase ");
  });

  it("keeps planning operations inline and continuation effect-aware", () => {
    for (const reference of planningReferences) {
      const source = readRepositoryFile(
        `packages/skill/skill-plan/plan-guide/references/${reference}.md`,
      );
      expect(source).not.toMatch(
        /explicitly requested provider persistence|provider destination/u,
      );
    }

    const continuation = readRepositoryFile(
      "packages/skill/skill-plan/plan-guide/references/plan-continue.md",
    );
    expect(continuation).toContain("Default to `inspect`");
    expect(continuation).toContain("Only `apply` may implement the target");
    expect(continuation).toContain(
      "Inspect and preview must not edit files, update providers, or mark plan tasks complete.",
    );
  });

  it("keeps Git merge, abandon, and cleanup as separate effects", () => {
    const merge = readRepositoryFile(
      "packages/skill/skill-git/git-guide/references/worktree-merge.md",
    );
    expect(merge).not.toContain("git worktree remove");
    expect(merge).toContain(
      "Never remove a worktree or branch during merge, whether the merge succeeds or fails",
    );

    const abandon = readRepositoryFile(
      "packages/skill/skill-git/git-guide/references/worktree-abandon.md",
    );
    expect(abandon).toContain(
      "it never commits, resets, removes, tags, pushes, or changes provider state",
    );
    expect(abandon).toContain(
      "Use separate Commit, Publish, or Cleanup operations",
    );

    const cleanup = readRepositoryFile(
      "packages/skill/skill-git/git-guide/references/worktree-cleanup.md",
    );
    expect(cleanup).toContain("Default to");
    expect(cleanup).toContain("`preview`");
    expect(cleanup).toContain("In explicit `apply`");
    expect(cleanup).toContain(
      "Require separate explicit authorization for forced removal or remote ref deletion.",
    );
  });

  it("does not restore the composition catalog layer", () => {
    expect(
      existsSync(
        resolve(repositoryRoot, "packages/skill/composition-catalog.json"),
      ),
    ).toBe(false);

    const surfaces = [
      "packages/skill/AGENTS.md",
      "packages/skill/skill-skill/skill-guide/SKILL.md",
      "packages/skill/skill-skill/skill-guide/references/composability.md",
      "packages/skill/skill-workflow/workflow-guide/references/capability-selection.md",
    ];
    for (const surface of surfaces) {
      expect(readRepositoryFile(surface)).not.toMatch(
        /composition-catalog\.json|semantic provision|preference overlay/iu,
      );
    }
  });
});
