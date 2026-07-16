import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {validateWalkingSkeleton} from "./walking-skeleton-helpers.mjs";

const readJson = (relativePath) =>
  JSON.parse(
    readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), "utf8"),
  );

const fixtures = {
  walking: readJson("../assets/walking-skeleton-fixtures.json"),
  governance: readJson("../assets/conformance-fixtures.json"),
  harness: readJson("../assets/harness-conformance-fixtures.json"),
  templates: readJson("../assets/harness-module-templates.json"),
  external: readJson("../assets/external-enforcement-fixtures.json"),
  enterprise: readJson("../assets/enterprise-platform-fixtures.json"),
};

const replaceScenario = (current, update) => ({
  ...current,
  walking: {
    ...current.walking,
    scenario: update(current.walking.scenario),
  },
});

const mutationCases = [
  {
    id: "apply-before-authorization",
    expectedCode: "onboarding:lifecycle-incomplete",
    apply: (current) =>
      replaceScenario(current, (scenario) => ({
        ...scenario,
        lifecycleEvents: [
          "discover",
          "assess",
          "recommend",
          "preview",
          "apply",
          "authorize",
          "verify",
          "record",
          "operate",
        ],
      })),
  },
  {
    id: "disabled-external-layer",
    expectedCode: "external:independent-enforcement-failed",
    apply: (current) =>
      replaceScenario(current, (scenario) => ({
        ...scenario,
        externalEnforcement: {
          ...scenario.externalEnforcement,
          targetMutation: "allowed",
        },
      })),
  },
  {
    id: "fresh-context-uses-conversation",
    expectedCode: "evidence:fresh-context-recovery-failed",
    apply: (current) =>
      replaceScenario(current, (scenario) => ({
        ...scenario,
        evidence: {
          ...scenario.evidence,
          freshContextRecovery: {
            ...scenario.evidence.freshContextRecovery,
            originalConversationAvailable: true,
          },
        },
      })),
  },
  {
    id: "duplicate-event-repeats-side-effect",
    expectedCode: "negative:unsafe-outcome:concurrent-duplicate-event",
    apply: (current) => ({
      ...current,
      walking: {
        ...current.walking,
        negativeCases: current.walking.negativeCases.map((testCase) =>
          testCase.id === "concurrent-duplicate-event"
            ? {...testCase, sideEffectCount: 2}
            : testCase,
        ),
      },
    }),
  },
];

const mutationFailures = mutationCases.flatMap((testCase) =>
  validateWalkingSkeleton(testCase.apply(fixtures)).includes(
    testCase.expectedCode,
  )
    ? []
    : [`mutation:${testCase.id}:expected ${testCase.expectedCode}`],
);
const failures = [...validateWalkingSkeleton(fixtures), ...mutationFailures];
if (failures.length > 0) {
  throw new Error(`walking skeleton fixture failures:\n${failures.join("\n")}`);
}

const moduleCount = fixtures.walking.scenario.recommendation.modules.length;
console.log(
  `walking skeleton fixtures valid: ${fixtures.walking.scenario.adoptionPaths.length} adoption paths, ${moduleCount} modules, ${fixtures.walking.negativeCases.length} negative cases, ${mutationCases.length} mutation guards`,
);
