#!/usr/bin/env node
import {readdirSync, readFileSync, statSync} from "node:fs";
import {basename, dirname, join, resolve} from "node:path";
import {parse as parseYaml} from "yaml";

const NAME_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const FRONTMATTER_RE = /^---\n([\s\S]*?)\n---\n([\s\S]*)$/;
const QUOTED_SCALAR_RE = /^[ \t]*[a-z_][\w-]*:[ \t]*(.+?)[ \t]*$/i;
const REF_LINK_RE = /references\/([\w./<>{}-]+\.md)/g;
const PROGRESSIVE_DISCLOSURE_RE =
  /^## *Progressive Disclosure\s*\n([\s\S]*?)(?=^## |(?![\s\S]))/im;
const HEADING_RE = /^#{1,6} /m;
const CODE_FENCE_OPEN_RE = /^```([\w+-]+)?[^\S\n]*$/gm;
const LOAD_WHEN_RE = /load when/i;

const HARNESS_PATTERNS: readonly [RegExp, string][] = [
  // Proprietary tool / function / mode names
  [/\bEnterPlanMode\b/, "Claude Code tool name"],
  [/\bExitPlanMode\b/, "Claude Code tool name"],
  [/\bAskUserQuestion\b/, "Claude Code tool name"],
  [/\bTodoWrite\b/, "Claude Code tool name"],
  [/\bsubagent_type\b/, "Claude Code parameter"],
  [/\bmodel=haiku\b/, "Vendor model id"],
  [/\bmodel=sonnet\b/, "Vendor model id"],
  [/\bmodel=opus\b/, "Vendor model id"],
  // Vendor-namespaced paths
  [/\.claude\/skills\//, "Vendor-namespaced path"],
  [/\.claude\/commands\//, "Vendor-namespaced path"],
  // Vendor-prefixed frontmatter keys
  [/\bclaude:/, "Vendor-prefixed frontmatter key"],
  // Vendor-specific instruction filenames
  [/\bCLAUDE\.md\b/, "Vendor-specific filename (use AGENTS.md)"],
];

type Frontmatter = Readonly<Record<string, unknown>>;

class Report {
  readonly passes: string[] = [];
  readonly warnings: string[] = [];
  readonly errors: string[] = [];

  addPass(msg: string): void {
    this.passes.push(msg);
  }

  addWarn(msg: string): void {
    this.warnings.push(msg);
  }

  addFail(msg: string): void {
    this.errors.push(msg);
  }
}

// Match Python's str.splitlines() for the line boundaries that occur in
// SKILL.md text: \r\n, \r, and \n (without keeping the terminators). Unlike a
// plain split, Python does NOT emit a trailing empty element for a terminal
// line break — e.g. "a\nb\n".splitlines() == ["a", "b"].
const splitLines = (text: string): string[] => {
  if (text.length === 0) return [];
  const parts = text.split(/\r\n|\r|\n/);
  if (parts.length > 0 && parts.at(-1) === "") parts.pop();
  return parts;
};

// Python type(x).__name__ for the YAML scalar/collection types we surface.
const pyTypeName = (value: unknown): string => {
  if (value === null) return "NoneType";
  if (typeof value === "boolean") return "bool";
  if (typeof value === "number")
    return Number.isInteger(value) ? "int" : "float";
  if (typeof value === "string") return "str";
  if (Array.isArray(value)) return "list";
  if (typeof value === "object") return "dict";
  return typeof value;
};

const isFile = (path: string): boolean => {
  try {
    return statSync(path).isFile();
  } catch {
    return false;
  }
};

const isDir = (path: string): boolean => {
  try {
    return statSync(path).isDirectory();
  } catch {
    return false;
  }
};

const resolveTarget = (arg: string): {skill: string; skillDir: string} => {
  const target = resolve(arg);
  let skill: string;
  if (isDir(target)) {
    if (isFile(join(target, "SKILL.md"))) {
      skill = join(target, "SKILL.md");
    } else {
      // Skill-package layout: the SKILL.md lives in a single guide subdir
      // (e.g. skill-c99/c99-guide/SKILL.md), so descend one level.
      const nested = readdirSync(target)
        .map((entry) => join(target, entry, "SKILL.md"))
        .filter((p) => isFile(p));
      if (nested.length > 1) {
        process.stderr.write(
          `Error: multiple SKILL.md found under ${target}; pass one explicitly\n`,
        );
        process.exit(2);
      }
      skill = nested[0] ?? join(target, "SKILL.md");
    }
  } else if (isFile(target)) {
    skill = target;
  } else {
    process.stderr.write(`Error: target not found: ${target}\n`);
    process.exit(2);
  }
  if (!isFile(skill)) {
    process.stderr.write(`Error: SKILL.md not found at ${skill}\n`);
    process.exit(2);
  }
  return {skill, skillDir: dirname(skill)};
};

const splitFrontmatter = (
  content: string,
): {fm: Frontmatter; body: string; fmRaw: string} => {
  const match = FRONTMATTER_RE.exec(content);
  if (!match) {
    return {fm: {}, body: content, fmRaw: ""};
  }
  const fmRaw = match[1] ?? "";
  const body = match[2] ?? "";
  let fm: unknown;
  try {
    fm = parseYaml(fmRaw) ?? {};
  } catch (error) {
    throw new Error(`Invalid YAML frontmatter: ${String(error)}`);
  }
  if (fm === null || typeof fm !== "object" || Array.isArray(fm)) {
    throw new Error("Frontmatter is not a mapping");
  }
  return {fm: fm as Frontmatter, body, fmRaw};
};

const checkLoaderQuoting = (fmRaw: string, report: Report): void => {
  let flagged = false;
  for (const rawLine of splitLines(fmRaw)) {
    const m = QUOTED_SCALAR_RE.exec(rawLine);
    if (!m) continue;
    const rest = m[1] ?? "";
    if (rest.length < 2 || (!rest.startsWith("'") && !rest.startsWith('"')))
      continue;
    const quote = rest.startsWith("'") ? "'" : '"';
    if (!rest.endsWith(quote)) continue; // multiline / unterminated — cannot judge naively
    const key = (rawLine.split(":", 1)[0] ?? "").trim();
    const count = rest.split(quote).length - 1;
    if (quote === "'" && rest.includes("''")) {
      report.addFail(
        `frontmatter: '${key}' uses single-quote '' escaping, which the skill ` +
          "loader's parser does not support (it reads the tail as an unknown " +
          "attribute). Use a double-quoted scalar with literal apostrophes, e.g. " +
          "\"… doesn't say 'TDD'.\"",
      );
      flagged = true;
    } else if (count !== 2) {
      report.addFail(
        `frontmatter: '${key}' has an inner ${quote} inside a ${quote}-quoted value; ` +
          "the skill loader stops at the first inner quote. Use a double-quoted scalar " +
          "with single quotes for inner phrases (apostrophes stay literal).",
      );
      flagged = true;
    }
  }
  if (!flagged) {
    report.addPass(
      "frontmatter: quoting is loader-safe (no '' escapes or inner delimiter quotes)",
    );
  }
};

const checkFrontmatter = (
  fm: Frontmatter,
  parentName: string,
  report: Report,
): void => {
  // name
  const name = fm.name;
  if (!name) {
    report.addFail("frontmatter: missing 'name'");
  } else if (typeof name !== "string") {
    report.addFail(`frontmatter: name is ${pyTypeName(name)}, expected string`);
  } else if (name.length > 64) {
    report.addFail(
      `frontmatter: name '${name}' is ${String(name.length)} chars (>64 spec limit)`,
    );
  } else if (!NAME_RE.test(name)) {
    report.addFail(
      `frontmatter: name '${name}' is not kebab-case ` +
        `(regex ^[a-z0-9]+(-[a-z0-9]+)*$)`,
    );
  } else if (name === parentName) {
    report.addPass(
      `name: '${name}' (${String(name.length)} chars, kebab-case, matches parent dir)`,
    );
  } else {
    report.addFail(
      `frontmatter: name '${name}' != parent directory '${parentName}'`,
    );
  }

  // name: reserved words / XML tags (platform spec)
  if (typeof name === "string") {
    if (/anthropic|claude/i.test(name)) {
      report.addFail(
        `frontmatter: name '${name}' uses a reserved word (anthropic/claude)`,
      );
    }
    if (/<[^>]+>/.test(name)) {
      report.addFail(`frontmatter: name '${name}' contains an XML tag`);
    }
  }

  // description
  const desc = fm.description;
  if (!desc) {
    report.addFail("frontmatter: missing 'description'");
  } else if (typeof desc === "string") {
    const descLen = desc.length;
    if (descLen > 1024) {
      report.addFail(
        `description: ${String(descLen)} chars (>1024 spec limit)`,
      );
    } else if (descLen > 820) {
      report.addWarn(
        `description: ${String(descLen)} chars (>80% of 1024 limit; consider trimming)`,
      );
      report.addPass(`description present (${String(descLen)} chars)`);
    } else {
      report.addPass(`description: ${String(descLen)} chars (under 1024)`);
    }

    if (/^\s*use (?:this skill )?when/i.test(desc)) {
      report.addPass("description: imperative phrasing ('Use when…')");
    } else {
      report.addWarn(
        "description: does not start with 'Use when…' / 'Use this skill when…' — " +
          "agents may not trigger reliably",
      );
    }

    if (/triggers? on/i.test(desc)) {
      report.addPass("description: includes trigger contexts ('Triggers on…')");
    } else {
      report.addWarn(
        "description: missing 'Triggers on…' — agents lose the trigger keyword list",
      );
    }

    if (/even when the user doesn'?t/i.test(desc)) {
      report.addPass(
        "description: includes non-obvious-trigger clause " +
          "('even when the user doesn't say…')",
      );
    } else {
      report.addWarn(
        "description: no 'even when the user doesn't say…' clause — " +
          "may miss implicit triggers",
      );
    }
  } else {
    report.addFail(
      `frontmatter: description is ${pyTypeName(desc)}, expected string`,
    );
  }

  // description: XML tags (platform spec) — warn, since descriptions legitimately
  // reference component/generic syntax (e.g. `<motion.*>`, `<T>`) that regex can't
  // distinguish from a real XML tag.
  if (typeof desc === "string" && /<[^>]+>/.test(desc)) {
    report.addWarn(
      "description: contains angle-bracket markup (`<…>`) — fine if it's a " +
        "component/generic reference, otherwise remove the XML tag",
    );
  }

  // compatibility (optional)
  const compat = fm.compatibility;
  if (compat !== undefined && compat !== null) {
    if (typeof compat !== "string") {
      report.addFail(
        `frontmatter: compatibility is ${pyTypeName(compat)}, expected string`,
      );
    } else if (compat.length > 500) {
      report.addFail(
        `frontmatter: compatibility is ${String(compat.length)} chars (>500)`,
      );
    } else {
      report.addPass(
        `compatibility: ${String(compat.length)} chars (under 500)`,
      );
    }
  }

  // license (optional)
  const lic = fm.license;
  if (lic !== undefined && lic !== null && typeof lic !== "string") {
    report.addWarn(
      `frontmatter: license is ${pyTypeName(lic)}, expected string`,
    );
  }

  // metadata (optional): string -> string map
  const meta = fm.metadata;
  if (meta !== undefined && meta !== null) {
    const okMap =
      typeof meta === "object" &&
      !Array.isArray(meta) &&
      Object.entries(meta as Record<string, unknown>).every(
        ([k, v]) => typeof k === "string" && typeof v === "string",
      );
    if (!okMap) {
      report.addWarn("frontmatter: metadata should be a string→string map");
    }
  }

  // allowed-tools (optional, experimental): space-separated token string
  const allowed = fm["allowed-tools"];
  if (allowed !== undefined && allowed !== null) {
    if (typeof allowed !== "string") {
      report.addWarn(
        `frontmatter: allowed-tools is ${pyTypeName(allowed)}, ` +
          "expected a space-separated string",
      );
    } else if (!/^[\w():*,.\- ]+$/.test(allowed)) {
      report.addWarn(
        "frontmatter: allowed-tools has unexpected characters " +
          "(expected space-separated tokens like 'Bash(git:*) Read')",
      );
    }
  }

  // unknown top-level fields
  const known = new Set([
    "name",
    "description",
    "license",
    "compatibility",
    "metadata",
    "allowed-tools",
  ]);
  const unknown = Object.keys(fm)
    .filter((k) => !known.has(k))
    .toSorted();
  if (unknown.length > 0) {
    report.addWarn(
      "frontmatter: unknown top-level field(s): " + unknown.join(", "),
    );
  }
};

const checkBody = (body: string, report: Report): void => {
  const lineCount = (body.match(/\n/g) ?? []).length;
  if (lineCount > 500) {
    report.addFail(`body: ${String(lineCount)} lines (>500 spec target)`);
  } else if (lineCount > 400) {
    report.addWarn(
      `body: ${String(lineCount)} lines (>80% of 500-line target)`,
    );
  } else {
    report.addPass(`body: ${String(lineCount)} lines (under 500)`);
  }

  if (HEADING_RE.test(body)) {
    report.addPass("body: has at least one heading");
  } else {
    report.addFail("body: no headings found");
  }

  // Code blocks: count opening fences with and without language tag
  const fences: (string | undefined)[] = [];
  for (const m of body.matchAll(CODE_FENCE_OPEN_RE)) {
    fences.push(m[1]);
  }
  if (fences.length > 0) {
    const opensNoLang = fences.filter((f) => !f).length;
    const opensWithLang = fences.length - opensNoLang;
    // Each ``` is either an opening or a closing; the count of openings-with-lang
    // tells us how many blocks have a language tag. The bare ``` are either
    // openings-without-lang OR closings. Heuristic: there are usually pairs,
    // so unmarked-openings ≈ (bare_count) - (with_lang_count).
    const unmarked = opensNoLang - opensWithLang;
    if (unmarked > 0) {
      report.addWarn(
        `body: ${String(unmarked)} code block(s) appear to lack a language marker`,
      );
    } else {
      report.addPass(
        `body: all code block(s) have language markers ` +
          `(${String(opensWithLang)} language-tagged opening fence(s) found)`,
      );
    }
  }

  if (/^## *Gotchas/m.test(body)) {
    report.addPass("content: '## Gotchas' section present");
  } else {
    report.addWarn(
      "content: no '## Gotchas' section — non-obvious env-specific facts have no home",
    );
  }
};

const checkReferences = (
  body: string,
  skillDir: string,
  report: Report,
): void => {
  // Find unique reference paths
  const found = new Set<string>();
  for (const m of body.matchAll(REF_LINK_RE)) {
    if (m[1] !== undefined) found.add(m[1]);
  }
  const refs = [...found].toSorted();
  if (refs.length === 0) {
    return;
  }

  let broken = 0;
  let placeholder = 0;

  for (const ref of refs) {
    // Skip placeholder paths (contain <…> or {…})
    if (ref.includes("<") || ref.includes("{")) {
      placeholder += 1;
      continue;
    }

    const target = join(skillDir, "references", ref);
    if (!isFile(target)) {
      report.addFail(`references: broken link → references/${ref}`);
      broken += 1;
    }

    if (ref.includes("/")) {
      report.addWarn(
        `references: 'references/${ref}' is nested deeper than one level`,
      );
    }

    const lastSegment = ref.includes("/")
      ? ref.slice(ref.lastIndexOf("/") + 1)
      : ref;
    const fname = lastSegment.endsWith(".md")
      ? lastSegment.slice(0, -".md".length)
      : lastSegment;
    if (!NAME_RE.test(fname)) {
      report.addWarn(
        `references: 'references/${ref}' filename is not kebab-case`,
      );
    }
  }

  const realRefs = refs.length - placeholder;
  if (realRefs > 0) {
    const resolved = realRefs - broken;
    if (broken === 0) {
      report.addPass(
        `references: ${String(resolved)}/${String(realRefs)} link(s) resolve`,
      );
    }
  }

  // @references prefix (defeats progressive disclosure)
  if (body.includes("@references/")) {
    report.addFail(
      "references: '@references/' prefix found — strips progressive disclosure " +
        "(use plain 'references/')",
    );
  } else {
    report.addPass("references: no '@references/' prefix");
  }

  // Load-when triggers in Progressive Disclosure section only
  const pdMatch = PROGRESSIVE_DISCLOSURE_RE.exec(body);
  if (pdMatch) {
    const pdBlock = pdMatch[1] ?? "";
    // Count lines referencing references/X.md within the PD block
    const refLineRe = /references\/[\w./<>{}-]+\.md/;
    const pdLines = splitLines(pdBlock).filter((line) => refLineRe.test(line));
    if (pdLines.length > 0) {
      const withTrigger = pdLines.filter((line) =>
        LOAD_WHEN_RE.test(line),
      ).length;
      const missing = pdLines.length - withTrigger;
      if (missing > 0) {
        report.addWarn(
          `references: ${String(missing)} of ${String(pdLines.length)} link(s) in Progressive ` +
            "Disclosure lack a 'Load when…' trigger",
        );
      } else {
        report.addPass(
          `references: all ${String(pdLines.length)} link(s) in Progressive Disclosure ` +
            "carry a 'Load when…' trigger",
        );
      }
    }
  }
};

const checkReferenceTocs = (skillDir: string, report: Report): void => {
  const refsDir = join(skillDir, "references");
  if (!isDir(refsDir)) {
    return;
  }
  const tocRe = /^## *(?:Contents|Table of Contents)\b/im;
  const missing: string[] = [];
  let checked = 0;

  const mdFiles = readdirSync(refsDir)
    .filter((entry) => entry.endsWith(".md"))
    .filter((entry) => isFile(join(refsDir, entry)))
    .toSorted();

  for (const entry of mdFiles) {
    const lines = splitLines(readFileSync(join(refsDir, entry), "utf8"));
    if (lines.length <= 100) {
      continue;
    }
    checked += 1;
    if (!tocRe.test(lines.slice(0, 15).join("\n"))) {
      missing.push(`${entry} (${String(lines.length)} lines)`);
    }
  }
  if (checked === 0) {
    return;
  }
  if (missing.length > 0) {
    report.addWarn(
      "references: file(s) >100 lines lack a '## Contents' table of contents — " +
        missing.join(", "),
    );
  } else {
    report.addPass(
      `references: all ${String(checked)} reference file(s) >100 lines open with a ` +
        "'## Contents' table of contents",
    );
  }
};

const checkHarnessNeutrality = (body: string, report: Report): void => {
  let hits = 0;
  for (const [pattern, label] of HARNESS_PATTERNS) {
    const lines = splitLines(body);
    for (const [i, line] of lines.entries()) {
      if (pattern.test(line)) {
        report.addFail(
          `harness-neutrality: ${label} — line ${String(i + 1)}: ${line.trim().slice(0, 80)}`,
        );
        hits += 1;
      }
    }
  }
  if (hits === 0) {
    report.addPass("harness-neutrality: clean");
  }
};

const renderReport = (
  report: Report,
  skillPath: string,
  skillDir: string,
): number => {
  console.log(`Validation: ${skillPath}`);
  console.log(`Skill dir: ${skillDir}`);
  console.log();
  for (const line of report.passes) {
    console.log(`[PASS] ${line}`);
  }
  for (const line of report.warnings) {
    console.log(`[WARN] ${line}`);
  }
  for (const line of report.errors) {
    console.log(`[FAIL] ${line}`);
  }
  console.log();

  const nErr = report.errors.length;
  const nWarn = report.warnings.length;
  if (nErr > 0) {
    console.log(
      `Result: FAIL (${String(nErr)} error(s), ${String(nWarn)} warning(s))`,
    );
    return 1;
  }
  if (nWarn > 0) {
    console.log(`Result: PASS with ${String(nWarn)} warning(s)`);
    return 0;
  }
  console.log("Result: PASS (no warnings)");
  return 0;
};

const HELP_TEXT = [
  "usage: moon-skill-validate [-h] [target]",
  "",
  "Validate a SKILL.md against the Agent Skills spec and authoring best practices.",
  "",
  "positional arguments:",
  "  target      skill directory or path to a SKILL.md (defaults to the current directory)",
  "",
  "options:",
  "  -h, --help  show this help message and exit",
].join("\n");

const main = (argv: readonly string[]): number => {
  if (argv.includes("-h") || argv.includes("--help")) {
    console.log(HELP_TEXT);
    return 0;
  }

  // Single positional target; ERGONOMIC ADDITION: default to cwd when omitted.
  const positional = argv.find((a) => !a.startsWith("-"));
  const target = positional ?? process.cwd();

  const {skill: skillPath, skillDir} = resolveTarget(target);
  const parentName = basename(skillDir);
  const content = readFileSync(skillPath, "utf8");

  let fm: Frontmatter;
  let body: string;
  let fmRaw: string;
  try {
    ({fm, body, fmRaw} = splitFrontmatter(content));
  } catch (error) {
    process.stderr.write(
      `Error: ${error instanceof Error ? error.message : String(error)}\n`,
    );
    return 2;
  }

  const report = new Report();
  checkFrontmatter(fm, parentName, report);
  checkLoaderQuoting(fmRaw, report);
  checkBody(body, report);
  checkReferences(body, skillDir, report);
  checkReferenceTocs(skillDir, report);
  checkHarnessNeutrality(body, report);

  return renderReport(report, skillPath, skillDir);
};

process.exit(main(process.argv.slice(2)));
