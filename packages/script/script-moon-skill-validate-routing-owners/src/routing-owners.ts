import {resolve} from "node:path";
import {missingValidationRoutingOwners} from "@xonovex/script-moon-skill-eval-common/routing-catalog";

const HELP = [
  "usage: moon-skill-validate-routing-owners [-h] [catalog-root]",
  "",
  "Check that every catalog skill owns at least one validation-split routing",
  "scenario: a query text it alone marks should_trigger, which another skill",
  "carries as a negative. Deleting a skill can silently strip another skill's",
  "only pairing, which is what this catches.",
  "",
  "positional arguments:",
  "  catalog-root  skill catalog directory (defaults to packages/skill)",
  "",
  "options:",
  "  -h, --help    show this help message and exit",
].join("\n");

export const main = (argv: readonly string[]): number => {
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
  let missing: readonly string[];
  try {
    missing = missingValidationRoutingOwners(catalogRoot);
  } catch (error) {
    const detail = error instanceof Error ? error.message : String(error);
    process.stderr.write(`Error: ${detail}\n`);
    return 2;
  }

  if (missing.length > 0) {
    for (const owner of missing) {
      console.log(
        `[FAIL] routing: ${owner} owns no validation-split routing scenario; ` +
          "another skill must carry one of its validation positives as a negative",
      );
    }
    console.log();
    console.log(`Routing owners: ${String(missing.length)} skill(s) unpaired`);
    return 1;
  }

  console.log("Routing owners: every skill owns a validation routing scenario");
  return 0;
};
