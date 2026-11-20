import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {
  checkReferenceFileLinks,
  type ReferenceDocument,
} from "../../../src/reference-file-links.js";

const REFERENCES = "/skill/references";

// Checks documents that sit in REFERENCES, where only the documents themselves
// exist as files.
const check = (documents: readonly ReferenceDocument[]) => {
  const files = new Set(documents.map(({name}) => join(REFERENCES, name)));
  return checkReferenceFileLinks(REFERENCES, documents, (path) =>
    files.has(path),
  );
};

describe("checkReferenceFileLinks", () => {
  it("fails the historical bug: a references/x.md link inside a reference file", () => {
    // The link resolves to references/references/execution.md, not the sibling.
    expect(
      check([
        {
          name: "changelog.md",
          text: "See [references/execution.md](references/execution.md).",
        },
        {name: "execution.md", text: "# Execution\n"},
      ]),
    ).toEqual([
      {
        level: "fail",
        message:
          "reference links: broken link in references/changelog.md → references/execution.md",
      },
    ]);
  });

  it("passes a correct sibling link written as a bare filename", () => {
    expect(
      check([
        {name: "changelog.md", text: "See [execution.md](execution.md)."},
        {name: "execution.md", text: "# Execution\n"},
      ]),
    ).toEqual([
      {level: "pass", message: "reference links: 1/1 link(s) resolve"},
    ]);
  });

  it("resolves a sibling link that carries an in-page fragment", () => {
    expect(
      check([
        {name: "guide.md", text: "See [execution.md](execution.md#setup)."},
        {name: "execution.md", text: "# Execution\n"},
      ]),
    ).toEqual([
      {level: "pass", message: "reference links: 1/1 link(s) resolve"},
    ]);
  });

  it("skips placeholder, ellipsis, external, and anchor forms without a finding", () => {
    const text = [
      "Teaching examples with placeholders:",
      "[a](references/<operation>.md)",
      "[b](references/{topic}.md)",
      "[c](<topic>.md)",
      "[d]({topic}.md)",
      "Authentic changelog form: [#PR](…/pull/PR) [hash](…/commit/hash).",
      "External: [site](https://example.com) [mail](mailto:x@y.z).",
      "In-page: [top](#overview).",
    ].join("\n");

    // Nothing resolvable was present, so no PASS line is emitted either.
    expect(check([{name: "teaching.md", text}])).toEqual([]);
  });

  it("emits no findings when the skill has no reference documents", () => {
    expect(check([])).toEqual([]);
  });
});
