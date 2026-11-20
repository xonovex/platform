import {execFileSync} from "node:child_process";
import {parseMoonProjects, type MoonProject} from "../moon-query.js";
import {resolveExecutable} from "./executable-node.js";

/** Lists the workspace projects by running moon in rootDir. */
export const queryMoonProjects = (rootDir: string): readonly MoonProject[] =>
  parseMoonProjects(
    execFileSync(resolveExecutable("npx"), ["moon", "query", "projects"], {
      encoding: "utf8",
      maxBuffer: 10 * 1024 * 1024,
      cwd: rootDir,
    }),
  );
