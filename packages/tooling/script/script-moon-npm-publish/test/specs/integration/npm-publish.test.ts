import {mkdtempSync, readFileSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {afterEach, beforeEach, expect, it} from "vitest";
import {main} from "../../../src/npm-publish.js";

// Drives main through the real disk adapter, so the manifest it injects into and
// restores is a file on disk rather than a memory entry.
let directory: string;

beforeEach(() => {
  directory = mkdtempSync(join(tmpdir(), "moon-npm-publish-"));
});

afterEach(() => {
  rmSync(directory, {recursive: true, force: true});
});

it("reads and restores the package manifest in the package root", () => {
  const manifest = '{"name":"example","version":"1.2.3"}\n';
  writeFileSync(join(directory, "package.json"), manifest);
  writeFileSync(
    join(directory, "platform.json"),
    JSON.stringify({os: ["linux"], cpu: ["x64"]}),
  );
  let during = "";

  const code = main([], directory, {
    fs: nodeFileSystem,
    output: memoryOutput(),
    isPublished: () => false,
    publish: () => {
      during = readFileSync(join(directory, "package.json"), "utf8");
    },
  });

  expect(code).toBe(0);
  expect(JSON.parse(during)).toMatchObject({os: ["linux"], cpu: ["x64"]});
  expect(readFileSync(join(directory, "package.json"), "utf8")).toBe(manifest);
});
