import {resolve} from "node:path";
import {
  nodeFileSystem,
  type FileSystem,
} from "@xonovex/script-moon-common/file-system";
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
  "  operations no rationale names an operation of a catalog skill that owns no",
  "             reference file for it, which is how a retired operation lingers.",
  "",
  "positional arguments:",
  "  catalog-root  skill catalog directory (defaults to packages/skill)",
  "",
  "options:",
  "  -h, --help    show this help message and exit",
].join("\n");

export const main = (
  argv: readonly string[],
  fs: FileSystem = nodeFileSystem,
): number => {
  if (argv.includes("-h") || argv.includes("--help")) {
    console.log(HELP);
    return 0;
  }
  const unknown = argv.find((argument) => argument.startsWith("-"));
  if (unknown !== undefined) {
    process.stderr.write(`Error: unrecognized argument: ${unknown}\n`);
    return 2;
  }

  const catalogRoot = resolve(argv[0] ?? "packages/skill");
  const failures: string[] = [];
  try {
    for (const owner of missingValidationRoutingOwners(catalogRoot, fs)) {
      failures.push(
        `routing: ${owner} owns no validation-split routing scenario; ` +
          "another skill must carry one of its validation positives as a negative",
      );
    }
    for (const {owners, query} of conflictingQueryOwners(catalogRoot, fs)) {
      failures.push(
        `routing: ${owners.join(" and ")} both mark "${query}" should_trigger; ` +
          "exactly one skill owns a query and the others carry it as a negative",
      );
    }
    for (const {operation, skill} of unresolvedOperationRationales(
      catalogRoot,
      fs,
    )) {
      failures.push(
        `routing: ${skill} cites the '${operation}' operation, which no skill ` +
          "owns a reference file for; name the operation that exists or drop it",
      );
    }
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    process.stderr.write(`Error: ${detail}\n`);
    return 2;
  }

  if (failures.length > 0) {
    for (const failure of failures) console.log(`[FAIL] ${failure}`);
    console.log();
    console.log(`Routing: ${String(failures.length)} finding(s)`);
    return 1;
  }

  console.log(
    "Routing: every skill owns a validation routing scenario, no query " +
      "is claimed twice, and every cited operation resolves",
  );
  return 0;
};
