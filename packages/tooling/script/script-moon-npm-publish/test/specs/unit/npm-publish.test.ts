import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {expect, it, vi} from "vitest";
import {main} from "../../../src/npm-publish.js";

const ROOT = "/workspace/package";

const setup = (published = false) => {
  const publish = vi.fn();
  const output = memoryOutput();
  const fs = memoryFileSystem({
    files: {[`${ROOT}/package.json`]: '{"name":"example","version":"1.2.3"}'},
  });
  return {
    publish,
    output,
    fs,
    dependencies: {fs, output, isPublished: () => published, publish},
  };
};

it("passes the dry-run option through the executable entrypoint", () => {
  const {dependencies, publish} = setup();

  expect(main(["--dry-run"], ROOT, dependencies)).toBe(0);
  expect(publish).toHaveBeenCalledWith(true);
});

it("prints the help text and publishes nothing on --help", () => {
  const {dependencies, publish, output} = setup();

  expect(main(["--help"], ROOT, dependencies)).toBe(0);
  expect(output.stdoutLines()[0]).toContain("moon-npm-publish");
  expect(publish).not.toHaveBeenCalled();
});

it("reports a version that npm already has and publishes nothing", () => {
  const {dependencies, publish, output} = setup(true);

  expect(main([], ROOT, dependencies)).toBe(0);
  expect(output.stdoutLines()).toEqual([
    expect.stringContaining("Skipping example@1.2.3"),
  ]);
  expect(publish).not.toHaveBeenCalled();
});

it("injects the platform fields while it publishes and restores the manifest", () => {
  const {dependencies, fs, output} = setup();
  fs.writeFile(
    `${ROOT}/platform.json`,
    JSON.stringify({os: ["linux"], cpu: ["x64"]}),
  );
  let during = "";
  const publish = vi.fn(() => {
    during = fs.readText(`${ROOT}/package.json`);
  });

  expect(main([], ROOT, {...dependencies, publish})).toBe(0);
  expect(JSON.parse(during)).toMatchObject({os: ["linux"], cpu: ["x64"]});
  expect(fs.readText(`${ROOT}/package.json`)).toBe(
    '{"name":"example","version":"1.2.3"}',
  );
  expect(output.stdoutLines()).toEqual([
    expect.stringContaining("Injected platform fields for example"),
    expect.stringContaining("Restored original package.json for example"),
  ]);
});
