import {
  fail,
  pass,
  type Finding,
} from "@xonovex/script-moon-skill-catalog-common/finding";
import {splitLines} from "../text-lines.js";

// A skill runs under every agent harness, so the names one vendor's harness
// gives its tools, models, paths and files do not belong in it.
const HARNESS_PATTERNS: readonly (readonly [RegExp, string])[] = [
  [/\bEnterPlanMode\b/, "Claude Code tool name"],
  [/\bExitPlanMode\b/, "Claude Code tool name"],
  [/\bAskUserQuestion\b/, "Claude Code tool name"],
  [/\bTodoWrite\b/, "Claude Code tool name"],
  [/\bsubagent_type\b/, "Claude Code parameter"],
  [/\bmodel=haiku\b/, "Vendor model id"],
  [/\bmodel=sonnet\b/, "Vendor model id"],
  [/\bmodel=opus\b/, "Vendor model id"],
  [/\.claude\/skills\//, "Vendor-namespaced path"],
  [/\.claude\/commands\//, "Vendor-namespaced path"],
  [/\bclaude:/, "Vendor-prefixed frontmatter key"],
  [/\bCLAUDE\.md\b/, "Vendor-specific filename (use AGENTS.md)"],
];

export const checkHarnessNeutrality = (body: string): readonly Finding[] => {
  const lines = splitLines(body);
  const hits = HARNESS_PATTERNS.flatMap(([pattern, label]) =>
    lines.flatMap((line, index) =>
      pattern.test(line)
        ? [
            fail(
              `harness-neutrality: ${label} — line ${String(index + 1)}: ${line.trim().slice(0, 80)}`,
            ),
          ]
        : [],
    ),
  );
  return hits.length > 0 ? hits : [pass("harness-neutrality: clean")];
};
