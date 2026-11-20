import type {ValidationReport} from "./validation.js";

export type Stream = "stdout" | "stderr";

export interface ReportLine {
  readonly stream: Stream;
  readonly text: string;
}

export interface RenderedReport {
  readonly lines: readonly ReportLine[];
  readonly exitCode: number;
}

// Issues and a failing verdict go to stderr so a passing run prints one line.
export const renderReport = (report: ValidationReport): RenderedReport => {
  const errors = report.issues.filter(({severity}) => severity === "error");
  const warnings = report.issues.filter(({severity}) => severity === "warning");
  const issueLines = report.issues.map((problem): ReportLine => ({
    stream: "stderr",
    text: `[${problem.severity === "warning" ? "WARN" : "FAIL"}] ${problem.path}: ${problem.code}: ${problem.message}`,
  }));
  return errors.length > 0
    ? {
        lines: [
          ...issueLines,
          {
            stream: "stderr",
            text: `Result: FAIL (${String(errors.length)} error(s), ${String(warnings.length)} warning(s), ${String(report.skills)} skill(s))`,
          },
        ],
        exitCode: 1,
      }
    : {
        lines: [
          ...issueLines,
          {
            stream: "stdout",
            text: `Result: PASS (${String(report.skills)} skill(s), ${String(warnings.length)} warning(s))`,
          },
        ],
        exitCode: 0,
      };
};
