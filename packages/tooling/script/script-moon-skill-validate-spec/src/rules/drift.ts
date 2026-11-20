import type {DriftLintMode} from "@xonovex/script-moon-common/drift-budgets";
import type {DriftLintReport} from "@xonovex/script-moon-skill-catalog-common/drift-lints";
import {
  drift,
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";

// Drift findings stay advisory in warn mode: --strict escalates authoring
// warnings, but not measurements the catalog has not yet been corrected for.
export const checkDrift = (
  report: DriftLintReport,
  mode: DriftLintMode,
): readonly Finding[] => [
  ...report.manifestErrors.map((error) => fail(error)),
  ...report.findings.map((finding) =>
    mode === "enforce" ? fail(finding) : drift(finding),
  ),
  ...(report.findings.length === 0 && report.manifestErrors.length === 0
    ? [pass("drift: file budgets and vocabulary ownership hold")]
    : []),
];
