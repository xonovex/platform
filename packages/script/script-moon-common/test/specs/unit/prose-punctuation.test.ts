import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, beforeEach, describe, expect, it} from "vitest";
import {
  prosePunctuationFindings,
  scanProsePunctuation,
} from "../../../src/prose-punctuation.js";

const EM_DASH = "\u2014";
const ELLIPSIS = "\u2026";
const LEFT_DOUBLE_QUOTE = "\u201C";
const RIGHT_DOUBLE_QUOTE = "\u201D";
const RIGHT_SINGLE_QUOTE = "\u2019";

describe("prosePunctuationFindings", () => {
  it("reports a literal em dash with its line and hint", () => {
    const findings = prosePunctuationFindings(
      "SKILL.md",
      `clean line\n- label ${EM_DASH} detail\n`,
    );
    expect(findings).toEqual([
      {
        excerpt: `- label ${EM_DASH} detail`,
        hint: "use a comma, colon, or full stop",
        label: "em dash",
        line: 2,
        path: "SKILL.md",
      },
    ]);
  });

  it("reports a literal ellipsis character", () => {
    const findings = prosePunctuationFindings(
      "SKILL.md",
      `Use when${ELLIPSIS}\n`,
    );
    expect(findings.map(({label, line}) => [label, line])).toEqual([
      ["ellipsis", 1],
    ]);
  });

  it("reports the JSON-escaped spellings a literal search misses", () => {
    const findings = prosePunctuationFindings(
      "eval-queries.json",
      '{"query": "a \\u2014 b", "rationale": "c \\u2026"}\n',
    );
    expect(findings.map(({label}) => label)).toEqual([
      "escaped em dash",
      "escaped ellipsis",
    ]);
  });

  it("reports typographic quotes with a straight-quote hint", () => {
    const findings = prosePunctuationFindings(
      "SKILL.md",
      `- an axis${RIGHT_SINGLE_QUOTE}s leaf\n- ${LEFT_DOUBLE_QUOTE}quoted${RIGHT_DOUBLE_QUOTE}\n`,
    );
    expect(findings.map(({hint, label, line}) => [label, line, hint])).toEqual([
      ["smart single quote", 1, "use a straight apostrophe"],
      ["smart double quote", 2, "use straight double quotes"],
    ]);
  });

  it("reports the escaped spelling of a typographic quote", () => {
    const findings = prosePunctuationFindings(
      "evals.json",
      String.raw`{"prompt": "an axis\u2019s leaf", "note": "\u201cquoted\u201d"}`,
    );
    expect(findings.map(({label}) => label)).toEqual([
      "escaped smart single quote",
      "escaped smart double quote",
    ]);
  });

  it("accepts plain punctuation", () => {
    expect(
      prosePunctuationFindings(
        "SKILL.md",
        "- label: detail, not a dash\n- Use when...\n" +
          "- an axis's leaf, \"quoted\", 'quoted'\n",
      ),
    ).toEqual([]);
  });

  it("reports every offending line, not just the first", () => {
    const findings = prosePunctuationFindings(
      "a.md",
      `one ${EM_DASH} two\nplain\nthree ${EM_DASH} four\n`,
    );
    expect(findings.map(({line}) => line)).toEqual([1, 3]);
  });
});

describe("scanProsePunctuation", () => {
  let root: string;

  beforeEach(() => {
    root = mkdtempSync(join(tmpdir(), "moon-common-punctuation-"));
  });

  afterEach(() => {
    rmSync(root, {recursive: true, force: true});
  });

  it("walks prose files and returns root-relative paths", () => {
    mkdirSync(join(root, "guide", "references"), {recursive: true});
    writeFileSync(join(root, "package.json"), '{"description": "plain"}\n');
    writeFileSync(join(root, "guide", "SKILL.md"), `a ${EM_DASH} b\n`);
    writeFileSync(
      join(root, "guide", "references", "one.md"),
      `x${ELLIPSIS}\n`,
    );
    expect(
      scanProsePunctuation(root).map(({label, path}) => [path, label]),
    ).toEqual([
      [join("guide", "SKILL.md"), "em dash"],
      [join("guide", "references", "one.md"), "ellipsis"],
    ]);
  });

  it("scans bundled scripts and asset templates", () => {
    mkdirSync(join(root, "scripts"), {recursive: true});
    writeFileSync(
      join(root, "scripts", "validate.py"),
      `# note ${EM_DASH} detail\n`,
    );
    writeFileSync(join(root, "SKILL.md.template"), `{a} ${EM_DASH} {b}\n`);
    expect(scanProsePunctuation(root).map(({path}) => path)).toEqual([
      "SKILL.md.template",
      join("scripts", "validate.py"),
    ]);
  });

  it("skips generated directories and non-prose files", () => {
    mkdirSync(join(root, "dist"), {recursive: true});
    mkdirSync(join(root, "node_modules"), {recursive: true});
    writeFileSync(join(root, "dist", "index.md"), `a ${EM_DASH} b\n`);
    writeFileSync(join(root, "node_modules", "dep.md"), `a ${EM_DASH} b\n`);
    writeFileSync(join(root, "logo.svg"), `<!-- ${EM_DASH} -->\n`);
    expect(scanProsePunctuation(root)).toEqual([]);
  });
});
