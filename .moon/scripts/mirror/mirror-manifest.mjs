import {readFileSync, renameSync, rmSync, writeFileSync} from "node:fs";
import {basename, dirname, join} from "node:path";

const REPOSITORY = "drodan/drodan-platform";
const RESERVED_MIRROR_PATHS = new Set([
  "moon.yml",
  "package.json",
  "mirror.json",
]);
const CANONICAL_MESSAGE =
  "is not in canonical form (sorted keys, 2-space indent, LF, final newline)";
const USAGE =
  "usage: mirror-manifest.mjs check <manifest> | list <manifest> | write <manifest> <commit>";
const COMMIT_PATTERN = /^[0-9a-f]{40}$/;
const SHA256_PATTERN = /^[0-9a-f]{64}$/;
const PATH_CHARACTERS_PATTERN = /^[A-Za-z0-9._@+/-]+$/;

const quote = (value) => JSON.stringify(value);
const isObject = (value) =>
  typeof value === "object" && value !== null && !Array.isArray(value);
const has = (object, key) => Object.hasOwn(object, key);
const matches = (pattern, value) =>
  typeof value === "string" && pattern.test(value);

function pathProblem(path) {
  if (path.length === 0) return "is empty";
  if (path.startsWith("/")) return "must be relative";
  if (path.endsWith("/")) return "must not end with /";
  if (!PATH_CHARACTERS_PATTERN.test(path)) {
    return "contains a character outside A-Z a-z 0-9 . _ @ + / -";
  }
  const segments = path.split("/");
  if (segments.includes("")) return "contains an empty segment";
  if (segments.some((segment) => segment === "." || segment === "..")) {
    return "contains a . or .. segment";
  }
  return null;
}

function mirrorPathProblem(path) {
  const problem = pathProblem(path);
  if (problem) return problem;
  if (RESERVED_MIRROR_PATHS.has(path)) return "is a reserved project file name";
  return null;
}

function unknownKeyProblems(object, knownKeys, where) {
  return Object.keys(object)
    .filter((key) => !knownKeys.includes(key))
    .map((key) => `unknown key ${quote(key)} ${where}`);
}

function sourceProblems(source, strict) {
  if (!isObject(source)) return ["source is not an object"];
  const problems = strict
    ? unknownKeyProblems(source, ["repository", "commit"], "in source")
    : [];
  if (!has(source, "repository")) {
    problems.push("missing source.repository");
  } else if (source.repository !== REPOSITORY) {
    problems.push(`source.repository must be ${quote(REPOSITORY)}`);
  }
  if (strict) {
    if (!has(source, "commit")) {
      problems.push("missing source.commit");
    } else if (!matches(COMMIT_PATTERN, source.commit)) {
      problems.push("source.commit is not 40 lowercase hex characters");
    }
  }
  return problems;
}

function entryProblems(mirror, entry, strict) {
  const label = `files[${quote(mirror)}]`;
  if (!isObject(entry))
    return {problems: [`${label} is not an object`], row: null};
  const problems = strict
    ? unknownKeyProblems(entry, ["source", "sha256"], `in ${label}`)
    : [];
  if (!has(entry, "source")) {
    problems.push(`missing ${label}.source`);
  } else if (typeof entry.source !== "string") {
    problems.push(`${label}.source is not a string`);
  } else {
    const problem = pathProblem(entry.source);
    if (problem)
      problems.push(`${label}.source is not a valid source path: ${problem}`);
  }
  if (strict) {
    if (!has(entry, "sha256")) {
      problems.push(`missing ${label}.sha256`);
    } else if (!matches(SHA256_PATTERN, entry.sha256)) {
      problems.push(`${label}.sha256 is not 64 lowercase hex characters`);
    }
  }
  return {problems, row: {mirror, source: entry.source, sha256: entry.sha256}};
}

function filesProblems(files, strict) {
  if (!isObject(files)) return {problems: ["files is not an object"], rows: []};
  const problems = [];
  const rows = [];
  for (const mirror of Object.keys(files).sort()) {
    const label = `files[${quote(mirror)}]`;
    const problem = mirrorPathProblem(mirror);
    if (problem)
      problems.push(`${label} is not a valid mirror path: ${problem}`);
    const entry = entryProblems(mirror, files[mirror], strict);
    problems.push(...entry.problems);
    if (entry.row) rows.push(entry.row);
  }
  return {problems, rows};
}

function validate(value, strict) {
  if (!isObject(value)) return {problems: ["is not a JSON object"], rows: []};
  const problems = strict
    ? unknownKeyProblems(value, ["source", "files"], "at top level")
    : [];
  if (has(value, "source")) {
    problems.push(...sourceProblems(value.source, strict));
  } else {
    problems.push("missing source");
  }
  if (has(value, "files")) {
    const files = filesProblems(value.files, strict);
    problems.push(...files.problems);
    return {problems, rows: files.rows};
  }
  problems.push("missing files");
  return {problems, rows: []};
}

// Rebuilding in the canonical key order and comparing bytes catches every
// formatting departure at once: unsorted keys, duplicate keys, indentation,
// CRLF and a BOM.
function reorder(object, leadingKeys, mapValue = (_key, value) => value) {
  const keys = [
    ...leadingKeys.filter((key) => has(object, key)),
    ...Object.keys(object)
      .filter((key) => !leadingKeys.includes(key))
      .sort(),
  ];
  return Object.fromEntries(
    keys.map((key) => [key, mapValue(key, object[key])]),
  );
}

function canonicalFiles(files) {
  return reorder(files, [], (_mirror, entry) =>
    isObject(entry) ? reorder(entry, ["source", "sha256"]) : entry,
  );
}

function canonicalManifest(value) {
  return reorder(value, ["source", "files"], (key, child) => {
    if (key === "source" && isObject(child))
      return reorder(child, ["repository", "commit"]);
    if (key === "files" && isObject(child)) return canonicalFiles(child);
    return child;
  });
}

function canonicalText(value) {
  return `${JSON.stringify(canonicalManifest(value), null, 2)}\n`;
}

function loadManifest(manifestPath) {
  let bytes;
  try {
    bytes = readFileSync(manifestPath);
  } catch {
    return {problems: [`cannot read ${manifestPath}`]};
  }
  let raw;
  try {
    raw = new TextDecoder("utf-8", {fatal: true, ignoreBOM: true}).decode(
      bytes,
    );
  } catch {
    return {problems: ["is not valid UTF-8"]};
  }
  try {
    return {raw, value: JSON.parse(raw.replace(/^\uFEFF/, ""))};
  } catch {
    return {problems: ["is not valid JSON"]};
  }
}

function report(problems) {
  for (const problem of problems) {
    process.stderr.write(`mirror.json: ${problem}\n`);
  }
  return 1;
}

function check(manifestPath) {
  const loaded = loadManifest(manifestPath);
  if (loaded.problems) return report(loaded.problems);
  const {problems, rows} = validate(loaded.value, true);
  if (isObject(loaded.value) && loaded.raw !== canonicalText(loaded.value)) {
    problems.push(CANONICAL_MESSAGE);
  }
  if (problems.length > 0) return report(problems);
  const lines = [
    loaded.value.source.commit,
    ...rows.map((row) => `${row.mirror}\t${row.source}\t${row.sha256}`),
  ];
  process.stdout.write(`${lines.join("\n")}\n`);
  return 0;
}

function list(manifestPath) {
  const loaded = loadManifest(manifestPath);
  if (loaded.problems) return report(loaded.problems);
  const {problems, rows} = validate(loaded.value, false);
  if (problems.length > 0) return report(problems);
  const lines = rows.map((row) => `${row.mirror}\t${row.source}`);
  process.stdout.write(lines.map((line) => `${line}\n`).join(""));
  return 0;
}

function readRows(input) {
  const body = input.endsWith("\n") ? input.slice(0, -1) : input;
  return body === "" ? [] : body.split("\n");
}

function rowProblems(line, number, seen) {
  const fields = line.split("\t");
  if (fields.length !== 3) {
    return {
      problems: [
        `row ${number} has ${fields.length} tab-separated fields, expected 3`,
      ],
      row: null,
    };
  }
  const [mirror, source, sha256] = fields;
  const problems = [];
  const mirrorProblem = mirrorPathProblem(mirror);
  if (mirrorProblem)
    problems.push(
      `row ${number} mirror path ${quote(mirror)}: ${mirrorProblem}`,
    );
  if (seen.has(mirror))
    problems.push(`row ${number} duplicates mirror path ${quote(mirror)}`);
  seen.add(mirror);
  const sourceProblem = pathProblem(source);
  if (sourceProblem)
    problems.push(
      `row ${number} source path ${quote(source)}: ${sourceProblem}`,
    );
  if (!matches(SHA256_PATTERN, sha256)) {
    problems.push(`row ${number} sha256 is not 64 lowercase hex characters`);
  }
  return {problems, row: {mirror, source, sha256}};
}

function write(manifestPath, commit) {
  const problems = [];
  if (!matches(COMMIT_PATTERN, commit)) {
    problems.push("source.commit is not 40 lowercase hex characters");
  }
  const seen = new Set();
  const rows = [];
  readRows(readFileSync(0, "utf8")).forEach((line, index) => {
    const row = rowProblems(line, index + 1, seen);
    problems.push(...row.problems);
    if (row.row) rows.push(row.row);
  });
  if (problems.length > 0) return report(problems);

  const text = canonicalText({
    source: {repository: REPOSITORY, commit},
    files: Object.fromEntries(
      rows.map((row) => [row.mirror, {source: row.source, sha256: row.sha256}]),
    ),
  });
  let current = null;
  try {
    current = readFileSync(manifestPath);
  } catch {
    current = null;
  }
  if (current !== null && current.equals(Buffer.from(text, "utf8"))) return 0;

  const temporary = join(
    dirname(manifestPath),
    `.${basename(manifestPath)}.${process.pid}.tmp`,
  );
  try {
    writeFileSync(temporary, text);
    renameSync(temporary, manifestPath);
  } catch {
    rmSync(temporary, {force: true});
    return report([`cannot write ${manifestPath}`]);
  }
  return 0;
}

function main(args) {
  const [command, ...operands] = args;
  switch (command) {
    case "check":
      if (operands.length === 1) return check(operands[0]);
      break;
    case "list":
      if (operands.length === 1) return list(operands[0]);
      break;
    case "write":
      if (operands.length === 2) return write(operands[0], operands[1]);
      break;
    default:
      break;
  }
  process.stderr.write(`${USAGE}\n`);
  return 2;
}

process.exitCode = main(process.argv.slice(2));
