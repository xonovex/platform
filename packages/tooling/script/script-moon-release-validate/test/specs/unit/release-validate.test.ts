import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {describe, expect, it} from "vitest";
import {main} from "../../../src/release-validate.js";

const emptyRepository = {trackedFiles: () => [], ignoredPaths: () => []};

describe("main", () => {
  it("prints each failure and returns 1 when a check fails", () => {
    const output = memoryOutput();
    const opened: string[] = [];

    const code = main("/repo/packages", {
      fs: memoryFileSystem({directories: ["/repo/.moon", "/repo/packages"]}),
      output,
      openRepository: (rootDir) => {
        opened.push(rootDir);
        return emptyRepository;
      },
    });

    expect(code).toBe(1);
    expect(opened).toEqual(["/repo"]);
    expect(output.stdoutLines()).toEqual([]);
    expect(output.stderrLines()[0]).toBe(
      "Release validation failed: 2/2 checks",
    );
    expect(output.stderrLines()[1]).toContain(
      "- .claude-plugin/marketplace.json",
    );
  });

  it("prints a summary and returns 0 when every check passes", () => {
    const output = memoryOutput();
    const marketplace = JSON.stringify({
      metadata: {version: "1.0.0"},
      plugins: [],
    });

    const code = main("/repo", {
      fs: memoryFileSystem({
        directories: ["/repo/.moon", "/repo/packages/runtime/plugin"],
        files: {
          "/repo/.claude-plugin/marketplace.json": marketplace,
          "/repo/.agents/plugins/marketplace.json": marketplace,
          "/repo/package-lock.json": "{}",
          "/repo/README.md": "# Repository\n",
          "/repo/.github/workflows/release.yml": `on:
  workflow_dispatch:
  pull_request:
jobs:
  release:
    steps:
      - name: Publish
        if: github.event_name == 'pull_request'
        run: publish
      - name: Publish (dry run)
        if: github.event_name == 'workflow_dispatch'
        run: dry-run
      - uses: example/report@sha
`,
          "/repo/.moon/tasks/tag-typescript.yml": "tasks: {}\n",
          "/repo/packages/AGENTS.md": "# Packages\n",
          "/repo/packages/CLAUDE.md":
            "See @AGENTS.md for complete documentation.\n",
          "/repo/packages/runtime/AGENTS.md": "# Runtime\n",
          "/repo/packages/runtime/CLAUDE.md":
            "See @AGENTS.md for complete documentation.\n",
          "/repo/packages/runtime/plugin/AGENTS.md": "# Plugins\n",
          "/repo/packages/runtime/plugin/CLAUDE.md":
            "See @AGENTS.md for complete documentation.\n",
          "/repo/packages/tool/moon.yml": "language: javascript\n",
          "/repo/packages/tool/package.json": JSON.stringify({name: "tool"}),
        },
      }),
      output,
      openRepository: () => emptyRepository,
    });

    expect(code).toBe(0);
    expect(output.stderrLines()).toEqual([]);
    expect(output.stdoutLines()).toEqual([
      expect.stringMatching(
        /^Release validation passed: \d+ checks, 0 lockstep packages$/u,
      ),
    ]);
  });
});
