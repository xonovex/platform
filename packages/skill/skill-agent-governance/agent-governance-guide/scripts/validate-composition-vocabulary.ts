import {readFileSync} from "node:fs";
import {expectedVocabulary} from "./conformance-helpers.mjs";

// Every closed governance vocabulary is owned once by expectedVocabulary and
// mirrored in prose tables, bullet lists, and cross-package composition docs.
// This guard extracts each machine-readable declaring site and fails the build
// when any site drifts from the owner, generalizing validate-executor-vocabulary
// and validate-event-intent-vocabulary to the composition catalog. Human display
// tables that name the same concepts in prose form (modules.md module-kinds
// table, architecture.md authority-zone table, adoption-map.md executor-class and
// intent prose) carry no machine tokens and are validated elsewhere or left as
// views; registering them here would couple the guard to display wording.

const repoRootUrl = new URL("../../../../../", import.meta.url);

const governance =
  "packages/skill/skill-agent-governance/agent-governance-guide";
const commandWorkflow = "packages/command/command-workflow";

const readSource = (path: string): string =>
  readFileSync(new URL(path, repoRootUrl), "utf8");

// A backtick-quoted inline list on a single labeled line, e.g.
//   Supported outcomes are `allow`, `deny`, ... and `emergency-exception`.
const extractBacktickLine = (
  source: string,
  anchor: RegExp,
): string[] | null => {
  const line = anchor.exec(source);
  if (line === null) return null;
  const tokens = [...line[1].matchAll(/`([^`]+)`/g)].map((match) => match[1]);
  return tokens.length > 0 ? tokens : null;
};

// Backtick tokens inside the first parenthetical that follows a label, e.g.
//   adoption mode (`workflow-only`, `governance-only`, ...).
const extractParenthetical = (
  source: string,
  label: string,
): string[] | null => {
  const paren = new RegExp(`${label}\\s*\\(([^)]*)\\)`).exec(source);
  if (paren === null) return null;
  const tokens = [...paren[1].matchAll(/`([^`]+)`/g)].map((match) => match[1]);
  return tokens.length > 0 ? tokens : null;
};

// Leading backtick token of every "- `token` — gloss" bullet within a section.
const extractBulletTokens = (
  source: string,
  heading: string,
): string[] | null => {
  const section = new RegExp(`## ${heading}\\n([\\s\\S]*?)(?=\\n## |$)`).exec(
    source,
  );
  if (section === null) return null;
  const tokens = [...section[1].matchAll(/^-\s+`([^`]+)`\s+—/gm)].map(
    (match) => match[1],
  );
  return tokens.length > 0 ? tokens : null;
};

// Pipe-delimited values following a "prefix: " label, e.g.
//   classification: knowledge-only | advisory | ... | privileged.
const extractPipeList = (source: string, prefix: string): string[] | null => {
  const line = new RegExp(`^${prefix}:\\s*(.+)$`, "m").exec(source);
  if (line === null) return null;
  const tokens = line[1]
    .split("|")
    .map((token) => token.trim())
    .filter((token) => token.length > 0);
  return tokens.length > 0 ? tokens : null;
};

// First column of the first pipe table under a heading, dropping the header row
// and the |---| separator, for tables whose first cell is plain (not backticked).
const extractTableFirstColumn = (
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
  const cells = lines
    .slice(start, end)
    .map((line) => line.split("|")[1]?.trim() ?? "")
    .filter((cell) => cell.length > 0 && !/^[-:\s]+$/.test(cell));
  const data = cells.slice(1);
  return data.length > 0 ? data : null;
};

// Comma/"and"-separated tokens captured between two prose anchors.
const extractProseSeries = (
  source: string,
  anchor: RegExp,
): string[] | null => {
  const match = anchor.exec(source);
  if (match === null) return null;
  const tokens = match[1]
    .replace(/,?\s+and\s+/g, ", ")
    .split(/,\s*/)
    .map((token) => token.trim())
    .filter((token) => token.length > 0);
  return tokens.length > 0 ? tokens : null;
};

// Middle-dot list inside the first backtick span following a label, e.g.
//   `lifecycle · governance · ... · distribution`.
const extractMiddotSpan = (source: string, anchor: RegExp): string[] | null => {
  const span = anchor.exec(source);
  if (span === null) return null;
  const tokens = span[1]
    .split(/\s*·\s*/)
    .map((token) => token.trim())
    .filter((token) => token.length > 0);
  return tokens.length > 0 ? tokens : null;
};

const absent = (
  members: readonly string[],
  other: readonly string[],
): string[] => members.filter((member) => !other.includes(member));

type Site = {site: string; values: readonly string[] | null};

const compareSite = (
  {site, values}: Site,
  expected: readonly string[],
  normalize: (value: string) => string,
): string[] => {
  if (values === null) {
    return [`${site}: vocabulary declaration not found`];
  }
  const actual = values.map(normalize);
  const canonical = expected.map(normalize);
  const missing = absent(canonical, actual);
  const unexpected = absent(actual, canonical);
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
  normalize: (value: string) => string;
  sites: Site[];
};

const identity = (value: string): string => value;
const lower = (value: string): string => value.toLowerCase();

const policyAuthority = readSource(
  `${governance}/references/policy-and-authority.md`,
);
const modules = readSource(`${governance}/references/modules.md`);
const catalog = readSource(`${governance}/references/catalog-and-inventory.md`);
const architecture = readSource(`${governance}/references/architecture.md`);
const architectureComposition = readSource(
  `${commandWorkflow}/docs/architecture-and-composition.md`,
);
const adoptionMap = readSource(`${commandWorkflow}/docs/adoption-map.md`);

const checks: Check[] = [
  {
    label: "policyOutcomes",
    expected: expectedVocabulary.policyOutcomes,
    normalize: identity,
    sites: [
      {
        site: `${governance}/references/policy-and-authority.md (Supported outcomes)`,
        values: extractBacktickLine(
          policyAuthority,
          /^Supported outcomes are (.*)$/m,
        ),
      },
    ],
  },
  {
    label: "moduleClassifications",
    expected: expectedVocabulary.moduleClassifications,
    normalize: identity,
    sites: [
      {
        site: `${governance}/references/catalog-and-inventory.md (Classify executable effect)`,
        values: extractBulletTokens(
          catalog,
          "Show adoption mode separately from authority",
        ),
      },
      {
        site: `${governance}/references/modules.md (Declaration classification)`,
        values: extractPipeList(modules, "classification"),
      },
    ],
  },
  {
    label: "adoptionModes",
    expected: expectedVocabulary.adoptionModes,
    normalize: lower,
    sites: [
      {
        site: `${governance}/references/catalog-and-inventory.md (adoption mode list)`,
        values: extractParenthetical(catalog, "adoption mode"),
      },
      {
        site: `${governance}/references/architecture.md (Adoption modes prose)`,
        values: extractProseSeries(
          architecture,
          /## Adoption modes\s+([\s\S]*?)\s+compositions are independently valid/,
        ),
      },
      {
        site: `${commandWorkflow}/docs/architecture-and-composition.md (Adoption modes table)`,
        values: extractTableFirstColumn(
          architectureComposition,
          "Adoption modes",
        ),
      },
      {
        site: `${commandWorkflow}/docs/adoption-map.md (Modes table)`,
        values: extractTableFirstColumn(
          adoptionMap,
          "Modes are named compositions",
        ),
      },
    ],
  },
  {
    label: "authorityZones",
    expected: expectedVocabulary.authorityZones,
    normalize: identity,
    sites: [
      {
        site: `${governance}/references/catalog-and-inventory.md (authority zone list)`,
        values: extractParenthetical(catalog, "authority zone"),
      },
    ],
  },
  {
    label: "profileFacets",
    expected: expectedVocabulary.profileFacets,
    normalize: identity,
    sites: [
      {
        site: `${commandWorkflow}/docs/adoption-map.md (Governance profile facets)`,
        values: extractMiddotSpan(
          adoptionMap,
          /Governance profile facets[\s\S]*?`([^`]+)`/,
        ),
      },
    ],
  },
];

// Each mutation guard replays a fork this validator exists to catch — an invented
// value, a renamed value, a dropped value, and an unparseable site — derived from
// the live owner. A guard that reports no failure means the comparison stopped
// working, so the dud-guard count must be zero.
const mutationGuards = ({expected, normalize}: Check): string[][] => {
  const owner = [...expected];
  return [
    compareSite(
      {site: "guard-invented", values: [...owner, "__invented__"]},
      expected,
      normalize,
    ),
    compareSite(
      {
        site: "guard-renamed",
        values: owner.map((value, index) =>
          index === 0 ? `${value}__renamed` : value,
        ),
      },
      expected,
      normalize,
    ),
    compareSite(
      {site: "guard-dropped", values: owner.slice(1)},
      expected,
      normalize,
    ),
    compareSite({site: "guard-unparseable", values: null}, expected, normalize),
  ];
};

const failures = checks.flatMap((check) =>
  check.sites.flatMap((site) =>
    compareSite(site, check.expected, check.normalize).map(
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
    `Composition vocabulary mutation guards failed: ${dudGuards.length} tampered vocabularies passed validation`,
  );
  process.exitCode = 1;
} else if (failures.length > 0) {
  console.error(
    "Composition vocabulary drift detected across declaring sites:",
  );
  for (const failure of failures) {
    console.error(`- ${failure}`);
  }
  process.exitCode = 1;
} else {
  console.log(
    `Composition vocabulary validation passed: ${checks.length} vocabularies across ${siteCount} declaring sites; ${guards.length} mutation guards`,
  );
}
