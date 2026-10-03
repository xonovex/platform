import {join} from "node:path";
import {
  memoryFileSystem,
  type MemoryTree,
} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  checkReferenceFileLinks,
  type LinkReport,
} from "../../../src/reference-file-links.js";

const SKILL_DIR = "/skill";

// A LinkReport that records what the check reports, so a test can assert on it.
const makeSink = (): LinkReport & {fails: string[]; passes: string[]} => {
  const fails: string[] = [];
  const passes: string[] = [];
  return {
    fails,
    passes,
    addFail: (message) => {
      fails.push(message);
    },
    addPass: (message) => {
      passes.push(message);
    },
  };
};

// A skill whose references/ holds the given files, keyed by filename.
const skillWith = (files: Readonly<Record<string, string>>) =>
  memoryFileSystem({
    files: Object.fromEntries(
      Object.entries(files).map(([name, content]) => [
        join(SKILL_DIR, "references", name),
        content,
      ]),
    ),
  } satisfies MemoryTree);

describe("checkReferenceFileLinks", () => {
  it("fails the historical bug: a references/x.md link inside a reference file", () => {
    // The link resolves to references/references/execution.md, not the sibling.
    const fs = skillWith({
      "changelog.md": "See [references/execution.md](references/execution.md).",
      "execution.md": "# Execution\n",
    });
    const report = makeSink();
    checkReferenceFileLinks(SKILL_DIR, report, fs);
    expect(report.fails).toHaveLength(1);
    expect(report.fails[0]).toContain("references/execution.md");
    expect(report.passes).toEqual([]);
  });

  it("passes a correct sibling link written as a bare filename", () => {
    const fs = skillWith({
      "changelog.md": "See [execution.md](execution.md).",
      "execution.md": "# Execution\n",
    });
    const report = makeSink();
    checkReferenceFileLinks(SKILL_DIR, report, fs);
    expect(report.fails).toEqual([]);
    expect(report.passes).toEqual(["reference links: 1/1 link(s) resolve"]);
  });

  it("resolves a sibling link that carries an in-page fragment", () => {
    const fs = skillWith({
      "guide.md": "See [execution.md](execution.md#setup).",
      "execution.md": "# Execution\n",
    });
    const report = makeSink();
    checkReferenceFileLinks(SKILL_DIR, report, fs);
    expect(report.fails).toEqual([]);
    expect(report.passes).toEqual(["reference links: 1/1 link(s) resolve"]);
  });

  it("skips placeholder, ellipsis, external, and anchor forms without failing", () => {
    const fs = skillWith({
      "teaching.md": [
        "Teaching examples with placeholders:",
        "[a](references/<operation>.md)",
        "[b](references/{topic}.md)",
        "[c](<topic>.md)",
        "[d]({topic}.md)",
        "Authentic changelog form: [#PR](…/pull/PR) [hash](…/commit/hash).",
        "External: [site](https://example.com) [mail](mailto:x@y.z).",
        "In-page: [top](#overview).",
      ].join("\n"),
    });
    const report = makeSink();
    checkReferenceFileLinks(SKILL_DIR, report, fs);
    expect(report.fails).toEqual([]);
    // Nothing resolvable was present, so no PASS line is emitted either.
    expect(report.passes).toEqual([]);
  });

  it("emits no findings when the skill has no references directory", () => {
    const report = makeSink();
    checkReferenceFileLinks(
      SKILL_DIR,
      report,
      memoryFileSystem({directories: [SKILL_DIR]}),
    );
    expect(report.fails).toEqual([]);
    expect(report.passes).toEqual([]);
  });
});
