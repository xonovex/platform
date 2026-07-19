import {readFileSync} from "node:fs";
import {expectedVocabulary} from "./conformance-helpers.mjs";

const repoRootUrl = new URL("../../../../../", import.meta.url);

const canonicalExecutorClasses = [
  "deterministic",
  "model",
  "agent",
  "human",
  "external",
];

// selectDevelopmentExecutor resolves work shape to an executor, so it declares
// only the three shape-selectable classes; human and external are selected on
// authority or ownership and have no work-shape path.
const workShapeSelections = {
  mechanical: "deterministic",
  "bounded-transform": "model",
  adaptive: "agent",
};

const permissionRejectionCodes = new Set([
  "executor-not-permitted",
  "least-adaptive-executor-bypassed",
]);

const workflowHelpersPath =
  "packages/skill/skill-workflow/workflow-guide/scripts/development-assurance-helpers.ts";

const readSource = (path) => readFileSync(new URL(path, repoRootUrl), "utf8");

const readJson = (path) => JSON.parse(readSource(path));

const extractSetMembers = (source, identifier) => {
  const declaration = new RegExp(
    `(?:const|let|var)\\s+${identifier}\\s*=\\s*new Set\\(\\[([\\s\\S]*?)\\]\\s*\\)`,
  ).exec(source);

  return declaration
    ? [...declaration[1].matchAll(/"([^"]+)"|'([^']+)'/g)].map(
        (match) => match[1] ?? match[2],
      )
    : null;
};

const extractTableClasses = (source, heading) => {
  const section = new RegExp(`## ${heading}\\n([\\s\\S]*?)(?=\\n## |$)`).exec(
    source,
  );
  const classes = section
    ? [...section[1].matchAll(/^\|\s*`([^`]+)`\s*\|/gm)].map(
        (match) => match[1],
      )
    : [];

  return classes.length > 0 ? classes : null;
};

const absent = (members, other) =>
  members.filter((member) => !other.includes(member));

const compareVocabulary = ({site, executors}, expected) => {
  if (executors === null) {
    return [`${site}: executor vocabulary declaration not found`];
  }

  const missing = absent(expected, executors);
  const unexpected = absent(executors, expected);
  const failures = [];

  if (missing.length > 0) {
    failures.push(`${site}: missing executor classes ${missing.join(", ")}`);
  }

  if (unexpected.length > 0) {
    failures.push(
      `${site}: unexpected executor classes ${unexpected.join(", ")}`,
    );
  }

  return failures;
};

const compareSites = (sites, expected) =>
  sites.flatMap((site) => compareVocabulary(site, expected));

const selectableExecutors = (select, workShape) =>
  canonicalExecutorClasses.filter(
    (executor) =>
      !permissionRejectionCodes.has(
        select({workShape, requestedExecutor: executor, bounds: {}}).code,
      ),
  );

const compareWorkShapeSelections = (select) =>
  Object.entries(workShapeSelections).flatMap(([workShape, expected]) => {
    const selectable = selectableExecutors(select, workShape);

    return selectable.length === 1 && selectable[0] === expected
      ? []
      : [
          `${workflowHelpersPath}: work shape ${workShape} selects ${
            selectable.length > 0 ? selectable.join(", ") : "no canonical class"
          } instead of ${expected}`,
        ];
  });

const {selectDevelopmentExecutor} = await import(
  new URL(workflowHelpersPath, repoRootUrl)
);

const fullVocabularySites = [
  {
    site: "packages/skill/skill-agent-governance/agent-governance-guide/references/execution.md (Executor classes table)",
    executors: extractTableClasses(
      readSource(
        "packages/skill/skill-agent-governance/agent-governance-guide/references/execution.md",
      ),
      "Executor classes",
    ),
  },
  {
    site: "packages/skill/skill-agent-governance/agent-governance-guide/scripts/conformance-helpers.mjs (expectedVocabulary.executorClasses)",
    executors: expectedVocabulary.executorClasses ?? null,
  },
  {
    site: "packages/skill/skill-agent-governance/agent-governance-guide/assets/conformance-fixtures.json (executorClasses)",
    executors:
      readJson(
        "packages/skill/skill-agent-governance/agent-governance-guide/assets/conformance-fixtures.json",
      ).executorClasses ?? null,
  },
  {
    site: "packages/skill/skill-plan/plan-guide/scripts/validate-early-lifecycle-fixtures.mjs (allowedExecutors)",
    executors: extractSetMembers(
      readSource(
        "packages/skill/skill-plan/plan-guide/scripts/validate-early-lifecycle-fixtures.mjs",
      ),
      "allowedExecutors",
    ),
  },
];

const failures = [
  ...compareSites(fullVocabularySites, canonicalExecutorClasses),
  ...compareWorkShapeSelections(selectDevelopmentExecutor),
];

const guards = [
  compareSites(
    [{site: "guard", executors: ["deterministic", "model", "agent", "human"]}],
    canonicalExecutorClasses,
  ),
  compareSites(
    [
      {
        site: "guard",
        executors: [
          "deterministic",
          "bounded-model",
          "adaptive-agent",
          "human",
          "qualified-human",
        ],
      },
    ],
    canonicalExecutorClasses,
  ),
  compareSites(
    [
      {
        site: "guard",
        executors: [...canonicalExecutorClasses, "qualified-human"],
      },
    ],
    canonicalExecutorClasses,
  ),
  compareSites([{site: "guard", executors: null}], canonicalExecutorClasses),
  compareWorkShapeSelections(({workShape, requestedExecutor}) =>
    workShape === "bounded-transform" && requestedExecutor === "model"
      ? {code: "executor-not-permitted"}
      : selectDevelopmentExecutor({workShape, requestedExecutor, bounds: {}}),
  ),
];

const dudGuards = guards.filter((guardFailures) => guardFailures.length === 0);

if (dudGuards.length > 0) {
  console.error(
    `Executor vocabulary mutation guards failed: ${dudGuards.length} tampered vocabularies passed validation`,
  );
  process.exitCode = 1;
} else if (failures.length > 0) {
  console.error(
    `Executor vocabulary drift detected; the canonical classes are ${canonicalExecutorClasses.join(", ")}:`,
  );
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exitCode = 1;
} else {
  console.log(
    `Executor vocabulary validation passed: ${fullVocabularySites.length} declaring sites and ${Object.keys(workShapeSelections).length} work-shape selections agree on ${canonicalExecutorClasses.join(", ")}; ${guards.length} mutation guards`,
  );
}
