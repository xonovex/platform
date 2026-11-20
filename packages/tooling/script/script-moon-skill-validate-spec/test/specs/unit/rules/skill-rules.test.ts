import {describe, expect, it} from "vitest";
import {checkBody} from "../../../../src/rules/body.js";
import {checkFrontmatter} from "../../../../src/rules/frontmatter.js";
import {checkHarnessNeutrality} from "../../../../src/rules/harness-neutrality.js";
import {checkLoaderQuoting} from "../../../../src/rules/loader-quoting.js";
import {
  checkReferences,
  checkReferenceTocs,
} from "../../../../src/rules/references.js";

const DESCRIPTION =
  "Use when testing. Triggers on fixtures, even when the user doesn't say test.";

const isFile = (path: string): boolean => path === "/s/references/one.md";

// A reference file just over the length that calls for a table of contents.
const long = (head: string): string => `${head}\n${"line\n".repeat(201)}`;

const messagesAt = (
  findings: readonly {readonly level: string; readonly message: string}[],
  level: string,
): readonly string[] =>
  findings
    .filter((finding) => finding.level === level)
    .map(({message}) => message);

describe("checkBody", () => {
  it("passes a short body with a heading, tagged code and gotchas", () => {
    const body = "# Guide\n\n```ts\nx\n```\n\n## Gotchas\n";

    expect(messagesAt(checkBody(body), "pass")).toEqual([
      "body: 7 lines (under 500)",
      "body: has at least one heading",
      "body: all code block(s) have language markers (1 language-tagged opening fence(s) found)",
      "content: '## Gotchas' section present",
    ]);
  });

  it("warns on a long body and on untagged code", () => {
    const body = `text\n\`\`\`\nx\n\`\`\`\n${"\n".repeat(400)}`;

    expect(messagesAt(checkBody(body), "warn")).toEqual([
      "body: 404 lines (>80% of 500-line target)",
      "body: 2 code block(s) appear to lack a language marker",
      "content: no '## Gotchas' section — non-obvious env-specific facts have no home",
    ]);
    expect(messagesAt(checkBody(body), "fail")).toEqual([
      "body: no headings found",
    ]);
  });

  it("fails a body over the line target", () => {
    expect(messagesAt(checkBody("\n".repeat(501)), "fail")).toContain(
      "body: 501 lines (>500 spec target)",
    );
  });
});

describe("checkFrontmatter", () => {
  it("passes a complete frontmatter that matches its directory", () => {
    const findings = checkFrontmatter(
      {
        name: "using-tests",
        description: DESCRIPTION,
        compatibility: "node",
        license: "MIT",
        metadata: {owner: "team"},
        "allowed-tools": "Bash(git:*) Read",
      },
      "using-tests",
    );

    expect(messagesAt(findings, "fail")).toEqual([]);
    expect(messagesAt(findings, "warn")).toEqual([]);
    expect(messagesAt(findings, "pass")).toContain(
      "compatibility: 4 chars (under 500)",
    );
  });

  it("reports wrong field types and unknown fields", () => {
    const findings = checkFrontmatter(
      {
        name: 7,
        description: ["list"],
        compatibility: 1.5,
        license: true,
        metadata: ["list"],
        "allowed-tools": {},
        extra: null,
      },
      "using-tests",
    );

    expect(messagesAt(findings, "fail")).toEqual([
      "frontmatter: name is int, expected string",
      "frontmatter: description is list, expected string",
      "frontmatter: compatibility is float, expected string",
    ]);
    expect(messagesAt(findings, "warn")).toEqual([
      "frontmatter: license is bool, expected string",
      "frontmatter: metadata should be a string→string map",
      "frontmatter: allowed-tools is dict, expected a space-separated string",
      "frontmatter: unknown top-level field(s): extra",
    ]);
  });

  it("reports missing fields and an overlong compatibility", () => {
    const findings = checkFrontmatter(
      {compatibility: "x".repeat(501), "allowed-tools": "Bash;rm"},
      "using-tests",
    );

    expect(messagesAt(findings, "fail")).toEqual([
      "frontmatter: missing 'name'",
      "frontmatter: missing 'description'",
      "frontmatter: compatibility is 501 chars (>500)",
    ]);
    expect(messagesAt(findings, "warn")).toEqual([
      "frontmatter: allowed-tools has unexpected characters (expected space-separated tokens like 'Bash(git:*) Read')",
    ]);
  });

  it.each([
    [
      "x".repeat(65),
      `frontmatter: name '${"x".repeat(65)}' is 65 chars (>64 spec limit)`,
    ],
    [
      "using-tests",
      "frontmatter: name 'using-tests' != parent directory 'other'",
    ],
    [
      "using-claude",
      "frontmatter: name 'using-claude' uses a reserved word (anthropic/claude)",
    ],
  ])("fails the name %s", (name, message) => {
    expect(
      messagesAt(
        checkFrontmatter({name, description: DESCRIPTION}, "other"),
        "fail",
      ),
    ).toContain(message);
  });

  it("allows the vendor adapter name and flags markup in a description", () => {
    const findings = checkFrontmatter(
      {name: "configuring-claude-code", description: "Do <b>things</b>."},
      "configuring-claude-code",
    );

    expect(messagesAt(findings, "fail")).toEqual([]);
    expect(messagesAt(findings, "warn")).toEqual([
      "description: does not start with 'Use when…' / 'Use this skill when…' — agents may not trigger reliably",
      "description: missing 'Triggers on…' — agents lose the trigger keyword list",
      "description: no 'even when the user doesn't say…' clause — may miss implicit triggers",
      "description: contains angle-bracket markup (`<…>`) — fine if it's a component/generic reference, otherwise remove the XML tag",
    ]);
  });
});

describe("checkLoaderQuoting", () => {
  it("passes a double-quoted description", () => {
    expect(checkLoaderQuoting('name: x\ndescription: "Use when."')).toEqual([
      {
        level: "pass",
        message: "frontmatter: description uses a double-quoted scalar",
      },
      {
        level: "pass",
        message:
          "frontmatter: quoting is loader-safe (no '' escapes or inner delimiter quotes)",
      },
    ]);
  });

  it("fails escaped and inner quotes the loader cuts short", () => {
    const findings = checkLoaderQuoting(
      "description: 'it''s'\nlicense: \"a \" b\"\nnote: 'open",
    );

    expect(findings.map(({message}) => message.slice(0, 60))).toEqual([
      "frontmatter: 'description' must use one double-quoted scalar",
      "frontmatter: 'description' uses single-quote '' escaping, wh",
      "frontmatter: 'license' has an inner \" inside a \"-quoted valu",
    ]);
  });
});

describe("checkHarnessNeutrality", () => {
  it("fails each line that names a vendor harness", () => {
    expect(checkHarnessNeutrality("Read CLAUDE.md\nok\nUse TodoWrite")).toEqual(
      [
        {
          level: "fail",
          message:
            "harness-neutrality: Claude Code tool name — line 3: Use TodoWrite",
        },
        {
          level: "fail",
          message:
            "harness-neutrality: Vendor-specific filename (use AGENTS.md) — line 1: Read CLAUDE.md",
        },
      ],
    );
  });
});

describe("checkReferences", () => {
  it("passes resolving links that carry load triggers", () => {
    const body =
      "## Progressive Disclosure\n\n- Read references/one.md - Load when testing\n";

    expect(messagesAt(checkReferences(body, "/s", isFile), "pass")).toEqual([
      "references: 1/1 link(s) resolve",
      "references: no '@references/' prefix",
      "references: all 1 link(s) in Progressive Disclosure carry a 'Load when…' trigger",
    ]);
  });

  it("reports broken, nested, misnamed and eager links", () => {
    const body = [
      "## Progressive Disclosure",
      "",
      "- references/deep/Two_Words.md",
      "- @references/one.md",
      "- references/<topic>.md",
    ].join("\n");
    const findings = checkReferences(body, "/s", isFile);

    expect(messagesAt(findings, "fail")).toEqual([
      "references: broken link → references/deep/Two_Words.md",
      "references: '@references/' prefix found — strips progressive disclosure (use plain 'references/')",
    ]);
    expect(messagesAt(findings, "warn")).toEqual([
      "references: 'references/deep/Two_Words.md' is nested deeper than one level",
      "references: 'references/deep/Two_Words.md' filename is not kebab-case",
      "references: 3 of 3 link(s) in Progressive Disclosure lack a 'Load when…' trigger",
    ]);
  });

  it("reports nothing for a body without reference links", () => {
    expect(checkReferences("# Guide", "/s", isFile)).toEqual([]);
  });
});

describe("checkReferenceTocs", () => {
  it("passes long files that open with contents", () => {
    expect(
      checkReferenceTocs([
        {name: "a.md", text: long("## Contents")},
        {name: "b.md", text: "short"},
      ]),
    ).toEqual([
      {
        level: "pass",
        message:
          "references: all 1 reference file(s) >200 lines open with a '## Contents' table of contents",
      },
    ]);
  });

  it("warns on a long file without contents", () => {
    expect(checkReferenceTocs([{name: "a.md", text: long("# A")}])).toEqual([
      {
        level: "warn",
        message:
          "references: file(s) >200 lines lack a '## Contents' table of contents — a.md (202 lines)",
      },
    ]);
  });
});
