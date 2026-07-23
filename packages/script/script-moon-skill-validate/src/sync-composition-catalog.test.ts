import {
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {syncCompositionCatalog} from "./sync-composition-catalog.js";

const temporaryDirectories: string[] = [];

const createRepository = (
  source: string,
  snapshot: string,
): {readonly root: string; readonly snapshotPath: string} => {
  const root = mkdtempSync(join(tmpdir(), "composition-catalog-sync-"));
  temporaryDirectories.push(root);
  const sourceDirectory = join(root, "packages", "skill");
  const snapshotDirectory = join(
    sourceDirectory,
    "skill-workflow",
    "workflow-guide",
    "assets",
  );
  mkdirSync(snapshotDirectory, {recursive: true});
  writeFileSync(join(sourceDirectory, "composition-catalog.json"), source);
  const snapshotPath = join(snapshotDirectory, "composition-catalog.json");
  writeFileSync(snapshotPath, snapshot);
  return {root, snapshotPath};
};

afterEach(() => {
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

describe("composition catalog synchronization", () => {
  it("copies exact canonical bytes into a stale packaged snapshot", () => {
    const repository = createRepository('{"contractVersion":"2.0.0"}\n', "{}");

    expect(syncCompositionCatalog(repository.root).changed).toBe(true);
    expect(readFileSync(repository.snapshotPath, "utf8")).toBe(
      '{"contractVersion":"2.0.0"}\n',
    );
  });

  it("does not rewrite a current packaged snapshot", () => {
    const repository = createRepository("{}\n", "{}\n");

    expect(syncCompositionCatalog(repository.root).changed).toBe(false);
  });
});
