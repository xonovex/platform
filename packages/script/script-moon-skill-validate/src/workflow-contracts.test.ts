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

describe("workflow composition contracts", () => {
  it("keeps core command subjects revision-addressable", () => {
    for (const command of coreWorkflowCommands) {
      const source = readRepositoryFile(
        `packages/command/command-workflow/commands/${command}.md`,
      );
      expect(source).toContain("[--subject-revision <revision>]");
      expect(source).toContain("`--subject-revision` (optional)");
    }
  });

  it("allows review to create fresh independent context", () => {
    const source = readRepositoryFile(
      "packages/command/command-workflow/commands/review.md",
    );

    expect(source).toMatch(/allowed-tools:[\s\S]*\n  - Task\n/u);
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
