import type {
  Finding,
  FindingLevel,
} from "@xonovex/script-moon-skill-catalog-common/finding";

export interface RenderedReport {
  readonly lines: readonly string[];
  readonly exitCode: number;
}

const messages = (
  findings: readonly Finding[],
  level: FindingLevel,
): readonly string[] =>
  findings
    .filter((finding) => finding.level === level)
    .map(({message}) => message);

const verdict = (
  failures: number,
  warnings: number,
  strict: boolean,
): RenderedReport => {
  if (failures > 0) {
    return {
      lines: [
        `Result: FAIL (${String(failures)} error(s), ${String(warnings)} warning(s))`,
      ],
      exitCode: 1,
    };
  }
  if (warnings === 0) {
    return {lines: ["Result: PASS (no warnings)"], exitCode: 0};
  }
  return strict
    ? {
        lines: [`Result: FAIL (${String(warnings)} warning(s) in strict mode)`],
        exitCode: 1,
      }
    : {
        lines: [`Result: PASS with ${String(warnings)} warning(s)`],
        exitCode: 0,
      };
};

// The report groups the findings by level; drift never decides the verdict.
export const renderReport = (
  findings: readonly Finding[],
  skillPath: string,
  skillDir: string,
  strict: boolean,
): RenderedReport => {
  const warnings = messages(findings, "warn");
  const failures = messages(findings, "fail");
  const result = verdict(failures.length, warnings.length, strict);
  return {
    lines: [
      `Validation: ${skillPath}`,
      `Skill dir: ${skillDir}`,
      "",
      ...messages(findings, "pass").map((line) => `[PASS] ${line}`),
      ...warnings.map((line) => `[WARN] ${line}`),
      ...messages(findings, "drift").map((line) => `[DRIFT] ${line}`),
      ...failures.map((line) => `[FAIL] ${line}`),
      "",
      ...result.lines,
    ],
    exitCode: result.exitCode,
  };
};
