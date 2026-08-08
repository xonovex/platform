import {expect, it, vi} from "vitest";
import {main} from "../../../src/cli.js";
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
