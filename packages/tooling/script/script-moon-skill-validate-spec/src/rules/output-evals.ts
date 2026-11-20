import {isRecord} from "@xonovex/script-moon-common/records";
import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import {parseCatalogJson, type CatalogFile} from "./catalog-json.js";

const OUTPUT_TIERS = new Set(["aggressive", "moderate", "conservative"]);

const checkEntry = (entry: unknown, index: number): readonly Finding[] => {
  const position = String(index + 1);
  if (!isRecord(entry)) {
    return [fail(`catalog: evals.json entry ${position} must be an object`)];
  }
  const hasPrompt =
    typeof entry.prompt === "string" && entry.prompt.trim().length > 0;
  const hasAssertions =
    Array.isArray(entry.assertions) &&
    entry.assertions.length > 0 &&
    entry.assertions.every(
      (assertion) =>
        typeof assertion === "string" && assertion.trim().length > 0,
    );
  return [
    ...(hasPrompt
      ? []
      : [
          fail(
            `catalog: evals.json entry ${position} needs a non-empty prompt`,
          ),
        ]),
    ...(hasAssertions
      ? []
      : [
          fail(
            `catalog: evals.json entry ${position} needs non-empty string assertions`,
          ),
        ]),
  ];
};

/** The output evals in evals.json, which name the skill they grade. */
export const checkOutputEvals = (
  file: CatalogFile | undefined,
  skillName: unknown,
): readonly Finding[] => {
  if (file === undefined) return [fail("catalog: missing evals.json")];
  const parsed = parseCatalogJson("evals.json", file);
  if (parsed.kind === "invalid") return [parsed.finding];
  const raw = parsed.value;
  if (!isRecord(raw)) return [fail("catalog: evals.json must be an object")];
  const header = [
    ...(raw.skill_name === skillName
      ? []
      : [
          fail(
            `catalog: evals.json skill_name '${String(raw.skill_name)}' does not match '${String(skillName)}'`,
          ),
        ]),
    ...(typeof raw.tier === "string" && OUTPUT_TIERS.has(raw.tier)
      ? []
      : [
          fail(
            "catalog: evals.json tier must be aggressive, moderate, or conservative",
          ),
        ]),
  ];
  if (!Array.isArray(raw.evals) || raw.evals.length < 3) {
    return [
      ...header,
      fail("catalog: evals.json must contain at least 3 output evals"),
    ];
  }
  return [
    ...header,
    ...raw.evals.flatMap((entry, index) => checkEntry(entry, index)),
    pass(
      `catalog: ${String(raw.evals.length)} output eval(s) are structurally valid`,
    ),
  ];
};
