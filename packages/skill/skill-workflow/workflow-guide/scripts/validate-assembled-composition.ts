import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {
  validateAssembledComposition,
  type AssembledComposition,
} from "./composition-helpers.ts";

// Prove a shipped reference profile assembles into a complete composition, and
// that each adversarial fixture is rejected with its own failure code, so a
// catalog cannot be assembled inconsistent or incomplete. Mutation guards replay
// a tamper on the passing composition; a dud (null) means a check silently broke.

const loadJson = (url: URL): any =>
  JSON.parse(readFileSync(fileURLToPath(url), "utf8"));

const mergeCapabilities = (
  ...sources: Record<string, string[]>[]
): Record<string, string[]> => {
  const merged: Record<string, string[]> = {};
  for (const source of sources) {
    for (const [category, names] of Object.entries(source ?? {})) {
      merged[category] = [...new Set([...(merged[category] ?? []), ...names])];
    }
  }
  return merged;
};

const providersFrom = (
  ...profiles: any[]
): {port: string; available: boolean; compatible: boolean}[] =>
  profiles.flatMap((profile) =>
    (profile.capabilities?.["provider-port"] ?? []).map((port: string) => ({
      port,
      available: true,
      compatible: true,
    })),
  );

const failures: string[] = [];

// 1. Assemble the shipped Phase 3 integrated reference profile with the governance
//    profile it pairs with; the whole composition must be complete.
const integrated = loadJson(
  new URL("../assets/profiles/integrated.json", import.meta.url),
);
const governanceOnly = loadJson(
  new URL(
    "../../../skill-agent-governance/agent-governance-guide/assets/profiles/governance-only.json",
    import.meta.url,
  ),
);
const capabilities = mergeCapabilities(
  integrated.capabilities,
  governanceOnly.capabilities,
);
const shippedComposition: AssembledComposition = {
  adoptionMode: integrated.adoptionMode,
  absenceReport: integrated.absenceReport,
  workflowProfile: integrated,
  governanceProfile: governanceOnly,
  capabilities,
  requiredCapabilities: capabilities,
  modules: [],
  conflictResolution: false,
  providers: providersFrom(integrated, governanceOnly),
  mandatoryControls: integrated.mandatoryControls ?? [],
};
const shippedCode = validateAssembledComposition(shippedComposition);
if (shippedCode !== null) {
  failures.push(
    `shipped integrated composition (integrated-profile + governance-only-profile) invalid: ${shippedCode}`,
  );
}

// 2. Each fixture case returns the expected failure code (or null when valid).
const fixture = loadJson(
  new URL("../assets/assembled-composition-fixtures.json", import.meta.url),
);
for (const testCase of fixture.cases) {
  const code = validateAssembledComposition(testCase.selection);
  const valid = code === null;
  const codeMatches = testCase.expectedValid || code === testCase.expectedCode;
  if (valid !== testCase.expectedValid || !codeMatches) {
    failures.push(
      `${testCase.id}: expected valid=${testCase.expectedValid} code=${testCase.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
    );
  }
}

// 3. Mutation guards — tamper the passing composition to trip each cross-plane
//    check; every guard must report a failure.
const base = fixture.cases.find(
  (testCase: any) => testCase.expectedValid,
).selection;
const guards: (string | null)[] = [
  validateAssembledComposition({...base, absenceReport: ""}),
  validateAssembledComposition({
    ...base,
    capabilities: {...base.capabilities, method: ["__dangling__"]},
  }),
  validateAssembledComposition({
    ...base,
    providers: [{port: "policy", available: true, compatible: false}],
  }),
  validateAssembledComposition({
    ...base,
    mandatoryControls: [
      {id: "x", enforcementPoints: [{supported: true, guaranteed: false}]},
    ],
  }),
];
const dudGuards = guards.filter((code) => code === null);

if (dudGuards.length > 0) {
  console.error(
    `Assembled composition mutation guards failed: ${dudGuards.length} tampered compositions passed validation`,
  );
  process.exitCode = 1;
} else if (failures.length > 0) {
  console.error("Assembled composition validation failed:");
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exitCode = 1;
} else {
  console.log(
    `Assembled composition validation passed: 1 shipped integrated composition + ${fixture.cases.length} fixtures; ${guards.length} mutation guards`,
  );
}
