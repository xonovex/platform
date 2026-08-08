import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {afterEach} from "vitest";

export const MEMORY_ROOT = "/catalog";

/**
 * Returns a factory for a skill held in memory: its SKILL.md, an empty
 * references directory, and the given SOURCES.md. A case reads back what the
 * auditor wrote without leaving anything on disk, which is why the unit tier
 * uses this one.
 */
const memorySkill = (
  name: string,
  sources: string,
): {readonly skill: string; readonly fs: FileSystem} => {
  const skill = join(MEMORY_ROOT, name);
  const fs = memoryFileSystem({
    directories: [join(skill, "references")],
    files: {
      [join(skill, "SKILL.md")]: `# ${name}\n`,
      [join(skill, "SOURCES.md")]: sources,
    },
  });
  return {fs, skill};
};

export const memorySkillDirectories = (): typeof memorySkill => memorySkill;

/**
 * Returns a factory that writes a throwaway skill directory on disk and removes
 * every directory it created after each test. The integration tier uses this one
 * because it drives git against a real checkout. Call it inside the suite that
 * uses it so the cleanup registers on that suite.
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
