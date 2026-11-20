import {isRecord} from "@xonovex/script-moon-common/records";
import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import {parseCatalogJson, type CatalogFile} from "./catalog-json.js";

const QUERY_SPLITS = new Set(["train", "validation"]);

// The query shapes an early generator produced for every skill; they say
// nothing about one skill, so a catalog query must not match them.
const GENERIC_TRIGGER_QUERY_RES = [
  /^Help me handle .+ correctly in this .+ task, including the edge cases that usually get missed\.$/i,
  /^I'm reviewing `[^`]+` in an? .+ project\. The happy path works, but .+ is unclear\./i,
  /^A teammate says our .+ change may mishandle .+ before release\. Inspect the likely risk around /i,
  /^CI started failing after we changed .+ in an? .+ project\. Use /i,
  /^A review comment on `work\//i,
  /^quick pre-merge sanity check: after the /i,
  /^The clean Linux CI job fails only for the minimal /i,
];

interface TriggerEntry {
  readonly generatedNegativeOwner: string | undefined;
  readonly shouldTrigger: boolean;
  readonly split: string | undefined;
}

const isGenericTriggerQuery = (query: string): boolean =>
  GENERIC_TRIGGER_QUERY_RES.some((pattern) => pattern.test(query.trim()));

const checkQuery = (
  query: unknown,
  index: number,
  seen: Set<string>,
): readonly Finding[] => {
  if (typeof query !== "string" || query.trim().length === 0) {
    return [
      fail(
        `catalog: eval-queries.json entry ${String(index + 1)} needs a non-empty query`,
      ),
    ];
  }
  const key = query.trim().toLowerCase();
  const findings = [
    ...(seen.has(key)
      ? [fail(`catalog: duplicate trigger query '${query}'`)]
      : []),
    ...(isGenericTriggerQuery(query)
      ? [
          fail(
            `catalog: generic trigger eval query must be replaced: '${query}'`,
          ),
        ]
      : []),
  ];
  seen.add(key);
  return findings;
};

const checkEntry = (
  value: unknown,
  index: number,
  seen: Set<string>,
): {readonly findings: readonly Finding[]; readonly entry?: TriggerEntry} => {
  const position = String(index + 1);
  if (!isRecord(value)) {
    return {
      findings: [
        fail(`catalog: eval-queries.json entry ${position} must be an object`),
      ],
    };
  }
  const queryFindings = checkQuery(value.query, index, seen);
  if (typeof value.should_trigger !== "boolean") {
    return {
      findings: [
        ...queryFindings,
        fail(
          `catalog: eval-queries.json entry ${position} needs boolean should_trigger`,
        ),
      ],
    };
  }
  const rationale = value.rationale;
  const generatedNegativeOwner =
    typeof rationale === "string" && !value.should_trigger
      ? /^near miss owned by (\S+)$/.exec(rationale.trim())?.[1]
      : undefined;
  const split =
    typeof value.split === "string" && QUERY_SPLITS.has(value.split)
      ? value.split
      : undefined;
  return {
    findings: [
      ...queryFindings,
      ...(split === undefined
        ? [
            fail(
              `catalog: eval-queries.json entry ${position} needs split train or validation`,
            ),
          ]
        : []),
    ],
    entry: {generatedNegativeOwner, shouldTrigger: value.should_trigger, split},
  };
};

const checkMinimums = (
  entries: readonly TriggerEntry[],
): readonly Finding[] => {
  const positive = entries.filter(({shouldTrigger}) => shouldTrigger).length;
  const negative = entries.length - positive;
  const splitCount = (polarity: string, split: string): number =>
    entries.filter(
      (entry) =>
        entry.split === split &&
        (entry.shouldTrigger ? "positive" : "negative") === polarity,
    ).length;
  return [
    ...(positive < 8 || negative < 8
      ? [
          fail(
            `catalog: trigger evals need at least 8 positive and 8 negative queries (found ${String(positive)}/${String(negative)})`,
          ),
        ]
      : []),
    ...["positive", "negative"].flatMap((polarity) =>
      [...QUERY_SPLITS].flatMap((split) =>
        splitCount(polarity, split) < 2
          ? [
              fail(
                `catalog: trigger evals need at least 2 ${polarity} ${split} queries`,
              ),
            ]
          : [],
      ),
    ),
  ];
};

// A generated near miss names the sibling that owns the query; drawing them
// from a single sibling would test one boundary only.
const checkGeneratedOwners = (
  entries: readonly TriggerEntry[],
): readonly Finding[] => {
  const owners = entries.flatMap(({generatedNegativeOwner}) =>
    generatedNegativeOwner === undefined ? [] : [generatedNegativeOwner],
  );
  const required = Math.min(3, owners.length);
  const distinct = new Set(owners).size;
  return distinct < required
    ? [
        fail(
          `catalog: generated negative routes need at least ${String(required)} sibling owners (found ${String(distinct)})`,
        ),
      ]
    : [];
};

/** The routing queries in eval-queries.json. */
export const checkTriggerEvals = (
  file: CatalogFile | undefined,
): readonly Finding[] => {
  if (file === undefined) return [fail("catalog: missing eval-queries.json")];
  const parsed = parseCatalogJson("eval-queries.json", file);
  if (parsed.kind === "invalid") return [parsed.finding];
  const raw = parsed.value;
  if (!Array.isArray(raw)) {
    return [fail("catalog: eval-queries.json must be an array")];
  }
  const seen = new Set<string>();
  const checked = raw.map((value, index) => checkEntry(value, index, seen));
  const entries = checked.flatMap(({entry}) =>
    entry === undefined ? [] : [entry],
  );
  const findings = [
    ...checked.flatMap(({findings: entryFindings}) => entryFindings),
    ...checkMinimums(entries),
    ...checkGeneratedOwners(entries),
  ];
  const positive = entries.filter(({shouldTrigger}) => shouldTrigger).length;
  return findings.some(({message}) => message.includes("trigger eval"))
    ? findings
    : [
        ...findings,
        pass(
          `catalog: trigger evals cover ${String(positive)} positive and ${String(entries.length - positive)} negative routes with train/validation splits`,
        ),
      ];
};
