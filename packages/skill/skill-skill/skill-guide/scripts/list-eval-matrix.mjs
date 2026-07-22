#!/usr/bin/env node
import {existsSync, readdirSync, readFileSync, statSync} from "node:fs";
import {join, resolve} from "node:path";
import {fileURLToPath} from "node:url";

const SHARED_EVAL_PATHS = [
  ".github/workflows/skill-evals.yml",
  ".moon/tasks/tag-skill.yml",
  "packages/script/script-moon-common/",
  "packages/script/script-moon-skill-eval-outputs/",
  "packages/script/script-moon-skill-eval-triggers/",
  "packages/skill/skill-skill/skill-guide/scripts/",
];

export const catalogEntries = (catalogRoot) =>
  readdirSync(catalogRoot)
    .flatMap((project) => {
      const directory = join(catalogRoot, project);
      if (!project.startsWith("skill-") || !statSync(directory).isDirectory()) {
        return [];
      }
      const guide = readdirSync(directory).find((entry) =>
        existsSync(join(directory, entry, "SKILL.md")),
      );
      return guide === undefined
        ? []
        : [
            {
              guide,
              package: project.replace(/^skill-/, ""),
              project,
            },
          ];
    })
    .toSorted((left, right) => left.project.localeCompare(right.project));

const rotatedLimit = (entries, limit, offset) => {
  if (entries.length <= limit) return entries;
  const start = ((offset % entries.length) + entries.length) % entries.length;
  return [...entries.slice(start), ...entries.slice(0, start)].slice(0, limit);
};

export const selectEvalMatrix = ({entries, changedFiles, limit, offset}) => {
  if (changedFiles === undefined) return rotatedLimit(entries, limit, offset);
  const selectAll = changedFiles.some((path) =>
    SHARED_EVAL_PATHS.some((shared) =>
      shared.endsWith("/") ? path.startsWith(shared) : path === shared,
    ),
  );
  const selected = selectAll
    ? entries
    : entries.filter(({project}) =>
        changedFiles.some((path) =>
          path.startsWith(`packages/skill/${project}/`),
        ),
      );
  return selectAll ? rotatedLimit(selected, limit, offset) : selected;
};

const main = () => {
  const catalogRoot = resolve(process.argv[2] ?? "packages/skill");
  const changedFile = process.argv[3];
  const limit = Number(process.argv[4] ?? "12");
  const offset = Number(process.argv[5] ?? "0");
  if (!Number.isSafeInteger(limit) || limit < 1) {
    throw new Error("limit must be a positive integer");
  }
  if (!Number.isSafeInteger(offset)) {
    throw new Error("offset must be an integer");
  }
  const changedFiles =
    changedFile === undefined || changedFile === "-"
      ? undefined
      : readFileSync(resolve(changedFile), "utf8")
          .split(/\r?\n/)
          .filter(Boolean);
  process.stdout.write(
    `${JSON.stringify(
      selectEvalMatrix({
        entries: catalogEntries(catalogRoot),
        changedFiles,
        limit,
        offset,
      }),
    )}\n`,
  );
};

if (
  process.argv[1] !== undefined &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  main();
}
