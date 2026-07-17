#!/usr/bin/env node
import {checkCrossPackageLinks} from "./cross-package-links.js";
import {type LinkReport} from "./reference-file-links.js";

const HELP_TEXT = [
  "usage: moon-skill-links [-h] [repo-root]",
  "",
  "Validate that every boundary-crossing relative markdown link across the",
  "composition surface (skill SKILL.md and references, command-workflow docs and",
  "commands) resolves to a real file.",
  "",
  "positional arguments:",
  "  repo-root   repository root to scan (defaults to the current directory)",
  "",
  "options:",
  "  -h, --help  show this help message and exit",
].join("\n");

const main = (argv: readonly string[]): number => {
  if (argv.includes("-h") || argv.includes("--help")) {
    console.log(HELP_TEXT);
    return 0;
  }

  const positional = argv.find((arg) => !arg.startsWith("-"));
  const repoRoot = positional ?? process.cwd();

  const passes: string[] = [];
  const errors: string[] = [];
  const report: LinkReport = {
    addPass: (message) => {
      passes.push(message);
    },
    addFail: (message) => {
      errors.push(message);
    },
  };

  checkCrossPackageLinks(repoRoot, report);

  for (const line of passes) {
    console.log(`[PASS] ${line}`);
  }
  for (const line of errors) {
    console.log(`[FAIL] ${line}`);
  }

  if (errors.length > 0) {
    console.log(
      `Result: FAIL (${String(errors.length)} broken cross-package link(s))`,
    );
    return 1;
  }
  console.log("Result: PASS (no broken cross-package links)");
  return 0;
};

process.exit(main(process.argv.slice(2)));
