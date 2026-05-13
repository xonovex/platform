#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = [
#   "pyyaml>=6.0",
# ]
# ///
"""Validate a SKILL.md against the Agent Skills spec and authoring best practices.

Usage:
    validate.py <skill-dir>
    validate.py <path-to-SKILL.md>

Exit codes:
    0 = PASS (no errors; warnings allowed)
    1 = FAIL (one or more errors)
    2 = usage error / file not found

Read-only — never modifies files.
"""

from __future__ import annotations

import re
import sys
from dataclasses import dataclass, field
from pathlib import Path

import yaml

NAME_RE = re.compile(r"^[a-z0-9]+(-[a-z0-9]+)*$")
FRONTMATTER_RE = re.compile(r"^---\n(.*?)\n---\n(.*)$", re.DOTALL)
REF_LINK_RE = re.compile(r"references/([a-zA-Z0-9_./<>{}-]+\.md)")
PROGRESSIVE_DISCLOSURE_RE = re.compile(
    r"^## *Progressive Disclosure\s*\n(.*?)(?=^## |\Z)",
    re.MULTILINE | re.DOTALL | re.IGNORECASE,
)
HEADING_RE = re.compile(r"^#{1,6} ", re.MULTILINE)
CODE_FENCE_OPEN_RE = re.compile(r"^```([a-zA-Z0-9_+-]+)?\s*$", re.MULTILINE)
LOAD_WHEN_RE = re.compile(r"load when", re.IGNORECASE)

HARNESS_PATTERNS = [
    # Proprietary tool / function / mode names
    (r"\bEnterPlanMode\b", "Claude Code tool name"),
    (r"\bExitPlanMode\b", "Claude Code tool name"),
    (r"\bAskUserQuestion\b", "Claude Code tool name"),
    (r"\bTodoWrite\b", "Claude Code tool name"),
    (r"\bsubagent_type\b", "Claude Code parameter"),
    (r"\bmodel=haiku\b", "Vendor model id"),
    (r"\bmodel=sonnet\b", "Vendor model id"),
    (r"\bmodel=opus\b", "Vendor model id"),
    # Vendor-namespaced paths
    (r"\.claude/skills/", "Vendor-namespaced path"),
    (r"\.claude/commands/", "Vendor-namespaced path"),
    # Vendor-prefixed frontmatter keys
    (r"\bclaude:", "Vendor-prefixed frontmatter key"),
    # Vendor-specific instruction filenames
    (r"\bCLAUDE\.md\b", "Vendor-specific filename (use AGENTS.md)"),
]


@dataclass
class Report:
    passes: list[str] = field(default_factory=list)
    warnings: list[str] = field(default_factory=list)
    errors: list[str] = field(default_factory=list)

    def add_pass(self, msg: str) -> None:
        self.passes.append(msg)

    def add_warn(self, msg: str) -> None:
        self.warnings.append(msg)

    def add_fail(self, msg: str) -> None:
        self.errors.append(msg)


def usage() -> None:
    sys.stderr.write("Usage: validate.py <skill-dir or SKILL.md>\n")
    sys.exit(2)


def resolve_target(arg: str) -> tuple[Path, Path]:
    """Return (skill_md_path, skill_dir)."""
    target = Path(arg).resolve()
    if target.is_dir():
        skill = target / "SKILL.md"
    elif target.is_file():
        skill = target
    else:
        sys.stderr.write(f"Error: target not found: {target}\n")
        sys.exit(2)
    if not skill.is_file():
        sys.stderr.write(f"Error: SKILL.md not found at {skill}\n")
        sys.exit(2)
    return skill, skill.parent


def split_frontmatter(content: str) -> tuple[dict, str]:
    match = FRONTMATTER_RE.match(content)
    if not match:
        return {}, content
    fm_raw, body = match.groups()
    try:
        fm = yaml.safe_load(fm_raw) or {}
    except yaml.YAMLError as e:
        raise ValueError(f"Invalid YAML frontmatter: {e}") from e
    if not isinstance(fm, dict):
        raise ValueError("Frontmatter is not a mapping")
    return fm, body


def check_frontmatter(fm: dict, parent_name: str, report: Report) -> None:
    # name
    name = fm.get("name")
    if not name:
        report.add_fail("frontmatter: missing 'name'")
    elif not isinstance(name, str):
        report.add_fail(f"frontmatter: name is {type(name).__name__}, expected string")
    elif len(name) > 64:
        report.add_fail(f"frontmatter: name '{name}' is {len(name)} chars (>64 spec limit)")
    elif not NAME_RE.match(name):
        report.add_fail(
            f"frontmatter: name '{name}' is not kebab-case "
            f"(regex ^[a-z0-9]+(-[a-z0-9]+)*$)"
        )
    elif name != parent_name:
        report.add_fail(f"frontmatter: name '{name}' != parent directory '{parent_name}'")
    else:
        report.add_pass(f"name: '{name}' ({len(name)} chars, kebab-case, matches parent dir)")

    # description
    desc = fm.get("description")
    if not desc:
        report.add_fail("frontmatter: missing 'description'")
    elif not isinstance(desc, str):
        report.add_fail(f"frontmatter: description is {type(desc).__name__}, expected string")
    else:
        desc_len = len(desc)
        if desc_len > 1024:
            report.add_fail(f"description: {desc_len} chars (>1024 spec limit)")
        elif desc_len > 820:
            report.add_warn(
                f"description: {desc_len} chars (>80% of 1024 limit; consider trimming)"
            )
            report.add_pass(f"description present ({desc_len} chars)")
        else:
            report.add_pass(f"description: {desc_len} chars (under 1024)")

        if re.match(r"^\s*use (this skill )?when", desc, re.IGNORECASE):
            report.add_pass("description: imperative phrasing ('Use when…')")
        else:
            report.add_warn(
                "description: does not start with 'Use when…' / 'Use this skill when…' — "
                "agents may not trigger reliably"
            )

        if re.search(r"triggers? on", desc, re.IGNORECASE):
            report.add_pass("description: includes trigger contexts ('Triggers on…')")
        else:
            report.add_warn(
                "description: missing 'Triggers on…' — agents lose the trigger keyword list"
            )

        if re.search(r"even when the user doesn'?t", desc, re.IGNORECASE):
            report.add_pass(
                "description: includes non-obvious-trigger clause "
                "('even when the user doesn't say…')"
            )
        else:
            report.add_warn(
                "description: no 'even when the user doesn't say…' clause — "
                "may miss implicit triggers"
            )

    # compatibility (optional)
    compat = fm.get("compatibility")
    if compat is not None:
        if not isinstance(compat, str):
            report.add_fail(
                f"frontmatter: compatibility is {type(compat).__name__}, expected string"
            )
        elif len(compat) > 500:
            report.add_fail(f"frontmatter: compatibility is {len(compat)} chars (>500)")
        else:
            report.add_pass(f"compatibility: {len(compat)} chars (under 500)")


def check_body(body: str, report: Report) -> None:
    line_count = body.count("\n")
    if line_count > 500:
        report.add_fail(f"body: {line_count} lines (>500 spec target)")
    elif line_count > 400:
        report.add_warn(f"body: {line_count} lines (>80% of 500-line target)")
    else:
        report.add_pass(f"body: {line_count} lines (under 500)")

    if HEADING_RE.search(body):
        report.add_pass("body: has at least one heading")
    else:
        report.add_fail("body: no headings found")

    # Code blocks: count opening fences with and without language tag
    fences = CODE_FENCE_OPEN_RE.findall(body)
    if fences:
        opens_no_lang = sum(1 for f in fences if not f)
        opens_with_lang = len(fences) - opens_no_lang
        # Each ``` is either an opening or a closing; the count of openings-with-lang
        # tells us how many blocks have a language tag. The bare ``` are either
        # openings-without-lang OR closings. Heuristic: there are usually pairs,
        # so unmarked-openings ≈ (bare_count) - (with_lang_count).
        unmarked = opens_no_lang - opens_with_lang
        if unmarked > 0:
            report.add_warn(
                f"body: {unmarked} code block(s) appear to lack a language marker"
            )
        else:
            report.add_pass(
                f"body: all code block(s) have language markers "
                f"({opens_with_lang} language-tagged opening fence(s) found)"
            )

    if re.search(r"^## *Gotchas", body, re.MULTILINE):
        report.add_pass("content: '## Gotchas' section present")
    else:
        report.add_warn(
            "content: no '## Gotchas' section — non-obvious env-specific facts have no home"
        )


def check_references(body: str, skill_dir: Path, report: Report) -> None:
    # Find unique reference paths
    refs = sorted({m for m in REF_LINK_RE.findall(body)})
    if not refs:
        return

    broken = 0
    deep = 0
    bad_name = 0
    placeholder = 0

    for ref in refs:
        # Skip placeholder paths (contain <…> or {…})
        if "<" in ref or "{" in ref:
            placeholder += 1
            continue

        target = skill_dir / "references" / ref
        if not target.is_file():
            report.add_fail(f"references: broken link → references/{ref}")
            broken += 1

        if "/" in ref:
            report.add_warn(f"references: 'references/{ref}' is nested deeper than one level")
            deep += 1

        fname = ref.rsplit("/", 1)[-1].removesuffix(".md")
        if not NAME_RE.match(fname):
            report.add_warn(f"references: 'references/{ref}' filename is not kebab-case")
            bad_name += 1

    real_refs = len(refs) - placeholder
    if real_refs > 0:
        resolved = real_refs - broken
        if broken == 0:
            report.add_pass(f"references: {resolved}/{real_refs} link(s) resolve")

    # @references prefix (defeats progressive disclosure)
    if "@references/" in body:
        report.add_fail(
            "references: '@references/' prefix found — strips progressive disclosure "
            "(use plain 'references/')"
        )
    else:
        report.add_pass("references: no '@references/' prefix")

    # Load-when triggers in Progressive Disclosure section only
    pd_match = PROGRESSIVE_DISCLOSURE_RE.search(body)
    if pd_match:
        pd_block = pd_match.group(1)
        # Count lines referencing references/X.md within the PD block
        pd_lines = [
            line for line in pd_block.splitlines()
            if re.search(r"references/[a-zA-Z0-9_./<>{}-]+\.md", line)
        ]
        if pd_lines:
            with_trigger = sum(1 for line in pd_lines if LOAD_WHEN_RE.search(line))
            missing = len(pd_lines) - with_trigger
            if missing > 0:
                report.add_warn(
                    f"references: {missing} of {len(pd_lines)} link(s) in Progressive "
                    "Disclosure lack a 'Load when…' trigger"
                )
            else:
                report.add_pass(
                    f"references: all {len(pd_lines)} link(s) in Progressive Disclosure "
                    "carry a 'Load when…' trigger"
                )


def check_harness_neutrality(body: str, report: Report) -> None:
    hits = 0
    for pattern, label in HARNESS_PATTERNS:
        for line_no, line in enumerate(body.splitlines(), start=1):
            if re.search(pattern, line):
                report.add_fail(
                    f"harness-neutrality: {label} — line {line_no}: {line.strip()[:80]}"
                )
                hits += 1
    if hits == 0:
        report.add_pass("harness-neutrality: clean")


def render_report(report: Report, skill_path: Path, skill_dir: Path) -> int:
    print(f"Validation: {skill_path}")
    print(f"Skill dir: {skill_dir}")
    print()
    for line in report.passes:
        print(f"[PASS] {line}")
    for line in report.warnings:
        print(f"[WARN] {line}")
    for line in report.errors:
        print(f"[FAIL] {line}")
    print()

    n_err = len(report.errors)
    n_warn = len(report.warnings)
    if n_err > 0:
        print(f"Result: FAIL ({n_err} error(s), {n_warn} warning(s))")
        return 1
    if n_warn > 0:
        print(f"Result: PASS with {n_warn} warning(s)")
        return 0
    print("Result: PASS (no warnings)")
    return 0


def main(argv: list[str]) -> int:
    if len(argv) != 2:
        usage()
        return 2  # unreachable; usage() exits

    skill_path, skill_dir = resolve_target(argv[1])
    parent_name = skill_dir.name
    content = skill_path.read_text(encoding="utf-8")

    try:
        fm, body = split_frontmatter(content)
    except ValueError as e:
        sys.stderr.write(f"Error: {e}\n")
        return 2

    report = Report()
    check_frontmatter(fm, parent_name, report)
    check_body(body, report)
    check_references(body, skill_dir, report)
    check_harness_neutrality(body, report)

    return render_report(report, skill_path, skill_dir)


if __name__ == "__main__":
    sys.exit(main(sys.argv))
