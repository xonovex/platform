import {readFileSync} from "node:fs";
import {developmentWorkShapes} from "./development-assurance-helpers.mjs";
import {independenceLevels} from "./independence-helpers.mjs";

// The workflow plane owns two closed vocabularies that are also declared in prose
// and fixtures: the independence levels (owned by independence-helpers, mirrored
// in the agent-governance actors reference) and the development work shapes (owned
// by development-assurance-helpers, mirrored in the development-contracts table and
// the assurance fixtures). This guard gives the workflow plane the cross-site
// vocabulary check the governance plane already has, failing the build when a site
// drifts from its owner. Independence's default `none` level is declared only by
// the owner and the resolution prose, so the actors levels table is validated
// against the above-`none` levels it enumerates.

const repoRootUrl = new URL("../../../../../", import.meta.url);

const workflow = "packages/skill/skill-workflow/workflow-guide";
const governance =
  "packages/skill/skill-agent-governance/agent-governance-guide";

const readSource = (path: string): string =>
  readFileSync(new URL(path, repoRootUrl), "utf8");

const readJson = (path: string): any => JSON.parse(readSource(path));

// First-column backtick tokens of the first pipe table under a heading, e.g.
// `distinct-identity` or `mechanical`. Only the first contiguous table block is
// read so a later table in the same section (the independence resolution table
// naming `none`) does not leak rows into the levels vocabulary.
const extractTableTokens = (
  source: string,
  heading: string,
): string[] | null => {
  const section = new RegExp(`## ${heading}\\n([\\s\\S]*?)(?=\\n## |$)`).exec(
    source,
  );
  if (section === null) return null;
  const lines = section[1].split("\n");
  const start = lines.findIndex((line) => line.trimStart().startsWith("|"));
  if (start === -1) return null;
  let end = start;
  while (end < lines.length && lines[end].trimStart().startsWith("|")) end += 1;
  const block = lines.slice(start, end).join("\n");
  const tokens = [...block.matchAll(/^\|\s*`([^`]+)`\s*\|/gm)].map(
    (match) => match[1],
  );
  return tokens.length > 0 ? tokens : null;
};

// Unique input.workShape values across the executor-selection fixture cases.
const extractFixtureWorkShapes = (data: any): string[] | null => {
  const cases = data?.cases;
  if (!Array.isArray(cases)) return null;
  const shapes = [
    ...new Set(
      cases
        .map((testCase: any) => testCase?.input?.workShape)
        .filter((shape: unknown): shape is string => typeof shape === "string"),
    ),
  ];
  return shapes.length > 0 ? shapes : null;
};

const absent = (
  members: readonly string[],
  other: readonly string[],
): string[] => members.filter((member) => !other.includes(member));

type Site = {site: string; values: readonly string[] | null};

const compareSite = (
  {site, values}: Site,
  expected: readonly string[],
): string[] => {
  if (values === null) {
    return [`${site}: vocabulary declaration not found`];
  }
  const missing = absent(expected, values);
  const unexpected = absent(values, expected);
  const failures: string[] = [];
  if (missing.length > 0) {
    failures.push(`${site}: missing ${missing.join(", ")}`);
  }
  if (unexpected.length > 0) {
    failures.push(`${site}: unexpected ${unexpected.join(", ")}`);
  }
  return failures;
};

type Check = {
  label: string;
  expected: readonly string[];
  sites: Site[];
};

const actors = readSource(`${governance}/references/actors.md`);
const developmentContracts = readSource(
  `${workflow}/references/development-contracts.md`,
);
const developmentFixtures = readJson(
  `${workflow}/assets/development-assurance-fixtures.json`,
);

const aboveNoneLevels = independenceLevels.filter((level) => level !== "none");

const checks: Check[] = [
  {
    label: "independenceLevels (above none)",
    expected: aboveNoneLevels,
    sites: [
      {
        site: `${governance}/references/actors.md (Independence levels table)`,
        values: extractTableTokens(actors, "Independence"),
      },
    ],
  },
  {
    label: "developmentWorkShapes",
    expected: developmentWorkShapes,
    sites: [
      {
        site: `${workflow}/references/development-contracts.md (Work shape table)`,
        values: extractTableTokens(developmentContracts, "Executor selection"),
      },
      {
        site: `${workflow}/assets/development-assurance-fixtures.json (input.workShape)`,
        values: extractFixtureWorkShapes(developmentFixtures),
      },
    ],
  },
];

// Each mutation guard replays a fork this validator exists to catch — an invented
// value, a renamed value, a dropped value, and an unparseable site — derived from
// the live owner. A guard that reports no failure means the comparison stopped
// working, so the dud-guard count must be zero.
const mutationGuards = ({expected}: Check): string[][] => {
  const owner = [...expected];
  return [
    compareSite(
      {site: "guard-invented", values: [...owner, "__invented__"]},
      expected,
    ),
    compareSite(
      {
        site: "guard-renamed",
        values: owner.map((value, index) =>
          index === 0 ? `${value}__renamed` : value,
        ),
      },
      expected,
    ),
    compareSite({site: "guard-dropped", values: owner.slice(1)}, expected),
    compareSite({site: "guard-unparseable", values: null}, expected),
  ];
};

const failures = checks.flatMap((check) =>
  check.sites.flatMap((site) =>
    compareSite(site, check.expected).map(
      (failure) => `[${check.label}] ${failure}`,
    ),
  ),
);

const guards = checks.flatMap((check) => mutationGuards(check));
const dudGuards = guards.filter((guardFailures) => guardFailures.length === 0);
const siteCount = checks.reduce(
  (total, check) => total + check.sites.length,
  0,
);

if (dudGuards.length > 0) {
  console.error(
    `Workflow vocabulary mutation guards failed: ${dudGuards.length} tampered vocabularies passed validation`,
  );
  process.exitCode = 1;
} else if (failures.length > 0) {
  console.error("Workflow vocabulary drift detected across declaring sites:");
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exitCode = 1;
} else {
  console.log(
    `Workflow vocabulary validation passed: ${checks.length} vocabularies across ${siteCount} declaring sites; ${guards.length} mutation guards`,
  );
}
