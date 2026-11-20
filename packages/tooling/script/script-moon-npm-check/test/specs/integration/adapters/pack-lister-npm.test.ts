import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {npmPackLister} from "../../../../src/adapters/pack-lister-npm.js";

describe("npmPackLister", () => {
  const roots: string[] = [];

  const packageRoot = (manifestText: string): string => {
    const root = mkdtempSync(join(tmpdir(), "npm-check-pack-"));
    roots.push(root);
    writeFileSync(join(root, "package.json"), manifestText);
    return root;
  };

  afterEach(() => {
    for (const root of roots) rmSync(root, {recursive: true, force: true});
    roots.length = 0;
  });

  it("lists the files npm would pack", () => {
    const root = packageRoot(
      JSON.stringify({
        name: "pack-lister-fixture",
        version: "1.0.0",
        files: ["dist"],
      }),
    );
    mkdirSync(join(root, "dist"));
    writeFileSync(join(root, "dist", "index.js"), "export {};\n");
    writeFileSync(join(root, "unpacked.js"), "export {};\n");

    expect([...npmPackLister(root)].toSorted()).toEqual([
      "dist/index.js",
      "package.json",
    ]);
  });

  it("fails with the npm error for a manifest npm cannot parse", () => {
    const root = packageRoot("{");

    expect(() => npmPackLister(root)).toThrow("npm pack failed");
  });
});
