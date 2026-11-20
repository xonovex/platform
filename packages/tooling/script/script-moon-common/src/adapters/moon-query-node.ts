import {execFileSync} from "node:child_process";
import type {ExecutableResolver} from "../executable.js";
import {parseMoonProjects, type MoonProject} from "../moon-query.js";

/** Lists the workspace projects by running moon in rootDir. */
export const queryMoonProjects =
  (resolveExecutable: ExecutableResolver) =>
  (rootDir: string): readonly MoonProject[] =>
    parseMoonProjects(
      execFileSync(resolveExecutable("npx"), ["moon", "query", "projects"], {
        encoding: "utf8",
        maxBuffer: 10 * 1024 * 1024,
        cwd: rootDir,
      }),
    );
