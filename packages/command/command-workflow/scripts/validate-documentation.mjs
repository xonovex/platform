#!/usr/bin/env node
import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {basename, dirname, extname, relative, resolve, sep} from "node:path";
import {fileURLToPath} from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const packageDirectory = resolve(scriptDirectory, "..");
const repositoryRoot = resolve(packageDirectory, "../../..");

export const EXPECTED_CORE_COMMANDS = Object.freeze([
  "abandon",
  "create",
  "decide",
  "execute",
  "publish",
  "review",
  "revise",
  "validate",
]);

export const EXPECTED_WORKSPACE_COMMANDS = Object.freeze([
  "workspace-abandon",
  "workspace-cleanup",
  "workspace-create",
  "workspace-merge",
]);

export const EXPECTED_COMMANDS = new Set([
  ...EXPECTED_CORE_COMMANDS,
  ...EXPECTED_WORKSPACE_COMMANDS,
]);

export const FORMER_COMMANDS = Object.freeze([
  "acceptance-decide",
  "acceptance-validate",
  "assessment-run",
  "corrective-action-run",
  "decision-accept",
  "decision-create",
  "decision-critique",
  "decision-revise",
  "deliver-publish",
  "develop-abandon",
  "develop-consolidate",
  "develop-run",
  "discovery-run",
  "experience-design-accept",
  "experience-design-create",
  "experience-design-critique",
  "experience-design-revise",
  "formulation-run",
  "git-commit",
  "incident-run",
  "integration-run",
  "integration-validate",
  "inventory-generate",
  "observe-run",
  "plan-accept",
  "plan-continue",
  "plan-create",
  "plan-critique",
  "plan-reject",
  "plan-research",
  "plan-revise",
  "plan-subplans-create",
  "plan-update",
  "plan-validate",
  "plan-worktree-abandon",
  "plan-worktree-cleanup",
  "plan-worktree-create",
  "plan-worktree-merge",
  "pr-create",
  "pr-review-analyze",
  "pr-review-post",
  "pr-review-refine",
  "pr-review-resolve",
  "qa-run",
  "release-run",
  "research-run",
  "retirement-run",
  "review-run",
  "solution-design-accept",
  "solution-design-create",
  "solution-design-critique",
  "solution-design-revise",
  "transition-run",
]);

const REQUIRED_PROMPT_SECTIONS = Object.freeze([
  "Goal",
  "Arguments",
  "Core Workflow",
  "Implementation",
  "Error Handling",
]);

const ACTIVE_TEXT_EXTENSIONS = new Set([
  ".c",
  ".cjs",
  ".dot",
  ".go",
  ".h",
  ".js",
  ".json",
  ".md",
  ".mjs",
  ".ts",
  ".tsx",
  ".yaml",
  ".yml",
]);

const SEMANTIC_RELEASE_ALLOWLIST = new Map([
  [
    "packages/command/command-workflow/CHANGELOG.md",
    new Set(["agent-trigger-api", "former-command"]),
  ],
]);

const escapeRegExp = (value) => value.replace(/[.*+?^${}()|[\]\\]/gu, "\\$&");

const formerCommandNames = FORMER_COMMANDS.map(escapeRegExp).join("|");

const semanticRules = Object.freeze([
  {
    id: "profile-flag",
    message: "removed --profile command mode",
    pattern: /(^|[\s`])--profile(?:\s|=|`|$)/imu,
  },
  {
    id: "former-command",
    message: "former workflow command invocation or file reference",
    pattern: new RegExp(
      `(?:/xonovex-workflow:|commands/)(?:${formerCommandNames})(?:\\.md)?(?:\\b|$)`,
      "iu",
    ),
  },
  {
    id: "workflow-runtime",
    message: "removed workflow-runtime identifier",
    pattern:
      /\b(?:xonovex-(?:skill-)?workflow-runtime|workflow-runtime-guide|workflow-runtime-plugin)\b/iu,
  },
  {
    id: "agent-trigger-api",
    message: "removed trigger or schedule API",
    pattern:
      /\b(?:AgentTrigger|AgentSchedule|agenttriggers?\.agent\.xonovex\.com|agentschedules?\.agent\.xonovex\.com|trigger receiver|trigger service)\b/iu,
  },
  {
    id: "trigger-executor-mode",
    message: "trigger or executor encoded as a command mode",
    pattern:
      /--(?:trigger|executor)(?:\b|=)|--mode(?:\s+|=)(?:trigger|executor)\b/iu,
    compositionOnly: true,
  },
  {
    id: "lifecycle-gate-claim",
    message: "command-enforced lifecycle, gate, or handoff claim",
    pattern:
      /\b(?:enforces?|requires?|mandates?) (?:an? |the )?(?:delivery )?(?:lifecycle|approval gate|governed tail|required handoff)\b/iu,
    compositionOnly: true,
  },
  {
    id: "central-resolution-claim",
    message: "central provider or reference resolution claim",
    pattern:
      /\b(?:uses?|requires?|through|via) (?:an? )?central (?:provider|reference) (?:resolver|resolution|schema|store)\b/iu,
    compositionOnly: true,
  },
  {
    id: "role-specific-api",
    message: "role-specific command API claim",
    pattern:
      /\b(?:provides?|requires?|selects?) (?:an? )?role-specific (?:command|API)|\b(?:PM|PO|UX|QA|developer)[- ]only command\b/iu,
    compositionOnly: true,
  },
  {
    id: "maturity-enforcement",
    message: "A1/A2/A3 label used as enforced command policy",
    pattern:
      /\bA[123]\b[^\n]{0,120}\b(?:authorizes?|cannot|forbids?|must|only|permits?|requires?|required)\b/iu,
    compositionOnly: true,
  },
]);

const isCompositionSurface = (path) =>
  /^(?:packages\/(?:command\/command-workflow|skill\/skill-plan|agent\/agent-operator-go|diagram\/diagram-agent-workflow)\/|\.agents\/plugins\/marketplace\.json$|\.claude-plugin\/marketplace\.json$)/u.test(
    path,
  );

const createValidation = () => {
  const failures = [];
  let checks = 0;
  return {
    check: (condition, message) => {
      checks += 1;
      if (!condition) failures.push(message);
    },
    result: () => ({checks, failures}),
  };
};

const read = (path) => readFileSync(path, "utf8");
const readJson = (path) => JSON.parse(read(path));
const relativePath = (path) => relative(packageDirectory, path);

const markdownFiles = (path) => {
  if (!existsSync(path)) return [];
  const entries = readdirSync(path, {withFileTypes: true});
  return entries.flatMap((entry) => {
    const entryPath = resolve(path, entry.name);
    if (entry.isDirectory()) return markdownFiles(entryPath);
    return extname(entry.name) === ".md" ? [entryPath] : [];
  });
};

const sameStrings = (left, right) => {
  const sortedLeft = [...left].toSorted();
  const sortedRight = [...right].toSorted();
  return (
    sortedLeft.length === sortedRight.length &&
    sortedLeft.every((value, index) => value === sortedRight[index])
  );
};

const numberWord = (value) => {
  const words = new Map([
    [4, "four"],
    [8, "eight"],
    [12, "twelve"],
  ]);
  return words.get(value) ?? String(value);
};

const commandTitle = (body) =>
  /^# \/xonovex-workflow:([a-z0-9-]+) — \S.+$/mu.exec(body)?.[1];

const parsePrompt = (content) => {
  const match = /^---\n(?<frontmatter>[\s\S]*?)\n---\n(?<body>[\s\S]*)$/u.exec(
    content,
  );
  if (match?.groups === undefined) return undefined;
  return {
    frontmatter: match.groups.frontmatter ?? "",
    body: match.groups.body ?? "",
  };
};

export const readCommandDirectory = (commandDirectory) =>
  new Map(
    readdirSync(commandDirectory, {withFileTypes: true})
      .filter((entry) => entry.isFile() && extname(entry.name) === ".md")
      .toSorted((left, right) => left.name.localeCompare(right.name))
      .map((entry) => [
        entry.name,
        read(resolve(commandDirectory, entry.name)),
      ]),
  );

export const validateCommandInventory = (commandFiles) => {
  const validation = createValidation();
  const actualFiles = new Set(commandFiles.keys());
  const expectedFiles = new Set(
    [...EXPECTED_COMMANDS].map((name) => `${name}.md`),
  );
  const missing = [...expectedFiles]
    .filter((name) => !actualFiles.has(name))
    .toSorted();
  const extra = [...actualFiles]
    .filter((name) => !expectedFiles.has(name))
    .toSorted();

  validation.check(
    missing.length === 0,
    `command inventory is missing: ${missing.join(", ")}`,
  );
  validation.check(
    extra.length === 0,
    `command inventory has extra files: ${extra.join(", ")}`,
  );

  const commandNames = [];
  for (const [fileName, content] of commandFiles) {
    const prompt = parsePrompt(content);
    validation.check(
      prompt !== undefined,
      `${fileName} has delimited YAML frontmatter and a prompt body`,
    );
    if (prompt === undefined) continue;

    const description = /^description:\s+(.+)$/mu.exec(prompt.frontmatter)?.[1];
    validation.check(
      description !== undefined && description.trim().length > 0,
      `${fileName} has a non-empty single-line description`,
    );
    validation.check(
      /^allowed-tools:\s*\n(?: {2}- \S.*\n?)+/mu.test(prompt.frontmatter),
      `${fileName} declares a non-empty allowed-tools list`,
    );
    validation.check(
      /^argument-hint:\s*>-\s*\n {2}\S/mu.test(prompt.frontmatter),
      `${fileName} declares a non-empty folded argument hint`,
    );

    const name = commandTitle(prompt.body);
    validation.check(
      name !== undefined,
      `${fileName} has a /xonovex-workflow command title`,
    );
    if (name !== undefined) {
      commandNames.push(name);
      validation.check(
        fileName === `${name}.md`,
        `${fileName} command name matches its file name`,
      );
    }
    validation.check(
      prompt.body.trim().length > 0,
      `${fileName} has a non-empty prompt body`,
    );
    for (const section of REQUIRED_PROMPT_SECTIONS) {
      validation.check(
        prompt.body.includes(`\n## ${section}\n`),
        `${fileName} has a ${section} section`,
      );
    }
  }

  const duplicates = commandNames.filter(
    (name, index) => commandNames.indexOf(name) !== index,
  );
  validation.check(
    duplicates.length === 0,
    `command names are unique; duplicates: ${[...new Set(duplicates)].join(", ")}`,
  );
  return validation.result();
};

export const validateDisplayedInventory = (readme, roleLenses) => {
  const validation = createValidation();
  const totalWord = numberWord(EXPECTED_COMMANDS.size);
  const coreWord = numberWord(EXPECTED_CORE_COMMANDS.length);
  const workspaceWord = numberWord(EXPECTED_WORKSPACE_COMMANDS.length);
  const totalLabel = `${totalWord[0]?.toUpperCase() ?? ""}${totalWord.slice(1)}`;
  const linkedCommands = [
    ...readme.matchAll(/\(commands\/([a-z0-9-]+)\.md\)/gu),
  ]
    .map((match) => match[1])
    .filter((name) => name !== undefined);

  validation.check(
    new RegExp(
      `(?:^|\\n\\n)${totalLabel} independently invocable commands`,
      "u",
    ).test(readme),
    `README displays the derived ${String(EXPECTED_COMMANDS.size)}-command total`,
  );
  validation.check(
    readme.includes(`The ${coreWord} core operations are siblings.`),
    `README displays the derived ${String(EXPECTED_CORE_COMMANDS.length)}-operation count`,
  );
  validation.check(
    readme.includes(`One of the ${coreWord} command verbs.`),
    "README operation dimension uses the derived core-operation count",
  );
  validation.check(
    readme.includes(`The ${workspaceWord} workspace utilities manage`),
    `README displays the derived ${String(EXPECTED_WORKSPACE_COMMANDS.length)}-utility count`,
  );
  validation.check(
    roleLenses.includes(`the same ${coreWord} operations.`),
    "role lenses use the derived core-operation count",
  );
  validation.check(
    sameStrings(linkedCommands, [...EXPECTED_COMMANDS]) &&
      new Set(linkedCommands).size === linkedCommands.length,
    "README links every expected command exactly once",
  );
  return validation.result();
};

export const validateManifestDependencies = (
  workflowPackage,
  workflowClaudeManifest,
  workflowCodexManifest,
) => {
  const validation = createValidation();
  const claudeDependencies = workflowClaudeManifest.dependencies;
  const codexDependencies = workflowCodexManifest.dependencies;
  const packageDependencies = Object.keys(workflowPackage.dependencies ?? {});
  validation.check(
    Array.isArray(claudeDependencies) &&
      claudeDependencies.every((dependency) => typeof dependency === "string"),
    "Claude manifest dependencies are a string list",
  );
  validation.check(
    Array.isArray(codexDependencies) &&
      codexDependencies.every((dependency) => typeof dependency === "string"),
    "Codex manifest dependencies are a string list",
  );
  if (!Array.isArray(claudeDependencies) || !Array.isArray(codexDependencies)) {
    return validation.result();
  }
  validation.check(
    sameStrings(claudeDependencies, codexDependencies),
    "workflow plugin dependencies match across harness manifests",
  );
  validation.check(
    claudeDependencies.length === 0,
    "workflow plugin has an empty universal dependency set",
  );
  validation.check(
    packageDependencies.length === 0,
    "workflow package has no universal skill dependencies",
  );
  return validation.result();
};

export const validateLocalMarkdownLinks = (files, rootDirectory) => {
  const validation = createValidation();
  for (const file of files) {
    const content = read(file);
    for (const match of content.matchAll(/\[[^\]]+\]\(([^)]+)\)/gu)) {
      const rawTarget = match[1];
      if (rawTarget === undefined) continue;
      const target = rawTarget.split("#", 1)[0];
      if (!target || /^(?:https?:|mailto:)/u.test(target)) continue;
      const resolvedTarget = resolve(dirname(file), target);
      const rootRelativeTarget = relative(rootDirectory, resolvedTarget);
      if (
        rootRelativeTarget === ".." ||
        rootRelativeTarget.startsWith(`..${sep}`)
      ) {
        continue;
      }
      validation.check(
        existsSync(resolvedTarget),
        `${relativePath(file)} link resolves: ${target}`,
      );
    }
  }
  return validation.result();
};

const activeSurfaceFiles = (root) => {
  if (!existsSync(root)) return [];
  return readdirSync(root, {withFileTypes: true}).flatMap((entry) => {
    const entryPath = resolve(root, entry.name);
    if (entry.isDirectory()) {
      if ([".moon", "coverage", "dist", "node_modules"].includes(entry.name)) {
        return [];
      }
      if (entryPath === resolve(packageDirectory, "scripts")) return [];
      return activeSurfaceFiles(entryPath);
    }
    return ACTIVE_TEXT_EXTENSIONS.has(extname(entry.name)) ? [entryPath] : [];
  });
};

export const validateSemanticResidue = (files) => {
  const failures = [];
  for (const [path, content] of files) {
    const allowlist = SEMANTIC_RELEASE_ALLOWLIST.get(path) ?? new Set();
    for (const rule of semanticRules) {
      if (rule.compositionOnly === true && !isCompositionSurface(path))
        continue;
      if (allowlist.has(rule.id)) continue;
      if (rule.pattern.test(content)) {
        failures.push(`${path} contains ${rule.message}`);
      }
    }
    const formerFileName = `${basename(path, extname(path))}.md`;
    if (
      /(?:^|\/)commands\/[^/]+\.md$/u.test(path) &&
      FORMER_COMMANDS.some((name) => `${name}.md` === formerFileName) &&
      !allowlist.has("former-command")
    ) {
      failures.push(`${path} is a former workflow command file`);
    }
  }
  return {checks: files.size * semanticRules.length, failures};
};

const appendFailures = (validation, result) => {
  validation.check(
    result.failures.length === 0,
    result.failures.length === 0
      ? "validation group passes"
      : result.failures.join("; "),
  );
};

const runDocumentationValidation = () => {
  const validation = createValidation();
  const commandDirectory = resolve(packageDirectory, "commands");
  const commandFiles = readCommandDirectory(commandDirectory);
  const workflowPackage = readJson(resolve(packageDirectory, "package.json"));
  const workflowClaudeManifest = readJson(
    resolve(packageDirectory, ".claude-plugin/plugin.json"),
  );
  const workflowCodexManifest = readJson(
    resolve(packageDirectory, ".codex-plugin/plugin.json"),
  );
  const documentationFiles = [
    resolve(packageDirectory, "README.md"),
    ...markdownFiles(resolve(packageDirectory, "docs")),
  ];
  const commandPaths = [...commandFiles.keys()].map((fileName) =>
    resolve(commandDirectory, fileName),
  );

  appendFailures(validation, validateCommandInventory(commandFiles));
  appendFailures(
    validation,
    validateDisplayedInventory(
      read(resolve(packageDirectory, "README.md")),
      read(resolve(packageDirectory, "docs/role-lenses.md")),
    ),
  );
  appendFailures(
    validation,
    validateManifestDependencies(
      workflowPackage,
      workflowClaudeManifest,
      workflowCodexManifest,
    ),
  );
  appendFailures(
    validation,
    validateLocalMarkdownLinks(
      [...documentationFiles, ...commandPaths],
      packageDirectory,
    ),
  );

  const publishedMarkdown = documentationFiles.map(read).join("\n");
  const forbiddenClaims = [
    /all harnesses (?:have|support|provide)/iu,
    /skills? (?:are|is|provide|provides) enforcement/iu,
    /install(?:ing|ed)? (?:a )?skills? (?:enforces|proves)/iu,
    /workflow ya?ml (?:is|are) required/iu,
    /(?:provides?|ensures?|establishes?|achieves?) automatic compliance/iu,
    /silently launch(?:es|ing)? (?:a |an )?(?:child )?agent/iu,
  ];
  for (const forbiddenClaim of forbiddenClaims) {
    validation.check(
      !forbiddenClaim.test(publishedMarkdown),
      `published docs reject ${String(forbiddenClaim)}`,
    );
  }

  const activeRoots = [
    resolve(repositoryRoot, "packages/command"),
    resolve(repositoryRoot, "packages/skill"),
    resolve(repositoryRoot, "packages/agent"),
    resolve(repositoryRoot, "packages/shared"),
    resolve(repositoryRoot, "packages/diagram"),
  ];
  const activeFiles = [
    ...activeRoots.flatMap(activeSurfaceFiles),
    resolve(repositoryRoot, ".claude-plugin/marketplace.json"),
    resolve(repositoryRoot, ".agents/plugins/marketplace.json"),
  ];
  const semanticFiles = new Map(
    activeFiles.map((file) => [relative(repositoryRoot, file), read(file)]),
  );
  appendFailures(validation, validateSemanticResidue(semanticFiles));

  const result = validation.result();
  if (result.failures.length > 0) {
    console.error(
      `Documentation validation failed: ${String(result.failures.length)}/${String(result.checks)} checks`,
    );
    for (const failure of result.failures) console.error(`- ${failure}`);
    process.exitCode = 1;
    return;
  }
  console.log(
    `Documentation validation passed: ${String(result.checks)} validation groups for ${String(EXPECTED_COMMANDS.size)} commands`,
  );
};

const invokedPath = process.argv[1];
if (
  invokedPath !== undefined &&
  resolve(invokedPath) === fileURLToPath(import.meta.url)
) {
  runDocumentationValidation();
}
