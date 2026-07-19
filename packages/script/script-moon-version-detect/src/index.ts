#!/usr/bin/env node
import {dirname} from "node:path";
import {fileURLToPath} from "node:url";
import {
  findWorkspaceRoot,
  parseCliArgs,
  queryMoonProjects,
} from "@xonovex/script-moon-common";
import {
  createGitReader,
  detectVersionChanges,
  resolveGitRef,
} from "./detect.js";

const ROOT_DIR = findWorkspaceRoot(dirname(fileURLToPath(import.meta.url)));

const {values, positionals} = parseCliArgs({
  name: "moon-version-detect",
  description: "Detect moon projects with version changes since a git ref",
  options: {
    ref: {
      type: "string",
      short: "r",
      description: "Git ref to compare against (default: HEAD~1)",
    },
  },
});
const ref = (values.ref as string | undefined) ?? positionals[0] ?? "HEAD~1";

const commit = resolveGitRef(ROOT_DIR, ref);
const projects = queryMoonProjects(ROOT_DIR);
const changed = detectVersionChanges(
  ROOT_DIR,
  commit,
  projects,
  createGitReader(ROOT_DIR),
);

console.log(JSON.stringify(changed));
