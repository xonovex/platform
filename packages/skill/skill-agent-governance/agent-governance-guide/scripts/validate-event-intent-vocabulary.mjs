import {readFileSync} from "node:fs";

const repoRootUrl = new URL("../../../../../", import.meta.url);

const ownerPath =
  "packages/skill/skill-agent-governance/agent-governance-guide/references/events-and-capabilities.md";

const adapterViewPath =
  "packages/command/command-workflow/docs/harness-capabilities.md";

const diagramViewPath =
  "packages/diagram/diagram-agent-workflow/workflow-diagram.dot";

const readSource = (path) => readFileSync(new URL(path, repoRootUrl), "utf8");

// The owner declares the families; the views are compared against whatever it
// declares, so a family added there fails every view until the view follows.
const extractOwnerFamilies = (source) => {
  const section = /## Event intent taxonomy\n([\s\S]*?)(?=\n## |$)/.exec(
    source,
  );
  const families = section
    ? [...section[1].matchAll(/^- \*\*([^*]+)\*\* —/gm)].map(
        (match) => match[1],
      )
    : [];

  return families.length > 0 ? families : null;
};

const extractProseFamilies = (source) => {
  const line = /^Families: (.*)$/m.exec(source);
  const families = line
    ? [...line[1].matchAll(/`([^`]+)`/g)].map((match) => match[1])
    : [];

  return families.length > 0 ? families : null;
};

// The first label line titles the node; every later line is a `|`-separated
// row of family names.
const extractDiagramFamilies = (source) => {
  const node = /^\s*intents \[label="([^"]*)"/m.exec(source);
  const [, ...familyLines] = node ? node[1].split("\\n") : [null];
  const families = familyLines
    .flatMap((line) => line.split("|"))
    .map((part) => part.trim())
    .filter((part) => part.length > 0);

  return families.length > 0 ? families : null;
};

const absent = (members, other) =>
  members.filter((member) => !other.includes(member));

const compareVocabulary = ({site, families}, expected) => {
  if (families === null) {
    return [`${site}: intent family declaration not found`];
  }

  const missing = absent(expected, families);
  const unexpected = absent(families, expected);
  const failures = [];

  if (missing.length > 0) {
    failures.push(`${site}: missing intent families ${missing.join(", ")}`);
  }

  if (unexpected.length > 0) {
    failures.push(
      `${site}: unexpected intent families ${unexpected.join(", ")}`,
    );
  }

  return failures;
};

const compareSites = (sites, expected) =>
  sites.flatMap((site) => compareVocabulary(site, expected));

const ownerFamilies = extractOwnerFamilies(readSource(ownerPath));

if (ownerFamilies === null) {
  console.error(
    `Event intent vocabulary owner not readable: ${ownerPath} declares no bold family under "## Event intent taxonomy"`,
  );
  process.exitCode = 1;
} else {
  const viewSites = [
    {
      site: `${adapterViewPath} (Families list)`,
      families: extractProseFamilies(readSource(adapterViewPath)),
    },
    {
      site: `${diagramViewPath} (intents node label)`,
      families: extractDiagramFamilies(readSource(diagramViewPath)),
    },
  ];

  const failures = compareSites(viewSites, ownerFamilies);

  // Each mutation guard replays a fork this validator exists to catch: an
  // invented family, a renamed family, a dropped family, and an unparseable
  // view. A guard that reports no failure means the comparison stopped working.
  const guards = [
    compareSites(
      [{site: "guard", families: [...ownerFamilies, "file"]}],
      ownerFamilies,
    ),
    compareSites(
      [
        {
          site: "guard",
          families: ownerFamilies.map((family) =>
            family === "subagent" ? "child agent" : family,
          ),
        },
      ],
      ownerFamilies,
    ),
    compareSites(
      [
        {
          site: "guard",
          families: ownerFamilies.filter((family) => family !== "permission"),
        },
      ],
      ownerFamilies,
    ),
    compareSites(
      [{site: "guard", families: [...ownerFamilies, "completion"]}],
      ownerFamilies,
    ),
    compareSites([{site: "guard", families: null}], ownerFamilies),
  ];

  const dudGuards = guards.filter(
    (guardFailures) => guardFailures.length === 0,
  );

  if (dudGuards.length > 0) {
    console.error(
      `Event intent vocabulary mutation guards failed: ${dudGuards.length} tampered vocabularies passed validation`,
    );
    process.exitCode = 1;
  } else if (failures.length > 0) {
    console.error(
      `Event intent vocabulary drift detected; ${ownerPath} owns the families ${ownerFamilies.join(", ")}:`,
    );
    for (const failure of failures) {
      console.error(`- ${failure}`);
    }
    process.exitCode = 1;
  } else {
    console.log(
      `Event intent vocabulary validation passed: ${viewSites.length} views derive from ${ownerFamilies.length} owned families (${ownerFamilies.join(", ")}); ${guards.length} mutation guards`,
    );
  }
}
