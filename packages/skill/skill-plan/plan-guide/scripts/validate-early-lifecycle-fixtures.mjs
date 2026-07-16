import {existsSync, readFileSync} from "node:fs";
import {dirname, join, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const scriptDirectory = dirname(fileURLToPath(import.meta.url));
const guideDirectory = resolve(scriptDirectory, "..");
const packageDirectory = resolve(guideDirectory, "..");
const repositoryDirectory = resolve(packageDirectory, "../../..");
const commandDirectory = join(
  repositoryDirectory,
  "packages/command/command-workflow/commands",
);

const fixturePath = join(
  guideDirectory,
  "assets/early-lifecycle-fixtures.json",
);
const fixtures = JSON.parse(readFileSync(fixturePath, "utf8"));

const allowedExecutors = new Set([
  "deterministic",
  "bounded-model",
  "adaptive-agent",
  "human",
  "qualified-human",
]);
const allowedPolicyIntents = new Set([
  "data-access",
  "external-research",
  "privacy",
  "accessibility",
  "security",
  "architecture",
  "regulated-applicability",
]);
const canonicalCapabilities = new Set([
  "Discovery",
  "Research",
  "Formulation",
  "ExperienceDesign",
  "SolutionDesign",
  "Decision",
  "Planning",
]);

const failures = [];

const fail = (message) => failures.push(message);

const readJson = (path) => JSON.parse(readFileSync(path, "utf8"));

const expectIncludes = (content, expected, context) => {
  if (!content.includes(expected)) {
    fail(`${context}: missing ${JSON.stringify(expected)}`);
  }
};

const classifyCase = (fixture) => {
  if (
    fixture.subjectRevision !== undefined &&
    fixture.provider?.nativeRevision !== fixture.subjectRevision
  ) {
    return "subject-revision-mismatch";
  }

  if (fixture.provider?.explicitlySelected && !fixture.provider.available) {
    return "explicit-provider-unavailable";
  }

  if (
    fixture.decisionAuthority &&
    !["human", "qualified-human"].includes(fixture.decisionAuthority.executor)
  ) {
    return "decision-authority-invalid";
  }

  if (!fixture.provider?.available) {
    return "provider-unavailable";
  }

  if (!fixture.provider.context || !fixture.provider.nativeReference) {
    return "provider-reference-missing";
  }

  if (
    fixture.resume?.conversationAvailable === false &&
    fixture.resume.resolvedFromNativeReference !== true
  ) {
    return "fresh-context-recovery-failed";
  }

  return null;
};

for (const contract of fixtures.operationContracts) {
  if (!fixtures.commandContracts.includes(contract.operation)) {
    fail(`operation ${contract.operation}: missing command contract`);
  }

  if (!canonicalCapabilities.has(contract.resultKind)) {
    fail(
      `operation ${contract.operation}: non-canonical result kind ${contract.resultKind}`,
    );
  }

  for (const executor of contract.preferredExecutors) {
    if (!allowedExecutors.has(executor)) {
      fail(`operation ${contract.operation}: unknown executor ${executor}`);
    }
  }

  for (const intent of contract.policyIntents) {
    if (!allowedPolicyIntents.has(intent)) {
      fail(`operation ${contract.operation}: unknown policy intent ${intent}`);
    }
  }
}

for (const command of fixtures.commandContracts) {
  const commandPath = join(commandDirectory, `${command}.md`);

  if (!existsSync(commandPath)) {
    fail(`command ${command}: file is missing`);
    continue;
  }

  const content = readFileSync(commandPath, "utf8");
  expectIncludes(content, "Skill", `command ${command}`);
  expectIncludes(content, "plan-guide", `command ${command}`);
  expectIncludes(content, "workflow-guide", `command ${command}`);
  expectIncludes(content, `**${command}**`, `command ${command}`);
}

for (const command of [
  "experience-design-revise",
  "experience-design-critique",
  "experience-design-accept",
  "solution-design-revise",
  "solution-design-critique",
  "solution-design-accept",
  "decision-revise",
  "decision-critique",
  "decision-accept",
]) {
  const content = readFileSync(join(commandDirectory, `${command}.md`), "utf8");
  expectIncludes(content, "--revision", `exact-revision command ${command}`);
  expectIncludes(content, "--provider", `provider command ${command}`);
}

for (const removedCommand of [
  "story-refine.md",
  "acceptance-formalize.md",
  "plan-decide.md",
]) {
  if (existsSync(join(commandDirectory, removedCommand))) {
    fail(`replaced command still exists: ${removedCommand}`);
  }
}

const commandManifests = [
  join(
    repositoryDirectory,
    "packages/command/command-workflow/.claude-plugin/plugin.json",
  ),
  join(
    repositoryDirectory,
    "packages/command/command-workflow/.codex-plugin/plugin.json",
  ),
];
const planManifests = [
  join(packageDirectory, ".claude-plugin/plugin.json"),
  join(packageDirectory, ".codex-plugin/plugin.json"),
];

for (const manifestPath of commandManifests) {
  const dependencies = new Set(readJson(manifestPath).dependencies ?? []);

  for (const dependency of fixtures.dependencyPolicy.requiredHardDependencies) {
    if (!dependencies.has(dependency)) {
      fail(`${manifestPath}: missing required dependency ${dependency}`);
    }
  }

  for (const dependency of fixtures.dependencyPolicy
    .forbiddenHardDependencies) {
    if (dependencies.has(dependency)) {
      fail(`${manifestPath}: forbidden hard dependency ${dependency}`);
    }
  }
}

for (const manifestPath of planManifests) {
  const dependencies = new Set(readJson(manifestPath).dependencies ?? []);

  if (!dependencies.has("xonovex-skill-workflow")) {
    fail(`${manifestPath}: workflow contract dependency is missing`);
  }

  for (const dependency of fixtures.dependencyPolicy
    .forbiddenHardDependencies) {
    if (dependencies.has(dependency)) {
      fail(`${manifestPath}: forbidden hard dependency ${dependency}`);
    }
  }
}

const requiredCaseIds = new Set([
  "lightweight-neutral-workflow",
  "optional-experience-design",
  "optional-solution-design",
  "regulated-ai-qualified-decision",
  "ai-advisory-bounded-exploration",
  "regulated-qualified-decision",
  "fresh-context-provider-recovery",
  "selectable-bdd-formulation-method",
  "exact-revision-mismatch",
  "explicit-provider-unavailable",
  "model-fabricates-authority",
]);

for (const fixture of fixtures.cases) {
  requiredCaseIds.delete(fixture.id);

  for (const capability of fixture.capabilities) {
    if (!canonicalCapabilities.has(capability)) {
      fail(`case ${fixture.id}: unknown capability ${capability}`);
    }
  }

  for (const executor of fixture.executors) {
    if (!allowedExecutors.has(executor)) {
      fail(`case ${fixture.id}: unknown executor ${executor}`);
    }
  }

  for (const intent of fixture.policyIntents) {
    if (!allowedPolicyIntents.has(intent)) {
      fail(`case ${fixture.id}: unknown policy intent ${intent}`);
    }
  }

  const code = classifyCase(fixture);

  if (fixture.valid && code !== null) {
    fail(`case ${fixture.id}: expected valid, received ${code}`);
  }

  if (!fixture.valid && code !== fixture.expectedCode) {
    fail(
      `case ${fixture.id}: expected ${fixture.expectedCode}, received ${code ?? "valid"}`,
    );
  }
}

for (const missingCaseId of requiredCaseIds) {
  fail(`required fixture case is missing: ${missingCaseId}`);
}

const sharedContract = readFileSync(
  join(guideDirectory, "references/early-lifecycle-contracts.md"),
  "utf8",
);

for (const executor of allowedExecutors) {
  expectIncludes(sharedContract, executor, "early lifecycle executor contract");
}

for (const intent of allowedPolicyIntents) {
  expectIncludes(sharedContract, intent, "early lifecycle policy contract");
}

for (const planningReference of [
  "plan-create.md",
  "plan-revise.md",
  "plan-critique.md",
  "plan-accept.md",
  "plan-subplans-create.md",
  "plan-continue.md",
  "plan-update.md",
  "plan-validate.md",
]) {
  const content = readFileSync(
    join(guideDirectory, "references", planningReference),
    "utf8",
  );
  expectIncludes(content, "native", `planning reference ${planningReference}`);
  expectIncludes(
    content,
    "provider",
    `planning reference ${planningReference}`,
  );
}

const onboarding = readFileSync(
  join(guideDirectory, "references/lifecycle-onboard-advise.md"),
  "utf8",
);
for (const concept of [
  "methods",
  "skills",
  "providers",
  "executor",
  "environment modules",
]) {
  expectIncludes(onboarding, concept, "lifecycle onboarding contract");
}

if (failures.length > 0) {
  console.error("Early lifecycle fixture validation failed:");
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exitCode = 1;
} else {
  console.log(
    `Early lifecycle fixture validation passed: ${fixtures.operationContracts.length} operations, ${fixtures.cases.length} profile/provider/executor/policy cases`,
  );
}
