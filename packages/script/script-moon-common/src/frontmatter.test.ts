import {mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {parseFrontmatterName} from "./frontmatter.js";

describe("parseFrontmatterName", () => {
  let root: string;

  const write = (contents: string): string => {
    const file = join(root, "SKILL.md");
    writeFileSync(file, contents);
    return file;
  };

  beforeEach(() => {
    root = mkdtempSync(join(tmpdir(), "moon-common-frontmatter-"));
  });

  afterEach(() => {
    rmSync(root, {recursive: true, force: true});
  });

  it("reads the declared name", () => {
    const file = write(
      "---\nname: git-guide\ndescription: Use git\n---\n\nBody",
    );
    expect(parseFrontmatterName(file)).toBe("git-guide");
  });

  it("strips surrounding quotes", () => {
    expect(parseFrontmatterName(write('---\nname: "git-guide"\n---\n'))).toBe(
      "git-guide",
    );
    expect(parseFrontmatterName(write("---\nname: 'git-guide'\n---\n"))).toBe(
      "git-guide",
    );
  });

  it("trims surrounding whitespace", () => {
    expect(parseFrontmatterName(write("---\nname:   git-guide  \n---\n"))).toBe(
      "git-guide",
    );
  });

  it("reads frontmatter written with CRLF line endings", () => {
    expect(
      parseFrontmatterName(write("---\r\nname: git-guide\r\n---\r\n")),
    ).toBe("git-guide");
  });

  it("returns the first name when the key repeats", () => {
    expect(
      parseFrontmatterName(write("---\nname: first\nname: second\n---\n")),
    ).toBe("first");
  });

  it("ignores a name declared outside the frontmatter block", () => {
    expect(
      parseFrontmatterName(write("---\ndescription: d\n---\nname: body\n")),
    ).toBeUndefined();
  });

  it("returns undefined when the document has no frontmatter block", () => {
    expect(parseFrontmatterName(write("name: git-guide\n"))).toBeUndefined();
  });

  it("returns undefined when the frontmatter declares no name", () => {
    expect(
      parseFrontmatterName(write("---\ndescription: Use git\n---\n")),
    ).toBeUndefined();
  });
});
