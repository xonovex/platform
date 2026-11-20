import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {checkCrossPackageLinks} from "../../../src/cross-package-links.js";

const REPO = "/repo";
const DOCS = "packages/runtime/plugin/plugin-test/docs";

const document = (path: string, text: string) => ({
  path: join(REPO, path),
  text,
});

const existing = new Set([
  join(
    REPO,
    "packages/runtime/plugin/plugin-x/skills/using-x/references/actors.md",
  ),
  join(REPO, "packages/runtime/plugin/plugin-x/skills/using-x"),
]);

const check = (text: string, path = `${DOCS}/foo.md`) =>
  checkCrossPackageLinks([document(path, text)], REPO, (target) =>
    existing.has(target),
  );

describe("checkCrossPackageLinks", () => {
  it("passes boundary-crossing links to an existing file or directory", () => {
    expect(
      check(
        "See [actors](../../plugin-x/skills/using-x/references/actors.md#cast) and [skill](../../plugin-x/skills/using-x).",
      ),
    ).toEqual([
      {level: "pass", message: "cross-package links: 2/2 link(s) resolve"},
    ]);
  });

  it("fails a boundary-crossing link whose target is missing, naming the source and link", () => {
    expect(
      check(
        "See [a](../../plugin-x/skills/using-x/references/actors.md) and [gone](../../plugin-x/gone.md).",
      ),
    ).toEqual([
      {
        level: "fail",
        message: `cross-package links: broken link in ${DOCS}/foo.md → ../../plugin-x/gone.md`,
      },
    ]);
  });

  it("ignores link-shaped text inside a fenced block or an inline span", () => {
    expect(
      check(
        [
          "```c",
          "  m->listeners[i](../../plugin-x/skills/using-x/references/gone.md);",
          "```",
          "",
          "A markdown hyperlink `[name](../../plugin-x/skills/using-x/nope.md)`.",
        ].join("\n"),
      ),
    ).toEqual([]);
  });

  it("ignores an intra-package link, even a broken one", () => {
    expect(check("See [b](./missing.md).")).toEqual([]);
  });

  it("skips external, placeholder, ellipsis, and anchor forms", () => {
    expect(
      check(
        [
          "[a](https://example.com)",
          "[b](mailto:x@y.z)",
          "[c](../../../skill/<topic>.md)",
          "[d](../../../skill/{topic}.md)",
          "[e](…/pull/PR)",
          "[f](#anchor)",
        ].join("\n"),
      ),
    ).toEqual([]);
  });

  it("treats every link of a file outside the packages as crossing", () => {
    expect(check("See [notes](notes.md).", "docs/guide.md")).toEqual([
      {
        level: "fail",
        message: "cross-package links: broken link in docs/guide.md → notes.md",
      },
    ]);
  });
});
