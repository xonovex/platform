import {mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {expect, it} from "vitest";
import {defaultDependencies} from "../../../src/cli.js";

// The real dependencies read and write package.json relative to the process
// directory, so proving they agree on that directory means changing it. That
// mutates state the whole test process shares, which is why this case sits in the
// integration tier rather than beside the injected-dependency cases.
it("reads and writes package metadata relative to the current directory", () => {
  const originalDirectory = process.cwd();
  const directory = mkdtempSync(join(tmpdir(), "moon-npm-publish-"));
  try {
    process.chdir(directory);
    defaultDependencies.writePackageJson('{"name":"example"}\n');
    expect(defaultDependencies.readPackageJson()).toBe('{"name":"example"}\n');
    expect(defaultDependencies.currentDirectory()).toBe(directory);
  } finally {
    process.chdir(originalDirectory);
    rmSync(directory, {recursive: true, force: true});
  }
});
