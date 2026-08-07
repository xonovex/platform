import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it} from "vitest";
import {
  checkReferenceFileLinks,
  type LinkReport,
} from "../../../src/reference-file-links.js";

const created: string[] = [];

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

// Build a throwaway skill dir whose references/ holds the given files, keyed by
// filename. Returns the skill dir to hand to checkReferenceFileLinks.
const makeSkill = (files: Record<string, string>): string => {
  const skillDir = mkdtempSync(join(tmpdir(), "skill-validate-"));
  created.push(skillDir);
  const refsDir = join(skillDir, "references");
  mkdirSync(refsDir);
  for (const [name, content] of Object.entries(files)) {
    writeFileSync(join(refsDir, name), content);
  }
  return skillDir;
};

afterEach(() => {
  while (created.length > 0) {
    const dir = created.pop();
    if (dir !== undefined) rmSync(dir, {recursive: true, force: true});
  }
});

describe("checkReferenceFileLinks", () => {
  it("fails the historical bug: a references/x.md link inside a reference file", () => {
    // The link resolves to references/references/execution.md, not the sibling.
    const skill = makeSkill({
      "changelog.md": "See [references/execution.md](references/execution.md).",
      "execution.md": "# Execution\n",
    });
    const report = makeSink();
    checkReferenceFileLinks(skill, report);
    expect(report.fails).toHaveLength(1);
    expect(report.fails[0]).toContain("references/execution.md");
    expect(report.passes).toEqual([]);
  });

  it("passes a correct sibling link written as a bare filename", () => {
    const skill = makeSkill({
      "changelog.md": "See [execution.md](execution.md).",
      "execution.md": "# Execution\n",
    });
    const report = makeSink();
    checkReferenceFileLinks(skill, report);
    expect(report.fails).toEqual([]);
    expect(report.passes).toEqual(["reference links: 1/1 link(s) resolve"]);
  });

  it("resolves a sibling link that carries an in-page fragment", () => {
    const skill = makeSkill({
      "guide.md": "See [execution.md](execution.md#setup).",
      "execution.md": "# Execution\n",
    });
    const report = makeSink();
    checkReferenceFileLinks(skill, report);
    expect(report.fails).toEqual([]);
    expect(report.passes).toEqual(["reference links: 1/1 link(s) resolve"]);
  });

  it("skips placeholder, ellipsis, external, and anchor forms without failing", () => {
    const skill = makeSkill({
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
    checkReferenceFileLinks(skill, report);
    expect(report.fails).toEqual([]);
    // Nothing resolvable was present, so no PASS line is emitted either.
    expect(report.passes).toEqual([]);
  });

  it("emits no findings when the skill has no references directory", () => {
    const skillDir = mkdtempSync(join(tmpdir(), "skill-validate-"));
    created.push(skillDir);
    const report = makeSink();
    checkReferenceFileLinks(skillDir, report);
    expect(report.fails).toEqual([]);
    expect(report.passes).toEqual([]);
  });
});
