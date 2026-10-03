import {describe, expect, it} from "vitest";
import {memoryFileSystem} from "../../../src/file-system-memory.js";
import {parseFrontmatterName} from "../../../src/frontmatter.js";

const nameIn = (contents: string): string | undefined =>
  parseFrontmatterName(
    "/skill/SKILL.md",
    memoryFileSystem({files: {"/skill/SKILL.md": contents}}),
  );

describe("parseFrontmatterName", () => {
  it("reads the declared name", () => {
    expect(
      nameIn("---\nname: git-guide\ndescription: Use git\n---\n\nBody"),
    ).toBe("git-guide");
  });

  it("strips surrounding quotes", () => {
    expect(nameIn('---\nname: "git-guide"\n---\n')).toBe("git-guide");
    expect(nameIn("---\nname: 'git-guide'\n---\n")).toBe("git-guide");
  });

  it("trims surrounding whitespace", () => {
    expect(nameIn("---\nname:   git-guide  \n---\n")).toBe("git-guide");
  });

  it("reads frontmatter written with CRLF line endings", () => {
    expect(nameIn("---\r\nname: git-guide\r\n---\r\n")).toBe("git-guide");
  });

  it("returns the first name when the key repeats", () => {
    expect(nameIn("---\nname: first\nname: second\n---\n")).toBe("first");
  });

  it("ignores a name declared outside the frontmatter block", () => {
    expect(nameIn("---\ndescription: d\n---\nname: body\n")).toBeUndefined();
  });

  it("returns undefined when the document has no frontmatter block", () => {
    expect(nameIn("name: git-guide\n")).toBeUndefined();
  });

  it("returns undefined when the frontmatter declares no name", () => {
    expect(nameIn("---\ndescription: Use git\n---\n")).toBeUndefined();
  });
});
