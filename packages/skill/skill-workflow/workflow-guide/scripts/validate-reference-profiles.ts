import {readdirSync, readFileSync} from "node:fs";
import {join} from "node:path";
import {fileURLToPath} from "node:url";
import {resolveCapability} from "../../../skill-agent-governance/agent-governance-guide/scripts/capability-registry-helpers.mjs";
import {validateProfile as validateGovernanceProfile} from "../../../skill-agent-governance/agent-governance-guide/scripts/conformance-helpers.mjs";
import {validateProfile as validateWorkflowProfile} from "./conformance-helpers.mjs";

// Validate the shipped reference profile library: every profile satisfies its
// plane's profile contract (validateProfile), declares the required contract
// fields, and references only capabilities the capability registry classifies as
// shipped or adopter-supplied — never a dangling reference. An integrated workflow
// profile must name a governance profile that exists in the governance library.
// One profile ships per adoption mode with every team shape represented.

const workflowProfilesUrl = new URL("../assets/profiles/", import.meta.url);
const governanceProfilesUrl = new URL(
  "../../../skill-agent-governance/agent-governance-guide/assets/profiles/",
  import.meta.url,
);

interface LoadedProfile {
  file: string;
  profile: any;
}

const loadProfiles = (dirUrl: URL): LoadedProfile[] => {
  const dir = fileURLToPath(dirUrl);
  return readdirSync(dir)
    .filter((entry) => entry.endsWith(".json"))
    .toSorted()
    .map((file) => ({
      file,
      profile: JSON.parse(readFileSync(join(dir, file), "utf8")),
    }));
};

const requiredCommon = [
  "identity",
  "version",
  "owner",
  "scope",
  "applicability",
  "adoptionMode",
  "teamShape",
  "capabilities",
  "failureBehavior",
];
const requiredWorkflow = [
  ...requiredCommon,
  "includedResults",
  "preservedResults",
  "mandatoryControls",
];
const requiredGovernance = [
  ...requiredCommon,
  "facets",
  "strengthening",
  "adequateEnforcement",
];

const missingFields = (profile: any, required: readonly string[]): string[] =>
  required.filter((field) => profile[field] === undefined);

// Capability references the registry cannot classify (unknown category or name),
// which are dangling references a profile must not carry.
const danglingCapabilities = (profile: any): string[] =>
  Object.entries(profile.capabilities ?? {}).flatMap(([category, names]) =>
    ((names as string[] | undefined) ?? [])
      .filter((name) => resolveCapability(category, name) === "dangling")
      .map((name) => `${category}/${name}`),
  );

const workflowProfiles = loadProfiles(workflowProfilesUrl);
const governanceProfiles = loadProfiles(governanceProfilesUrl);
const governanceIdentities = new Set(
  governanceProfiles.map(({profile}) => profile.identity),
);

const failures: string[] = [];

for (const {file, profile} of workflowProfiles) {
  const missing = missingFields(profile, requiredWorkflow);
  if (missing.length > 0) {
    failures.push(`workflow/${file}: missing ${missing.join(", ")}`);
    continue;
  }
  const code = validateWorkflowProfile(profile);
  if (code !== null) {
    failures.push(`workflow/${file}: profile invalid (${code})`);
  }
  for (const capability of danglingCapabilities(profile)) {
    failures.push(`workflow/${file}: unresolvable capability ${capability}`);
  }
  const governanceRef = profile.governance?.profile;
  if (
    governanceRef !== undefined &&
    governanceRef !== null &&
    !governanceIdentities.has(governanceRef)
  ) {
    failures.push(
      `workflow/${file}: governance facet names unknown governance profile ${governanceRef}`,
    );
  }
}

for (const {file, profile} of governanceProfiles) {
  const missing = missingFields(profile, requiredGovernance);
  if (missing.length > 0) {
    failures.push(`governance/${file}: missing ${missing.join(", ")}`);
    continue;
  }
  const code = validateGovernanceProfile(profile);
  if (code !== null) {
    failures.push(`governance/${file}: profile invalid (${code})`);
  }
  for (const capability of danglingCapabilities(profile)) {
    failures.push(`governance/${file}: unresolvable capability ${capability}`);
  }
}

// Coverage: one reference profile per adoption mode, every team shape represented.
const allProfiles = [...workflowProfiles, ...governanceProfiles].map(
  ({profile}) => profile,
);
const modes = new Set(allProfiles.map((profile) => profile.adoptionMode));
const shapes = new Set(allProfiles.map((profile) => profile.teamShape));
for (const mode of [
  "workflow-only",
  "governance-only",
  "enablement-only",
  "external-enforcement-only",
  "integrated",
]) {
  if (!modes.has(mode)) {
    failures.push(`coverage: no reference profile for adoption mode ${mode}`);
  }
}
for (const shape of ["solo", "small-team", "regulated"]) {
  if (!shapes.has(shape)) {
    failures.push(`coverage: no reference profile for team shape ${shape}`);
  }
}

// Mutation guards — the field and capability checks must catch a tampered profile;
// a dud (empty result) means a check silently broke.
const guards: string[][] = [
  missingFields({identity: "x"}, requiredWorkflow),
  danglingCapabilities({capabilities: {method: ["__invented__"]}}),
  danglingCapabilities({capabilities: {"__unknown-category__": ["x"]}}),
];
const dudGuards = guards.filter((guardFailures) => guardFailures.length === 0);

if (dudGuards.length > 0) {
  console.error(
    `Reference profile mutation guards failed: ${dudGuards.length} tampered checks passed validation`,
  );
  process.exitCode = 1;
} else if (failures.length > 0) {
  console.error("Reference profile validation failed:");
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exitCode = 1;
} else {
  console.log(
    `Reference profile validation passed: ${workflowProfiles.length} workflow + ${governanceProfiles.length} governance profiles across ${modes.size} adoption modes and ${shapes.size} team shapes; ${guards.length} mutation guards`,
  );
}
