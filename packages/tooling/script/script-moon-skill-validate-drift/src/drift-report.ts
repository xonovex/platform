import type {DriftLintMode} from "@xonovex/script-moon-common/drift-budgets";

export interface DriftResult {
  readonly fileCount: number;
  readonly manifestErrors: readonly string[];
  readonly findings: readonly string[];
  readonly mode: DriftLintMode;
}

export interface RenderedDriftReport {
  readonly lines: readonly string[];
  readonly exitCode: number;
}

// A malformed manifest fails in either mode; a drift finding fails only when
// the mode enforces it.
export const renderDriftReport = (result: DriftResult): RenderedDriftReport => {
  const {fileCount, findings, manifestErrors, mode} = result;
  const findingLevel = mode === "enforce" ? "FAIL" : "DRIFT";
  return {
    lines: [
      ...manifestErrors.map((error) => `[FAIL] ${error}`),
      ...findings.map((finding) => `[${findingLevel}] ${finding}`),
      "",
      `Drift (${mode} mode): ${String(fileCount)} files, ` +
        `${String(findings.length)} finding(s)`,
    ],
    exitCode:
      manifestErrors.length > 0 || (mode === "enforce" && findings.length > 0)
        ? 1
        : 0,
  };
};
