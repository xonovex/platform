import {existsSync, readFileSync} from "node:fs";
import {join, relative, resolve} from "node:path";
import {z} from "zod";
import {type CommandArgument} from "./argument-hint.js";
import {type CommandDocument} from "./command-document.js";
import {validateCommandPackage} from "./command-validation.js";
import {
  issue,
  type ValidationIssue,
  type ValidationReport,
} from "./validation.js";

const ArgumentSchema = z.strictObject({
  name: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u),
  kind: z.enum(["flag", "positional"]),
  required: z.boolean(),
  repeatable: z.boolean(),
  valueName: z
    .string()
    .regex(/^[^\s<>]+$/u)
    .optional(),
});

const CommandSchema = z.strictObject({
  name: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u),
  operation: z.string().min(1),
  family: z.enum(["core", "workspace"]),
  reference: z.string().regex(/^references\/[a-z0-9-]+\.md$/u),
  effectModes: z.array(z.enum(["inspect", "preview", "apply"])).min(1),
  toolProfile: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u),
  argumentSets: z.array(z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u)),
  arguments: z.array(ArgumentSchema),
});

const WorkflowContractSchema = z.strictObject({
  contractVersion: z.literal(1),
  namespace: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u),
  delegation: z.strictObject({
    packageName: z.string().regex(/^@xonovex\/skill-[a-z0-9-]+$/u),
    plugin: z.string().regex(/^xonovex-skill-[a-z0-9-]+$/u),
    skill: z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u),
    skillDirectory: z.string().min(1),
  }),
  forbiddenArguments: z.array(z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u)),
  argumentSets: z.record(
    z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u),
    z.array(ArgumentSchema).min(1),
  ),
  toolProfiles: z.record(
    z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/u),
    z.array(z.string().min(1)).min(1),
  ),
  commands: z.array(CommandSchema).min(1),
});

export type WorkflowContract = z.infer<typeof WorkflowContractSchema>;
type WorkflowCommandContract = WorkflowContract["commands"][number] & {
  readonly arguments: readonly CommandArgument[];
  readonly allowedTools: readonly string[];
};

const EXPECTED_COMMANDS = new Set([
  "abandon",
  "create",
  "decide",
  "execute",
  "publish",
  "review",
  "revise",
  "validate",
  "workspace-abandon",
  "workspace-cleanup",
  "workspace-create",
  "workspace-merge",
]);

const argumentKey = (argument: CommandArgument): string =>
  `${argument.kind}:${argument.name}`;

const compareArguments = (
  expected: readonly CommandArgument[],
  actual: readonly CommandArgument[],
  path: string,
  issues: ValidationIssue[],
): void => {
  const expectedByKey = new Map(
    expected.map((argument) => [argumentKey(argument), argument]),
  );
  const actualByKey = new Map(
    actual.map((argument) => [argumentKey(argument), argument]),
  );
  for (const [key, argument] of expectedByKey) {
    const found = actualByKey.get(key);
    if (found === undefined) {
      issues.push(
        issue(
          "workflow.argument-missing",
          path,
          `argument '${argument.name}' is missing from argument-hint`,
        ),
      );
    } else if (
      found.required !== argument.required ||
      found.repeatable !== argument.repeatable ||
      found.valueName !== argument.valueName
    ) {
      issues.push(
        issue(
          "workflow.argument-shape",
          path,
          `argument '${argument.name}' does not match required/repeatable/value contract`,
        ),
      );
    }
  }
  for (const [key, argument] of actualByKey) {
    if (!expectedByKey.has(key)) {
      issues.push(
        issue(
          "workflow.argument-unexpected",
          path,
          `argument '${argument.name}' is not declared by the workflow contract`,
        ),
      );
    }
  }
};

const validateCommand = (
  contract: WorkflowContract,
  expected: WorkflowCommandContract,
  document: CommandDocument,
  skillDirectory: string,
  issues: ValidationIssue[],
): void => {
  if (document.namespace !== contract.namespace) {
    issues.push(
      issue(
        "workflow.namespace",
        document.path,
        `expected namespace '${contract.namespace}'`,
      ),
    );
  }
  const delegation = document.delegation;
  if (
    delegation?.plugin !== contract.delegation.plugin ||
    delegation.skill !== contract.delegation.skill ||
    delegation.operation.toLowerCase() !== expected.operation.toLowerCase()
  ) {
    issues.push(
      issue(
        "workflow.delegation",
        document.path,
        `expected ${contract.delegation.skill}/${expected.operation} delegation`,
      ),
    );
  }
  compareArguments(
    expected.arguments,
    document.arguments,
    document.path,
    issues,
  );
  const expectedTools = new Set(expected.allowedTools);
  const actualTools = new Set(document.frontmatter.allowedTools);
  if (
    expectedTools.size !== actualTools.size ||
    [...expectedTools].some((tool) => !actualTools.has(tool))
  ) {
    issues.push(
      issue(
        "workflow.tool-profile",
        document.path,
        `allowed-tools must match profile '${expected.toolProfile}'`,
      ),
    );
  }
  for (const argument of expected.arguments) {
    if (!document.documentedArguments.has(argument.name)) {
      issues.push(
        issue(
          "workflow.argument-undocumented",
          document.path,
          `argument '${argument.name}' is not documented in Arguments`,
        ),
      );
    }
  }
  for (const forbidden of contract.forbiddenArguments) {
    if (document.arguments.some(({name}) => name === forbidden)) {
      issues.push(
        issue(
          "workflow.argument-forbidden",
          document.path,
          `forbidden global argument '${forbidden}' is exposed`,
        ),
      );
    }
  }
  const referencePath = join(skillDirectory, expected.reference);
  if (!existsSync(referencePath)) {
    issues.push(
      issue(
        "workflow.reference-missing",
        document.path,
        `skill reference '${expected.reference}' does not exist`,
      ),
    );
  }
};

const sameModes = (
  actual: readonly string[],
  expected: readonly string[],
): boolean =>
  actual.length === expected.length &&
  actual.every((mode, index) => mode === expected[index]);

const expectedEffectModes = (command: string): readonly string[] => {
  if (command === "execute") return ["inspect", "preview", "apply"];
  if (
    [
      "publish",
      "workspace-cleanup",
      "workspace-create",
      "workspace-merge",
    ].includes(command)
  ) {
    return ["preview", "apply"];
  }
  return ["inspect"];
};

const expectedToolProfile = (command: string): string => {
  if (["execute", "publish"].includes(command)) return "effectful";
  if (
    ["workspace-cleanup", "workspace-create", "workspace-merge"].includes(
      command,
    )
  ) {
    return "workspace-effect";
  }
  if (command === "workspace-abandon") return "inline-workspace";
  return "inline";
};

const validateCommandArchitecture = (
  command: WorkflowCommandContract,
  path: string,
  issues: ValidationIssue[],
): void => {
  const expectedModes = expectedEffectModes(command.name);
  if (!sameModes(command.effectModes, expectedModes)) {
    issues.push(
      issue(
        "workflow.effect-modes",
        path,
        `${command.name} must use effect modes ${expectedModes.join("|")}`,
      ),
    );
  }
  if (command.toolProfile !== expectedToolProfile(command.name)) {
    issues.push(
      issue(
        "workflow.tool-profile-selection",
        path,
        `${command.name} must use '${expectedToolProfile(command.name)}'`,
      ),
    );
  }
  const argumentNames = new Set(command.arguments.map(({name}) => name));
  if (command.effectModes.length > 1 && !argumentNames.has("effect")) {
    issues.push(
      issue(
        "workflow.effect-argument",
        path,
        `${command.name} must expose its effect mode`,
      ),
    );
  }
  if (
    command.name !== "publish" &&
    ["destination", "result"].some((name) => argumentNames.has(name))
  ) {
    issues.push(
      issue(
        "workflow.publish-separation",
        path,
        `${command.name} cannot expose an external result destination`,
      ),
    );
  }
  if (
    command.name === "publish" &&
    ["destination-provider", "destination-reference"].some(
      (name) => !argumentNames.has(name),
    )
  ) {
    issues.push(
      issue(
        "workflow.publish-destination",
        path,
        "publish must expose an explicit provider-bound destination",
      ),
    );
  }
  const perspective = command.arguments.find(
    ({name}) => name === "perspective",
  );
  if (perspective !== undefined && !perspective.repeatable) {
    issues.push(
      issue(
        "workflow.perspective-repeatable",
        path,
        `${command.name} perspective must be repeatable`,
      ),
    );
  }
  if (
    ["workspace-abandon", "workspace-merge"].includes(command.name) &&
    ["remove", "remove-reference", "prune"].some((name) =>
      argumentNames.has(name),
    )
  ) {
    issues.push(
      issue(
        "workflow.cleanup-separation",
        path,
        `${command.name} cannot remove workspace resources`,
      ),
    );
  }
  if (command.arguments.every(({name}) => name !== "request")) {
    issues.push(
      issue(
        "workflow.request-binding",
        path,
        `${command.name} must accept an advanced request file`,
      ),
    );
  }
};

const resolveCommands = (
  contract: WorkflowContract,
  path: string,
  issues: ValidationIssue[],
): readonly WorkflowCommandContract[] =>
  contract.commands.map((command) => {
    const inherited = command.argumentSets.flatMap((setName) => {
      const argumentSet = contract.argumentSets[setName];
      if (argumentSet === undefined) {
        issues.push(
          issue(
            "workflow.argument-set-missing",
            path,
            `${command.name} names missing argument set '${setName}'`,
          ),
        );
        return [];
      }
      return argumentSet;
    });
    const argumentsFound = [...inherited, ...command.arguments];
    const argumentKeys = argumentsFound.map(argumentKey);
    if (new Set(argumentKeys).size !== argumentKeys.length) {
      issues.push(
        issue(
          "workflow.argument-duplicate",
          path,
          `${command.name} resolves duplicate arguments`,
        ),
      );
    }
    const allowedTools = contract.toolProfiles[command.toolProfile];
    if (allowedTools === undefined) {
      issues.push(
        issue(
          "workflow.tool-profile-missing",
          path,
          `${command.name} names missing tool profile '${command.toolProfile}'`,
        ),
      );
    }
    return {
      ...command,
      arguments: argumentsFound,
      allowedTools: allowedTools ?? [],
    };
  });

const validateArchitecture = (
  contract: WorkflowContract,
  commands: readonly WorkflowCommandContract[],
  path: string,
  issues: ValidationIssue[],
): void => {
  for (const forbidden of [
    "capability",
    "confirm",
    "dry-run",
    "kind",
    "provider",
    "reference",
    "result",
  ]) {
    if (!contract.forbiddenArguments.includes(forbidden)) {
      issues.push(
        issue(
          "workflow.forbidden-arguments",
          path,
          `contract must forbid legacy global argument '${forbidden}'`,
        ),
      );
    }
  }
  for (const [profile, tools] of Object.entries(contract.toolProfiles)) {
    const uniqueTools = new Set(tools);
    if (uniqueTools.size !== tools.length) {
      issues.push(
        issue(
          "workflow.tool-profile-duplicate",
          path,
          `tool profile '${profile}' contains duplicates`,
        ),
      );
    }
    if (!uniqueTools.has("Skill")) {
      issues.push(
        issue(
          "workflow.tool-profile-skill",
          path,
          `tool profile '${profile}' must include Skill`,
        ),
      );
    }
    if (
      ["inline", "inline-workspace"].includes(profile) &&
      ["Write", "Edit", "Bash"].some((tool) => uniqueTools.has(tool))
    ) {
      issues.push(
        issue(
          "workflow.tool-profile-inline-effect",
          path,
          `tool profile '${profile}' cannot grant Write, Edit, or Bash`,
        ),
      );
    }
    if (
      profile === "workspace-effect" &&
      (uniqueTools.has("Write") || uniqueTools.has("Edit"))
    ) {
      issues.push(
        issue(
          "workflow.tool-profile-workspace-write",
          path,
          "workspace-effect cannot grant Write or Edit",
        ),
      );
    }
  }
  const names = commands.map(({name}) => name);
  const uniqueNames = new Set(names);
  const inventoryComplete = [...EXPECTED_COMMANDS].every((name) =>
    uniqueNames.has(name),
  );
  if (!inventoryComplete || uniqueNames.size !== EXPECTED_COMMANDS.size) {
    issues.push(
      issue(
        "workflow.inventory",
        path,
        "contract must declare the eight core and four workspace commands exactly once",
      ),
    );
  }
  if (uniqueNames.size !== names.length) {
    issues.push(
      issue(
        "workflow.inventory-duplicate",
        path,
        "contract command names must be unique",
      ),
    );
  }
  for (const command of commands) {
    validateCommandArchitecture(command, path, issues);
  }
};

export const readWorkflowContract = (path: string): WorkflowContract => {
  let input: unknown;
  try {
    input = JSON.parse(readFileSync(path, "utf8"));
  } catch (error) {
    throw new Error(`cannot parse workflow contract JSON at ${path}`, {
      cause: error,
    });
  }
  const parsed = WorkflowContractSchema.safeParse(input);
  if (!parsed.success) {
    throw new Error(
      `invalid workflow contract at ${path}: ${z.prettifyError(parsed.error)}`,
    );
  }
  return parsed.data;
};

export const validateWorkflowContract = (
  contractPath: string,
  packageDir: string,
  repositoryRoot = resolve(packageDir, "../../.."),
): ValidationReport => {
  const contract = readWorkflowContract(contractPath);
  const packageValidation = validateCommandPackage(packageDir, repositoryRoot);
  const issues = [...packageValidation.report.issues];
  const displayContractPath = relative(repositoryRoot, contractPath);
  const commands = resolveCommands(contract, displayContractPath, issues);
  validateArchitecture(contract, commands, displayContractPath, issues);
  const contractByName = new Map(
    commands.map((command) => [command.name, command]),
  );
  const documentByName = new Map(
    packageValidation.documents.map((document) => [
      document.fileName,
      document,
    ]),
  );
  const skillDirectory = resolve(
    repositoryRoot,
    contract.delegation.skillDirectory,
  );
  for (const [name, expected] of contractByName) {
    const document = documentByName.get(name);
    if (document === undefined) {
      issues.push(
        issue(
          "workflow.command-missing",
          displayContractPath,
          `command '${name}' does not exist`,
        ),
      );
      continue;
    }
    validateCommand(contract, expected, document, skillDirectory, issues);
  }
  for (const name of documentByName.keys()) {
    if (!contractByName.has(name)) {
      issues.push(
        issue(
          "workflow.command-unexpected",
          displayContractPath,
          `command '${name}' is not declared by the contract`,
        ),
      );
    }
  }
  return {
    commands: packageValidation.report.commands,
    issues,
  };
};
