import {execFileSync} from "node:child_process";
import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {processExecutableResolver} from "@xonovex/script-moon-common/adapters/executable-node";
import {nodeFileSystem} from "@xonovex/script-moon-common/adapters/file-system-node";
import {nodeGitRepository} from "@xonovex/script-moon-common/adapters/git-node";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {main} from "../../../src/version-detect.js";

// Drives main through the real git and disk adapters against a fresh repository.
// The project list stands in for moon, which needs a configured workspace.
const resolveExecutable = processExecutableResolver(nodeFileSystem);

describe("version detect", () => {
  let root: string;

  const git = (...args: readonly string[]): void => {
    execFileSync(resolveExecutable("git"), [...args], {cwd: root});
  };

  const writeManifest = (version: string): void => {
    writeFileSync(
      join(root, "packages", "example", "package.json"),
      JSON.stringify({name: "@xonovex/example", version}),
    );
  };

  beforeEach(() => {
    root = mkdtempSync(join(tmpdir(), "version-detect-"));
    mkdirSync(join(root, ".moon"));
    mkdirSync(join(root, "packages", "example"), {recursive: true});
    git("init", "--quiet");
    git("config", "user.name", "Test Author");
    git("config", "user.email", "test@example.com");
    git("config", "commit.gpgsign", "false");
    writeManifest("1.0.0");
    git("add", "--all");
    git("commit", "--quiet", "--message", "feat: initial");
  });

  afterEach(() => {
    rmSync(root, {recursive: true, force: true});
  });

  it("prints the projects whose version differs from the ref", () => {
    writeManifest("1.1.0");
    const output = memoryOutput();

    const code = main(["--ref", "HEAD"], join(root, "packages"), {
      openRepository: nodeGitRepository(resolveExecutable),
      listProjects: () => [{id: "example", source: "packages/example"}],
      fs: nodeFileSystem,
      output,
    });

    expect(code).toBe(0);
    expect(output.stdoutLines()).toEqual([JSON.stringify(["example"])]);
  });

  it("fails for a ref that names no commit", () => {
    expect(() =>
      main(["--ref", "not-a-ref"], root, {
        openRepository: nodeGitRepository(resolveExecutable),
        listProjects: () => [],
        fs: nodeFileSystem,
        output: memoryOutput(),
      }),
    ).toThrow();
  });
});
