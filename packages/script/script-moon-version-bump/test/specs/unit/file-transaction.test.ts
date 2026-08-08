import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {
  applyFileChanges,
  type FileTransactionIo,
} from "../../../src/file-transaction.js";

const DIRECTORY = "/repo/packages/example";

/**
 * A FileTransactionIo backed by a map. It reproduces the two behaviours the
 * transaction relies on: an exclusive write refuses an existing path, and a
 * rename of an absent path fails. `wrap` receives the working io so a case can
 * fail one operation and delegate the rest.
 */
const memoryIo = (
  initial: Readonly<Record<string, string>> = {},
  wrap: (base: FileTransactionIo) => Partial<FileTransactionIo> = () => ({}),
): {readonly io: FileTransactionIo; readonly files: Map<string, string>} => {
  const files = new Map(Object.entries(initial));
  let sequence = 0;
  const base: FileTransactionIo = {
    exists: (path) => files.has(path),
    write: (path, content) => {
      if (files.has(path)) {
        throw new Error(`EEXIST: file already exists, open '${path}'`);
      }
      files.set(path, content);
    },
    rename: (from, to) => {
      const content = files.get(from);
      if (content === undefined) {
        throw new Error(`ENOENT: no such file or directory, rename '${from}'`);
      }
      files.delete(from);
      files.set(to, content);
    },
    remove: (path) => {
      files.delete(path);
    },
    id: () => {
      sequence += 1;
      return String(sequence);
    },
  };
  return {files, io: {...base, ...wrap(base)}};
};

const EXISTING = join(DIRECTORY, "existing.json");
const ADDED = join(DIRECTORY, "added.json");
const FIRST = join(DIRECTORY, "first.json");
const SECOND = join(DIRECTORY, "second.json");
const TARGET = join(DIRECTORY, "package.json");

describe("applyFileChanges", () => {
  it("installs existing and new files and removes transaction artifacts", () => {
    const {files, io} = memoryIo({[EXISTING]: "existing-original\n"});

    applyFileChanges(
      [
        {path: EXISTING, content: "existing-updated\n"},
        {path: ADDED, content: "added\n"},
      ],
      io,
    );

    expect([...files.keys()].toSorted()).toEqual([ADDED, EXISTING]);
    expect(files.get(EXISTING)).toBe("existing-updated\n");
    expect(files.get(ADDED)).toBe("added\n");
  });

  it("restores every original file when a later rename fails", () => {
    const {files, io} = memoryIo(
      {[FIRST]: "first-original\n", [SECOND]: "second-original\n"},
      (base) => ({
        rename: (from, to) => {
          if (to === SECOND && from.endsWith(".temporary")) {
            throw new Error("simulated rename failure");
          }
          base.rename(from, to);
        },
      }),
    );

    expect(() => {
      applyFileChanges(
        [
          {path: FIRST, content: "first-updated\n"},
          {path: SECOND, content: "second-updated\n"},
        ],
        io,
      );
    }).toThrow("original files restored");
    expect([...files.keys()].toSorted()).toEqual([FIRST, SECOND]);
    expect(files.get(FIRST)).toBe("first-original\n");
    expect(files.get(SECOND)).toBe("second-original\n");
  });

  it("rejects duplicate targets before staging files", () => {
    const {files, io} = memoryIo();

    expect(() => {
      applyFileChanges(
        [
          {path: TARGET, content: "one"},
          {path: TARGET, content: "two"},
        ],
        io,
      );
    }).toThrow(`duplicate file change for ${TARGET}`);
    expect([...files.keys()]).toEqual([]);
  });

  it("removes staged files when writing a later temporary file fails", () => {
    let writes = 0;
    const {files, io} = memoryIo({}, (base) => ({
      write: (path, content) => {
        writes += 1;
        if (writes === 2) throw new Error("simulated write failure");
        base.write(path, content);
      },
    }));

    expect(() => {
      applyFileChanges(
        [
          {path: FIRST, content: "first\n"},
          {path: SECOND, content: "second\n"},
        ],
        io,
      );
    }).toThrow("original files restored");
    expect([...files.keys()]).toEqual([]);
  });

  it("reports the original failure together with rollback failures", () => {
    let renames = 0;
    const {io} = memoryIo({[TARGET]: "original\n"}, (base) => ({
      rename: (from, to) => {
        renames += 1;
        if (renames === 2) throw new Error("install failed");
        if (renames === 3) throw new Error("restore failed");
        base.rename(from, to);
      },
    }));

    expect(() => {
      applyFileChanges([{path: TARGET, content: "updated\n"}], io);
    }).toThrow("failed to apply file changes and restore every original file");
  });
});
