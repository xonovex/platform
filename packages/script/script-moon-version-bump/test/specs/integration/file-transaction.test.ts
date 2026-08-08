import {readdirSync, readFileSync, writeFileSync} from "node:fs";
import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {applyFileChanges} from "../../../src/file-transaction.js";
import {temporaryDirectories} from "../../util/git-repository.js";

// The rollback rules are driven from an in-memory io in test/specs/unit; what is
// left here is the default io itself, whose exclusive write and rename reach a
// real disk and are the reason the transaction can restore what it replaced.
describe("applyFileChanges against a real disk", () => {
  const directory = temporaryDirectories();

  it("installs existing and new files and removes transaction artifacts", () => {
    const root = directory("version-transaction-");
    const existing = join(root, "existing.json");
    const added = join(root, "added.json");
    writeFileSync(existing, "existing-original\n");

    applyFileChanges([
      {path: existing, content: "existing-updated\n"},
      {path: added, content: "added\n"},
    ]);

    expect(readFileSync(existing, "utf8")).toBe("existing-updated\n");
    expect(readFileSync(added, "utf8")).toBe("added\n");
    expect(readdirSync(root).toSorted()).toEqual([
      "added.json",
      "existing.json",
    ]);
  });

  it("leaves the directory untouched when the changes are rejected", () => {
    const root = directory("version-transaction-");
    const target = join(root, "package.json");

    expect(() => {
      applyFileChanges([
        {path: target, content: "one"},
        {path: target, content: "two"},
      ]);
    }).toThrow(`duplicate file change for ${target}`);
    expect(readdirSync(root)).toEqual([]);
  });
});
