import {readdirSync, statSync} from "node:fs";
import {join} from "node:path";

export const isFile = (path: string): boolean => {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
};

export const isDirectory = (path: string): boolean => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

export const resolveGuideDirectory = (base: string): string => {
  if (isFile(join(base, "SKILL.md"))) return base;
  const nested = readdirSync(base)
    .map((entry) => join(base, entry))
    .filter((path) => isFile(join(path, "SKILL.md")));
  if (nested.length > 1) {
    throw new Error(
      `multiple SKILL.md found under ${base}; pass one explicitly`,
    );
  }
  return nested[0] ?? base;
};
