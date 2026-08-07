import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach} from "vitest";

/**
 * Returns a factory that writes a throwaway skill directory holding a SKILL.md,
 * an empty references directory, and the given SOURCES.md, and removes every
 * directory it created after each test. Call it inside the suite that uses it so
 * the cleanup registers on that suite.
 */
export const skillDirectories = (): ((
  name: string,
  sources: string,
) => string) => {
  const roots: string[] = [];

  afterEach(() => {
    for (const root of roots) {
      rmSync(root, {recursive: true, force: true});
    }
    roots.length = 0;
  });

  return (name, sources) => {
    const root = mkdtempSync(join(tmpdir(), "skill-audit-sources-"));
    roots.push(root);
    const skill = join(root, name);
    mkdirSync(join(skill, "references"), {recursive: true});
    writeFileSync(join(skill, "SKILL.md"), `# ${name}\n`);
    writeFileSync(join(skill, "SOURCES.md"), sources);
    return skill;
  };
};
