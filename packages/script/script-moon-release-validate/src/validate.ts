export interface Checker {
  check: (condition: boolean, message: string) => void;
  checks: () => number;
  failures: () => readonly string[];
}

export const createChecker = (): Checker => {
  const failures: string[] = [];
  let checks = 0;
  return {
    check: (condition, message) => {
      checks += 1;
      if (!condition) failures.push(message);
    },
    checks: () => checks,
    failures: () => failures,
  };
};

export const tableIds = (content: string, prefix: string): Set<string> =>
  new Set(
    [
      ...content.matchAll(
        new RegExp(String.raw`^\|\s*(${prefix}[^|\s]+)\s*\|`, "gm"),
      ),
    ].flatMap((match) => (match[1] === undefined ? [] : [match[1]])),
  );

export const markdownLinkTargets = (content: string): string[] =>
  [...content.matchAll(/\[[^\]]+\]\(([^)]+)\)/g)]
    .flatMap((match) => {
      const target = match[1]?.split("#", 1)[0];
      return target === undefined ? [] : [target];
    })
    .filter((target) => target !== "" && !/^(?:https?:|mailto:)/.test(target));

export const forbiddenClaims: readonly RegExp[] = [
  /all harnesses (?:have|support|provide)/i,
  /skills? (?:are|is|provide|provides) enforcement/i,
  /install(?:ing|ed)? (?:a )?skills? (?:enforces|proves)/i,
  /workflow ya?ml (?:is|are) required/i,
  /(?:provides?|ensures?|establishes?|achieves?) automatic compliance/i,
  /silently launch(?:es|ing)? (?:a |an )?(?:child )?agent/i,
];

const cellCount = (row: string): number =>
  row.trim().replace(/^\|/, "").replace(/\|$/, "").split("|").length;

export const tableShapeFailures = (
  content: string,
  label: string,
): string[] => {
  const failures: string[] = [];
  const lines = content.split("\n");
  let headerCells: number | null = null;
  for (const [index, line] of lines.entries()) {
    const isRow = line.trimStart().startsWith("|");
    if (!isRow) {
      headerCells = null;
      continue;
    }
    if (headerCells === null) {
      headerCells = cellCount(line);
      continue;
    }
    if (/^\s*\|[\s:-]+\|/.test(line) && /^[\s|:-]+$/.test(line)) continue;
    if (cellCount(line) !== headerCells) {
      failures.push(
        `${label} line ${String(index + 1)}: row has ${String(cellCount(line))} cells, header has ${String(headerCells)}`,
      );
    }
  }
  return failures;
};

export const legendClasses = (content: string): Set<string> =>
  new Set(
    [...content.matchAll(/^- \*\*([^*]+)\*\*/gm)].flatMap((match) =>
      match[1] === undefined ? [] : [match[1].trim()],
    ),
  );

export const offLegendClassifications = (
  content: string,
  legend: Set<string>,
): string[] =>
  [...content.matchAll(/^\| (D-\d{3}) \| [^|]* \| ([^|]+) \|/gm)].flatMap(
    (match) => {
      const id = match[1];
      const classification = match[2]?.trim();
      if (id === undefined || classification === undefined) return [];
      return legend.has(classification)
        ? []
        : [`${id}: classification '${classification}' is not in the legend`];
    },
  );

export const blanketIdBlockFailures = (content: string): string[] => {
  const blocks = new Map<string, {rows: number; distinct: Set<string>}>();
  for (const match of content.matchAll(
    /^\| (subplan-\d{2}-[^ ]+) \| \d+ \| [^|]* \| ([^|]*\|[^|]*\|[^|]*) \|$/gm,
  )) {
    const subplan = match[1];
    const block = match[2];
    if (subplan === undefined || block === undefined) continue;
    const entry = blocks.get(subplan) ?? {rows: 0, distinct: new Set<string>()};
    entry.rows += 1;
    entry.distinct.add(block.trim());
    blocks.set(subplan, entry);
  }
  return [...blocks.entries()].flatMap(([subplan, {rows, distinct}]) =>
    rows > 1 && distinct.size === 1
      ? [`${subplan}: every task row carries an identical ID block`]
      : [],
  );
};

export const releaseWorkflowFailures = (content: string): string[] => {
  const failures: string[] = [];
  if (/^\s{2}push:\s*$/m.test(content) || content.includes("refs/tags/")) {
    failures.push("release workflow must not publish from pushed tags");
  }
  if (/inputs\.dry_run\s*!=\s*true/.test(content)) {
    failures.push(
      "release workflow must not derive publish authority from a dispatch input",
    );
  }

  const publishStep =
    /\n\s+- name: Publish\n([\s\S]*?)(?=\n\s+- (?:name:|uses:))/u.exec(
      content,
    )?.[1];
  if (
    publishStep === undefined ||
    !/if:\s*github\.event_name == 'pull_request'/.test(publishStep)
  ) {
    failures.push("publish step must be restricted to the pull_request event");
  }

  const dryRunStep =
    /\n\s+- name: Publish \(dry run\)\n([\s\S]*?)(?=\n\s+- (?:name:|uses:))/u.exec(
      content,
    )?.[1];
  if (
    dryRunStep === undefined ||
    !/if:\s*github\.event_name == 'workflow_dispatch'/.test(dryRunStep)
  ) {
    failures.push(
      "manual dispatch must be restricted to the dry-run publish step",
    );
  }

  return failures;
};
