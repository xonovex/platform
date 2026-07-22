import {readdirSync, readFileSync, statSync} from "node:fs";
import {join, resolve} from "node:path";
import {describe, expect, it} from "vitest";

const workspaceRoot = resolve(import.meta.dirname, "../../../../..");
const skillPackages = join(workspaceRoot, "packages/skill");

describe("skill catalog Moon task ownership", () => {
  it("keeps strict validation and source-audit commands in shared tasks", () => {
    const overrides = [];
    for (const packageName of readdirSync(skillPackages).toSorted()) {
      const packageDirectory = join(skillPackages, packageName);
      if (
        !packageName.startsWith("skill-") ||
        !statSync(packageDirectory).isDirectory()
      ) {
        continue;
      }
      const moon = readFileSync(join(packageDirectory, "moon.yml"), "utf8");
      if (/^  skill-(?:validate|audit-sources):/m.test(moon)) {
        overrides.push(packageName);
      }
    }

    expect(overrides).toEqual([]);
  });
});
