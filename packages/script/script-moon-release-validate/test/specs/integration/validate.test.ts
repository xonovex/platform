import {execFileSync} from "node:child_process";
import {dirname, resolve} from "node:path";
import {fileURLToPath} from "node:url";
import {describe, expect, it} from "vitest";

describe("release validator", () => {
  // Runs the built entrypoint as its own process against this checkout, so it
  // reports on the real repository rather than on a fixture.
  it("validates the repository release inputs end to end", () => {
    const sourceDirectory = dirname(fileURLToPath(import.meta.url));
    const entrypoint = resolve(sourceDirectory, "../../../dist/src/index.js");

    const output = execFileSync(process.execPath, [entrypoint], {
      encoding: "utf8",
    });

    expect(output).toContain("Release validation passed:");
  });
});
