import {existsSync, readdirSync, readFileSync} from "node:fs";
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
  "research",
  "create",
  "critique",
  "revise",
  "expand",
  "continue",
  "update",
  "validate",
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

// The safety core from the architecture contract. This literal stays a literal on
// purpose: the commands are the only place the flags exist, so deriving the expected
// set from them would compare the commands to themselves and assert nothing. Like a
// recorded budget, it is an external anchor that makes growing the surface a
// decision rather than an accident.
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

// Every guide name the catalog ships, so the inventory ban derives from what
// exists rather than from a list that ages the moment a skill is added.
const installedGuideNames = (): readonly string[] =>
  readdirSync(resolve(repositoryRoot, "packages/skill"), {
    withFileTypes: true,
  })
    .filter((entry) => entry.isDirectory() && entry.name.startsWith("skill-"))
    .flatMap((entry) =>
      readdirSync(resolve(repositoryRoot, "packages/skill", entry.name), {
        withFileTypes: true,
      })
        .filter((guide) => guide.isDirectory() && guide.name.endsWith("-guide"))
        .map((guide) => guide.name),
    );

const commandFileName = (operation: string): string =>
  operation.trim().toLowerCase().replaceAll(" ", "-");

// The 'Modes Per Operation' table in effects.md, expanded to one entry per
// command. An operation that takes no --effect maps to undefined.
const effectDefaultsFromSkill = (
  effects: string,
): ReadonlyMap<string, string | undefined> => {
  const section = effects.split("## Modes Per Operation", 2).at(1) ?? "";
  const defaults = new Map<string, string | undefined>();
  for (const line of section.split("\n")) {
    const cells = line.split("|").map((cell) => cell.trim());
    if (cells.length < 5 || cells[1] === undefined) continue;
    const accepts = cells[2] ?? "";
    const fallback = /`([a-z]+)`/u.exec(cells[3] ?? "")?.at(1);
    if (fallback === undefined) continue;
    const takesArgument = !accepts.includes("no `--effect`");
    for (const operation of cells[1].split(",")) {
      defaults.set(
        commandFileName(operation),
        takesArgument ? fallback : undefined,
      );
    }
  }
  return defaults;
};

// The default one command documents for its own --effect argument, or
// undefined when it exposes no such argument.
const effectDefaultFromCommand = (command: string): string | undefined => {
  const source = readRepositoryFile(
    `packages/command/command-workflow/commands/${command}.md`,
  );
  if (!source.includes("`--effect`")) return undefined;
  return /`--effect`[^\n]*Defaults to `([a-z]+)`/u.exec(source)?.at(1);
};

describe("workflow composition contracts", () => {
  // Reads the workflow command files, which live in the command packages.
  it.skip("keeps the command flag surface at the safety core", () => {
    // Every occurrence, not just the argument-hint ones: a backtick guard here
    // would leave the Arguments section and the delegation prose unchecked, so a
    // flag documented in only one of them would pass.
    const used = new Set<string>();
    for (const source of workflowCommandSources()) {
      for (const match of source.matchAll(/--[a-z][a-z-]*/gu)) {
        used.add(match[0]);
      }
    }

    expect([...used].toSorted()).toEqual([...SAFETY_CORE_FLAGS]);
  });

  // Reads the workflow command files, which live in the command packages.
  it.skip("keeps core command subjects revision-addressable", () => {
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
    // The 16-field context protocol the contract replaced must not return,
    // including in the plugin README a user reads before the guide.
    const readme = readRepositoryFile(
      "packages/command/command-workflow/README.md",
    );
    for (const source of [handoffs, readme]) {
      expect(source).not.toMatch(
        /Context digest|Context version|Audience:|Visibility:/u,
      );
    }
    expect(readme).not.toMatch(/Context has stable identity/u);
  });

  // Reads the workflow command files, which live in the command packages.
  it.skip("exposes revision and retry protection for mutating provider operations", () => {
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
      readRepositoryFile(
        `packages/skill/skill-workflow/workflow-guide/${file}`,
      ).includes("untrusted data"),
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

  it("keeps SDLC composition free of a fixed project inventory", () => {
    const sdlc = readRepositoryFile(
      "packages/skill/skill-workflow/workflow-guide/references/sdlc.md",
    );

    // Derived from the catalog, not listed here: a guide added tomorrow is
    // banned from this file without anyone remembering to extend a literal.
    for (const guide of installedGuideNames()) {
      expect(sdlc).not.toContain(guide);
    }
    expect(sdlc).toContain("capability-selection.md");
    // The 22-phase role matrix the rewrite deleted must not return.
    expect(sdlc).not.toContain("Incident commander");
    expect(sdlc).not.toContain("| Phase ");
  });

  it("keeps the architecture contract to its clauses", () => {
    const contract = readRepositoryFile(
      "packages/skill/skill-workflow/workflow-guide/references/contract.md",
    );

    // The frozen scenarios are a planning record; the skill keeps the clauses
    // that every operation obeys and nothing that dates on a new project.
    expect(contract).toContain("## Clauses");
    expect(contract).not.toContain("## Frozen Scenarios");
    for (const scenario of ["XP1", "DR1", "NG1", "IO1"]) {
      expect(contract).not.toContain(scenario);
    }
  });

  // Reads the workflow command files, which live in the command packages.
  it.skip("agrees on the default effect mode across skill and commands", () => {
    const effects = readRepositoryFile(
      "packages/skill/skill-workflow/workflow-guide/references/effects.md",
    );
    // inline was a result-placement label, not an external-state mode.
    expect(effects).not.toContain("`inline`");

    // Both sides are read from the files. The test asserts they agree rather
    // than restating the default, so a table edit that misses a command fails
    // here instead of passing on two literals that were updated together.
    const table = effectDefaultsFromSkill(effects);
    expect(table.size).toBe(
      coreWorkflowCommands.length + workspaceCommands.length,
    );

    for (const [command, expected] of table) {
      const documented = effectDefaultFromCommand(command);
      expect(
        {command, effect: documented},
        `${command} must document the default effects.md assigns`,
      ).toEqual({command, effect: expected});
    }
    // Every mode the table names is a mode the mode table defines.
    for (const mode of new Set(table.values())) {
      if (mode === undefined) continue;
      expect(effects).toContain(`| \`${mode}\``);
    }
  });

  // Reads the workflow command files, which live in the command packages.
  it.skip("lets every operation read the subject it is given", () => {
    // A provider-native subject or --context reference is read through a
    // provider CLI, so withholding Bash would block the operation outright.
    for (const command of [...coreWorkflowCommands, ...workspaceCommands]) {
      expect(
        readRepositoryFile(
          `packages/command/command-workflow/commands/${command}.md`,
        ),
      ).toMatch(/^ {2}- Bash$/mu);
    }
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
      "packages/skill/skill-plan/plan-guide/references/continue.md",
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
