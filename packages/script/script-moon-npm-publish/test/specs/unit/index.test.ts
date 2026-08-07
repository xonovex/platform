import {mkdtempSync, rmSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {expect, it, vi} from "vitest";
import {defaultDependencies, main} from "../../../src/cli.js";
import type {PublishDependencies} from "../../../src/publish.js";

it("passes the dry-run option through the executable entrypoint", () => {
  const publish = vi.fn();
  const dependencies: PublishDependencies = {
    readPackageJson: () => '{"name":"example","version":"1.2.3"}',
    writePackageJson: vi.fn(),
    currentDirectory: () => "/workspace/package",
    readPlatformMeta: () => {
      return;
    },
    isPublished: () => false,
    publish,
    log: vi.fn(),
  };

  expect(main(["--dry-run"], dependencies)).toBe(0);
  expect(publish).toHaveBeenCalledWith(true);
});

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
