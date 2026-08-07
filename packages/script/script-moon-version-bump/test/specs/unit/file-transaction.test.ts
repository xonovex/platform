import {
  existsSync,
  mkdtempSync,
  readdirSync,
  readFileSync,
  renameSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it, vi} from "vitest";
import {
  applyFileChanges,
  type FileTransactionIo,
} from "../../../src/file-transaction.js";

const directories: string[] = [];

const temporaryDirectory = (): string => {
  const directory = mkdtempSync(join(tmpdir(), "version-transaction-"));
  directories.push(directory);
  return directory;
};

afterEach(() => {
  for (const directory of directories) {
    rmSync(directory, {recursive: true, force: true});
  }
  directories.length = 0;
});

describe("applyFileChanges", () => {
  it("installs existing and new files and removes transaction artifacts", () => {
    const directory = temporaryDirectory();
    const existing = join(directory, "existing.json");
    const added = join(directory, "added.json");
    writeFileSync(existing, "existing-original\n");

    applyFileChanges([
      {path: existing, content: "existing-updated\n"},
      {path: added, content: "added\n"},
    ]);

    expect(readFileSync(existing, "utf8")).toBe("existing-updated\n");
    expect(readFileSync(added, "utf8")).toBe("added\n");
    expect(readdirSync(directory).toSorted()).toEqual([
      "added.json",
      "existing.json",
    ]);
  });

  it("restores every original file when a later rename fails", () => {
    const directory = temporaryDirectory();
    const first = join(directory, "first.json");
    const second = join(directory, "second.json");
    writeFileSync(first, "first-original\n");
    writeFileSync(second, "second-original\n");

    let id = 0;
    const rename = vi.fn((from: string, to: string): void => {
      if (to === second && from.endsWith(".temporary")) {
        throw new Error("simulated rename failure");
      }
      renameSync(from, to);
    });
    const io: FileTransactionIo = {
      exists: existsSync,
      write: (path, content) => {
        writeFileSync(path, content, "utf8");
      },
      rename,
      remove: (path) => {
        rmSync(path, {force: true});
      },
      id: () => {
        const value = String(id);
        id += 1;
        return value;
      },
    };

    expect(() => {
      applyFileChanges(
        [
          {path: first, content: "first-updated\n"},
          {path: second, content: "second-updated\n"},
        ],
        io,
      );
    }).toThrow("original files restored");
    expect(readFileSync(first, "utf8")).toBe("first-original\n");
    expect(readFileSync(second, "utf8")).toBe("second-original\n");
    expect(readdirSync(directory).toSorted()).toEqual([
      "first.json",
      "second.json",
    ]);
  });

  it("rejects duplicate targets before staging files", () => {
    const directory = temporaryDirectory();
    const target = join(directory, "package.json");

    expect(() => {
      applyFileChanges([
        {path: target, content: "one"},
        {path: target, content: "two"},
      ]);
    }).toThrow(`duplicate file change for ${target}`);
    expect(readdirSync(directory)).toEqual([]);
  });

  it("removes staged files when writing a later temporary file fails", () => {
    const directory = temporaryDirectory();
    const first = join(directory, "first.json");
    const second = join(directory, "second.json");
    let writes = 0;
    const io: FileTransactionIo = {
      exists: existsSync,
      write: (path, content) => {
        writes += 1;
        if (writes === 2) throw new Error("simulated write failure");
        writeFileSync(path, content, "utf8");
      },
      rename: renameSync,
      remove: (path) => {
        rmSync(path, {force: true});
      },
      id: () => String(writes),
    };

    expect(() => {
      applyFileChanges(
        [
          {path: first, content: "first\n"},
          {path: second, content: "second\n"},
        ],
        io,
      );
    }).toThrow("original files restored");
    expect(readdirSync(directory)).toEqual([]);
  });

  it("reports the original failure together with rollback failures", () => {
    const directory = temporaryDirectory();
    const target = join(directory, "package.json");
    writeFileSync(target, "original\n");
    let renameCount = 0;
    const io: FileTransactionIo = {
      exists: existsSync,
      write: (path, content) => {
        writeFileSync(path, content, "utf8");
      },
      rename: (from, to) => {
        renameCount += 1;
        if (renameCount === 2) throw new Error("install failed");
        if (renameCount === 3) throw new Error("restore failed");
        renameSync(from, to);
      },
      remove: (path) => {
        rmSync(path, {force: true});
      },
      id: () => "failure",
    };

    expect(() => {
      applyFileChanges([{path: target, content: "updated\n"}], io);
    }).toThrow("failed to apply file changes and restore every original file");
  });
});
