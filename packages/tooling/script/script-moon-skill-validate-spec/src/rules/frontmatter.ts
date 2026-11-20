import {
  isSkillName,
  SKILL_NAME_RULE,
} from "@xonovex/script-moon-common/skill-name";
import {
  fail,
  pass,
  warn,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import type {Frontmatter} from "../skill-document.js";
import {KEBAB_CASE_RE} from "../text-lines.js";
import {descriptionRoutingErrors} from "./description-routing.js";

// A vendor adapter skill carries the vendor in its name by design, so the
// reserved-word rule lets it through.
const RESERVED_VENDOR_ADAPTER_NAMES = new Set(["configuring-claude-code"]);

// The routing run on the workflow skills kept descriptions of up to this
// length; the specification allows 1,024.
const DESCRIPTION_LIMIT = 820;

const KNOWN_FIELDS = new Set([
  "name",
  "description",
  "license",
  "compatibility",
  "metadata",
  "allowed-tools",
]);

// yamlTypeName returns stable names for surfaced scalar and collection types.
const yamlTypeName = (value: unknown): string => {
  if (value === null) return "NoneType";
  if (typeof value === "boolean") return "bool";
  if (typeof value === "number")
    return Number.isInteger(value) ? "int" : "float";
  if (typeof value === "string") return "str";
  if (Array.isArray(value)) return "list";
  if (typeof value === "object") return "dict";
  return typeof value;
};

const nameShape = (name: unknown, parentName: string): Finding => {
  if (!name) return fail("frontmatter: missing 'name'");
  if (typeof name !== "string") {
    return fail(`frontmatter: name is ${yamlTypeName(name)}, expected string`);
  }
  if (name.length > 64) {
    return fail(
      `frontmatter: name '${name}' is ${String(name.length)} chars (>64 spec limit)`,
    );
  }
  if (!KEBAB_CASE_RE.test(name)) {
    return fail(
      `frontmatter: name '${name}' is not kebab-case ` +
        `(regex ^[a-z0-9]+(-[a-z0-9]+)*$)`,
    );
  }
  if (!isSkillName(name)) {
    return fail(
      `frontmatter: name '${name}' is not a gerund and then its object ` +
        `(regex ${SKILL_NAME_RULE})`,
    );
  }
  if (name !== parentName) {
    return fail(
      `frontmatter: name '${name}' != parent directory '${parentName}'`,
    );
  }
  return pass(
    `name: '${name}' (${String(name.length)} chars, gerund first, matches parent dir)`,
  );
};

const checkName = (fm: Frontmatter, parentName: string): readonly Finding[] => {
  const name = fm.name;
  if (typeof name !== "string") return [nameShape(name, parentName)];
  return [
    nameShape(name, parentName),
    ...(/anthropic|claude/i.test(name) &&
    !RESERVED_VENDOR_ADAPTER_NAMES.has(name)
      ? [
          fail(
            `frontmatter: name '${name}' uses a reserved word (anthropic/claude)`,
          ),
        ]
      : []),
    ...(/<[^>]+>/.test(name)
      ? [fail(`frontmatter: name '${name}' contains an XML tag`)]
      : []),
  ];
};

const checkDescriptionText = (
  desc: string,
  name: unknown,
): readonly Finding[] => {
  const routingErrors = descriptionRoutingErrors(
    desc,
    typeof name === "string" ? name : undefined,
  );
  return [
    desc.length > DESCRIPTION_LIMIT
      ? fail(
          `description: ${String(desc.length)} chars (over the ${String(DESCRIPTION_LIMIT)}-character catalog limit)`,
        )
      : pass(
          `description: ${String(desc.length)} chars (within ${String(DESCRIPTION_LIMIT)})`,
        ),
    /^\s*use (?:this skill )?when/i.test(desc)
      ? pass("description: imperative phrasing ('Use when…')")
      : warn(
          "description: does not start with 'Use when…' / 'Use this skill when…' — " +
            "agents may not trigger reliably",
        ),
    /triggers? on/i.test(desc)
      ? pass("description: includes trigger contexts ('Triggers on…')")
      : warn(
          "description: missing 'Triggers on…' — agents lose the trigger keyword list",
        ),
    /even when the user doesn'?t/i.test(desc)
      ? pass(
          "description: includes non-obvious-trigger clause " +
            "('even when the user doesn't say…')",
        )
      : warn(
          "description: no 'even when the user doesn't say…' clause — " +
            "may miss implicit triggers",
        ),
    ...routingErrors.map((error) => fail(error)),
    ...(routingErrors.length === 0
      ? [
          pass(
            "description: routing uses positive triggers without other skill names",
          ),
        ]
      : []),
    ...(/<[^>]+>/.test(desc)
      ? [
          warn(
            "description: contains angle-bracket markup (`<…>`) — fine if it's a " +
              "component/generic reference, otherwise remove the XML tag",
          ),
        ]
      : []),
  ];
};

const checkDescription = (fm: Frontmatter): readonly Finding[] => {
  const desc = fm.description;
  if (!desc) return [fail("frontmatter: missing 'description'")];
  if (typeof desc !== "string") {
    return [
      fail(
        `frontmatter: description is ${yamlTypeName(desc)}, expected string`,
      ),
    ];
  }
  return checkDescriptionText(desc, fm.name);
};

const checkCompatibility = (fm: Frontmatter): readonly Finding[] => {
  const compat = fm.compatibility;
  if (compat === undefined || compat === null) return [];
  if (typeof compat !== "string") {
    return [
      fail(
        `frontmatter: compatibility is ${yamlTypeName(compat)}, expected string`,
      ),
    ];
  }
  return compat.length > 500
    ? [
        fail(
          `frontmatter: compatibility is ${String(compat.length)} chars (>500)`,
        ),
      ]
    : [pass(`compatibility: ${String(compat.length)} chars (under 500)`)];
};

const checkLicense = (fm: Frontmatter): readonly Finding[] => {
  const lic = fm.license;
  return lic === undefined || lic === null || typeof lic === "string"
    ? []
    : [warn(`frontmatter: license is ${yamlTypeName(lic)}, expected string`)];
};

const checkMetadata = (fm: Frontmatter): readonly Finding[] => {
  const meta = fm.metadata;
  if (meta === undefined || meta === null) return [];
  const isStringMap =
    typeof meta === "object" &&
    !Array.isArray(meta) &&
    Object.values(meta as Record<string, unknown>).every(
      (value) => typeof value === "string",
    );
  return isStringMap
    ? []
    : [warn("frontmatter: metadata should be a string→string map")];
};

const checkAllowedTools = (fm: Frontmatter): readonly Finding[] => {
  const allowed = fm["allowed-tools"];
  if (allowed === undefined || allowed === null) return [];
  if (typeof allowed !== "string") {
    return [
      warn(
        `frontmatter: allowed-tools is ${yamlTypeName(allowed)}, ` +
          "expected a space-separated string",
      ),
    ];
  }
  return /^[\w():*,.\- ]+$/.test(allowed)
    ? []
    : [
        warn(
          "frontmatter: allowed-tools has unexpected characters " +
            "(expected space-separated tokens like 'Bash(git:*) Read')",
        ),
      ];
};

const checkUnknownFields = (fm: Frontmatter): readonly Finding[] => {
  const unknown = Object.keys(fm)
    .filter((key) => !KNOWN_FIELDS.has(key))
    .toSorted();
  return unknown.length > 0
    ? [warn("frontmatter: unknown top-level field(s): " + unknown.join(", "))]
    : [];
};

/** The Agent Skills frontmatter fields, checked against the parent directory name. */
export const checkFrontmatter = (
  fm: Frontmatter,
  parentName: string,
): readonly Finding[] => [
  ...checkName(fm, parentName),
  ...checkDescription(fm),
  ...checkCompatibility(fm),
  ...checkLicense(fm),
  ...checkMetadata(fm),
  ...checkAllowedTools(fm),
  ...checkUnknownFields(fm),
];
