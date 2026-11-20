import type {Finding} from "@xonovex/script-moon-skill-catalog-common/finding";

export interface RenderedLinkReport {
  readonly lines: readonly string[];
  readonly exitCode: number;
}

const linesAt = (
  findings: readonly Finding[],
  level: Finding["level"],
  label: string,
): readonly string[] =>
  findings
    .filter((finding) => finding.level === level)
    .map(({message}) => `[${label}] ${message}`);

/** Groups the findings by level and closes with the verdict. */
export const renderLinkReport = (
  findings: readonly Finding[],
): RenderedLinkReport => {
  const errors = findings.filter(({level}) => level === "fail").length;
  return {
    lines: [
      ...linesAt(findings, "pass", "PASS"),
      ...linesAt(findings, "fail", "FAIL"),
      ...linesAt(findings, "warn", "WARN"),
      errors > 0
        ? `Result: FAIL (${String(errors)} dependency or link error(s))`
        : "Result: PASS (links, packaging, and hard dependencies are valid)",
    ],
    exitCode: errors > 0 ? 1 : 0,
  };
};
