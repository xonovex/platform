#!/usr/bin/env node
import {existsSync, readdirSync, statSync} from "node:fs";
import {homedir} from "node:os";
import {join, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const RETIRED = new Map([
  ["xonovex-skill-general-fp", "xonovex-skill-fp"],
  ["xonovex-skill-general-oop", "xonovex-skill-oop"],
  ["xonovex-skill-insights", "xonovex-skill-reflect"],
  ["xonovex-skill-prompt", "xonovex-skill-command"],
]);

const defaultRoots = () => [
  join(homedir(), ".codex", "plugins", "cache", "xonovex-marketplace"),
  join(homedir(), ".claude", "plugins", "cache", "xonovex-marketplace"),
];

const parseArgs = (args) => {
  const roots = [];
  let json = false;
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === "--json") {
      json = true;
    } else if (argument === "--root") {
      const root = args[index + 1];
      if (root === undefined) throw new Error("--root requires a path");
      roots.push(resolve(root));
      index += 1;
    } else if (argument === "-h" || argument === "--help") {
      process.stdout.write(
        "Usage: check-retired-plugins.mjs [--root PATH]... [--json]\n",
      );
      process.exit(0);
    } else {
      throw new Error(`unknown option: ${argument}`);
    }
  }
  return {json, roots: roots.length > 0 ? roots : defaultRoots()};
};

const installedVersions = (pluginDirectory) => {
  if (
    !existsSync(pluginDirectory) ||
    !statSync(pluginDirectory).isDirectory()
  ) {
    return [];
  }
  return readdirSync(pluginDirectory)
    .filter((entry) => statSync(join(pluginDirectory, entry)).isDirectory())
    .toSorted();
};

export const findRetired = (roots) => {
  const findings = [];
  for (const root of roots) {
    if (!existsSync(root) || !statSync(root).isDirectory()) continue;
    for (const [retired, replacement] of RETIRED) {
      const pluginDirectory = join(root, retired);
      for (const version of installedVersions(pluginDirectory)) {
        findings.push({
          path: join(pluginDirectory, version),
          replacement,
          retired,
          root,
          version,
        });
      }
    }
  }
  return findings;
};

const main = () => {
  const {json, roots} = parseArgs(process.argv.slice(2));
  const findings = findRetired(roots);
  if (json) {
    process.stdout.write(JSON.stringify({findings, roots}, null, 2) + "\n");
  } else if (findings.length === 0) {
    process.stdout.write("No retired Xonovex plugin bundles found.\n");
  } else {
    process.stdout.write(
      `Found ${String(findings.length)} retired Xonovex plugin bundle(s):\n`,
    );
    for (const finding of findings) {
      process.stdout.write(
        `- ${finding.retired}@${finding.version} -> ${finding.replacement} (${finding.path})\n`,
      );
    }
    process.stdout.write(
      "Uninstall these entries through the host plugin browser, install the replacements, and start a new session.\n",
    );
  }
  return findings.length === 0 ? 0 : 1;
};

if (
  process.argv[1] !== undefined &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  try {
    process.exit(main());
  } catch (error) {
    process.stderr.write(
      `Error: ${error instanceof Error ? error.message : String(error)}\n`,
    );
    process.exit(2);
  }
}
