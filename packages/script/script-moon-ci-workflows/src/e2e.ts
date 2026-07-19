export const E2E_SUITES = [
  {name: "e2e", task: "agent-operator-go:go-test-e2e"},
  {name: "e2e-gvisor", task: "agent-operator-go:go-test-e2e-gvisor"},
  {name: "e2e-kata", task: "agent-operator-go:go-test-e2e-kata"},
  {name: "e2e-coco", task: "agent-operator-go:go-test-e2e-coco"},
] as const;

export type E2eSuite = (typeof E2E_SUITES)[number]["name"];

export const e2eMatrix = (): readonly E2eSuite[] =>
  E2E_SUITES.map(({name}) => name);

export const e2eTask = (suite: string): string => {
  const selected = E2E_SUITES.find(({name}) => name === suite);
  if (selected === undefined) {
    throw new Error(`unknown E2E suite: ${suite}`);
  }
  return selected.task;
};

const countResult = (lines: readonly string[], marker: string): number =>
  lines.filter((line) => line.includes(marker)).length;

export const renderE2eSummary = (suite: string, log: string): string => {
  const lines = log.split(/\r?\n/u);
  const passed = countResult(lines, "--- PASS:");
  const failed = countResult(lines, "--- FAIL:");
  const skippedLines = lines.filter((line) => line.includes("--- SKIP:"));
  const rows = [
    `### ${suite}`,
    "",
    "| Executed (pass) | Executed (fail) | Skipped |",
    "| --- | --- | --- |",
    `| ${String(passed)} | ${String(failed)} | ${String(skippedLines.length)} |`,
    "",
  ];

  if (skippedLines.length > 0) {
    rows.push(
      "Skipped tests (kata/coco isolation tests self-skip in unprivileged Kind):",
      "",
      "```",
      ...skippedLines,
      "```",
      "",
    );
  }

  return `${rows.join("\n")}\n`;
};

export const renderMissingE2eSummary = (suite: string): string =>
  `### ${suite}\n\nNo test log produced (suite failed before running tests).\n`;
