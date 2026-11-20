import {resolve} from "node:path";
import type {FileSystem} from "@xonovex/script-moon-common/file-system";
import type {Output} from "@xonovex/script-moon-common/output";
import {
  conflictingQueryOwners,
  missingValidationRoutingOwners,
  unresolvedOperationRationales,
} from "@xonovex/script-moon-skill-eval-common/routing-catalog";

const HELP = [
  "usage: moon-skill-validate-routing [-h] [catalog-root]",
  "",
  "Check the catalog's routing declarations:",
  "",
  "  owners     every skill owns at least one validation-split routing scenario:",
  "             a query text it alone marks should_trigger, which another skill",
  "             carries as a negative. Deleting a skill can silently strip",
  "             another skill's only pairing, which is what this catches.",
  "  conflicts  no query is marked should_trigger by two skills at once.",
  "  operations no rationale cites in backticks an operation that no catalog",
  "             skill owns a reference file for, which is how a retired",
  "             operation lingers.",
  "",
  "positional arguments:",
  "  catalog-root  skill catalog directory (defaults to packages/runtime/plugin)",
  "",
  "options:",
  "  -h, --help    show this help message and exit",
].join("\n");

const routingFindings = (
  catalogRoot: string,
  fs: FileSystem,
): readonly string[] => [
  ...missingValidationRoutingOwners(catalogRoot, fs).map(
    (owner) =>
      `routing: ${owner} owns no validation-split routing scenario; ` +
      "another skill must carry one of its validation positives as a negative",
  ),
  ...conflictingQueryOwners(catalogRoot, fs).map(
    ({owners, query}) =>
      `routing: ${owners.join(" and ")} both mark "${query}" should_trigger; ` +
      "exactly one skill owns a query and the others carry it as a negative",
  ),
  ...unresolvedOperationRationales(catalogRoot, fs).map(
    ({operation, skill}) =>
      `routing: ${skill} cites the '${operation}' operation, which no skill ` +
      "owns a reference file for; name the operation that exists or drop it",
  ),
];

export interface RenderedRoutingReport {
  readonly lines: readonly string[];
  readonly exitCode: number;
}

export const renderRoutingReport = (
  failures: readonly string[],
): RenderedRoutingReport =>
  failures.length > 0
    ? {
        lines: [
          ...failures.map((failure) => `[FAIL] ${failure}`),
          "",
          `Routing: ${String(failures.length)} finding(s)`,
        ],
        exitCode: 1,
      }
    : {
        lines: [
          "Routing: every skill owns a validation routing scenario, no query " +
            "is claimed twice, and every cited operation resolves",
        ],
        exitCode: 0,
      };

export interface ValidateRoutingDependencies {
  readonly fs: FileSystem;
  readonly output: Output;
}

/** Checks the routing declarations of a catalog and returns the exit code. */
export const main = (
  argv: readonly string[],
  cwd: string,
  dependencies: ValidateRoutingDependencies,
): number => {
  const {fs, output} = dependencies;
  if (argv.includes("-h") || argv.includes("--help")) {
    output.stdout(HELP);
    return 0;
  }
  const unknown = argv.find((argument) => argument.startsWith("-"));
  if (unknown !== undefined) {
    output.stderr(`Error: unrecognized argument: ${unknown}`);
    return 2;
  }

  const catalogRoot = resolve(cwd, argv[0] ?? "packages/runtime/plugin");
  let failures: readonly string[];
  try {
    failures = routingFindings(catalogRoot, fs);
  } catch (error) {
    output.stderr(
      `Error: ${error instanceof Error ? error.message : String(error)}`,
    );
    return 2;
  }
  const report = renderRoutingReport(failures);
  for (const line of report.lines) output.stdout(line);
  return report.exitCode;
};
