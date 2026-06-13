#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Audit a skill's upstream sources for drift, against the SOURCES.md convention.

Skills here are distilled from external docs rather than a vendored source
checkout, so there are no version pins to bump. Instead every skill records its
provenance in SOURCES.md: one block per source with a URL, a "Last reviewed"
date, and a mapping of which SKILL.md sections / reference files each source
feeds. This tool closes the mechanical part of keeping that fresh and points a
maintainer at the rest.

What it does, deterministically:
  1. Parse each source block in the skill's SOURCES.md (title, URL, last
     reviewed, and the reference files it feeds).
  2. Report staleness: days since "Last reviewed" vs --max-age.
  3. Report dangling provenance: referenced reference files that no longer exist.
  4. Report reference files NOT covered by any source (repo-original, or a
     provenance gap — informational).
  5. (--fetch) confirm each URL still resolves and report Last-Modified.
  6. (--mark-reviewed) stamp matching sources' "Last reviewed" date to today —
     the write-action analog of bumping a version pin, done AFTER a human or
     agent has re-verified the distilled prose against the source.

What it deliberately does NOT do: rewrite the distilled prose. That stays a
human (or follow-up agent) decision — the report tells you exactly which
reference files to review.

Usage:
    audit-sources.py <skill-dir> [options]
    audit-sources.py --all [root] [options]

      skill-dir            path to a skill dir (containing SOURCES.md) or to a
                           SOURCES.md file directly
      --all [root]         audit every */SOURCES.md under root (default: cwd)
      --max-age DAYS       staleness threshold in days (default: 180)
      --fetch              HTTP-check each URL still resolves (stdlib urllib)
      --mark-reviewed [T]  stamp "Last reviewed" to today for sources whose title
                           contains T (case-insensitive); omit T to stamp all
      --json               emit a JSON report instead of text
      -h, --help           show this help

Exit codes: 0 = all sources fresh and provenance intact; 1 = stale source,
dangling provenance, or (with --fetch) an unreachable URL; 2 = usage / IO error.

Cross-platform: stdlib only.
"""

from __future__ import annotations

import argparse
import datetime as dt
import json
import re
import sys
import urllib.error
import urllib.request
from dataclasses import dataclass, field
from pathlib import Path

URL_RE = re.compile(r"\*\*URL:\*\*\s*(\S+)")
REVIEWED_RE = re.compile(r"\*\*Last reviewed:\*\*\s*(\d{4}-\d{2}-\d{2})")
REF_RE = re.compile(r"references/([a-z0-9][a-z0-9-]*\.md)")
HEADER_RE = re.compile(r"^##\s+(.*\S)\s*$")


@dataclass
class Source:
    title: str
    url: str | None = None
    reviewed: dt.date | None = None
    reviewed_raw: str | None = None
    refs: set[str] = field(default_factory=set)
    line_no: int = 0  # 0-based index of the header line


def parse_sources(text: str) -> list[Source]:
    """Parse SOURCES.md into source blocks. Blocks without a URL are notes, skipped."""
    sources: list[Source] = []
    current: Source | None = None
    for i, line in enumerate(text.splitlines()):
        header = HEADER_RE.match(line)
        if header:
            current = Source(title=header.group(1), line_no=i)
            sources.append(current)
            continue
        if current is None:
            continue
        if m := URL_RE.search(line):
            current.url = m.group(1)
        if m := REVIEWED_RE.search(line):
            current.reviewed_raw = m.group(1)
            try:
                current.reviewed = dt.date.fromisoformat(m.group(1))
            except ValueError:
                current.reviewed = None
        current.refs.update(REF_RE.findall(line))
    # A block is a real source only if it carries a URL (others are prose notes).
    return [s for s in sources if s.url]


def resolve_guide_dir(base: Path) -> Path:
    """The skill/guide dir for a base dir.

    base itself if base/SKILL.md exists, else the single immediate subdir
    containing SKILL.md (skill-package layout, e.g. skill-c99/c99-guide/SKILL.md).
    >1 match is ambiguous; 0 falls back to base so the existing not-found error
    still fires.
    """
    if (base / "SKILL.md").is_file():
        return base
    nested = sorted(p.parent for p in base.glob("*/SKILL.md") if p.is_file())
    if len(nested) > 1:
        sys.stderr.write(f"error: multiple SKILL.md found under {base}; pass one explicitly\n")
        raise SystemExit(2)
    return nested[0] if nested else base


def resolve_sources_file(target: Path) -> Path | None:
    if target.is_file() and target.name == "SOURCES.md":
        return target
    if target.is_dir():
        if (target / "SOURCES.md").is_file():
            return target / "SOURCES.md"
        # No SOURCES.md directly: descend into the single guide subdir (identified
        # by SKILL.md, per the skill-package layout) and use its SOURCES.md.
        guide_dir = resolve_guide_dir(target)
        if (guide_dir / "SOURCES.md").is_file():
            return guide_dir / "SOURCES.md"
    return None


def fetch_status(url: str, timeout: int = 15) -> tuple[str, str]:
    """Return (status, detail). status in {ok, redirect, missing, error}."""
    req = urllib.request.Request(
        url, method="GET", headers={"User-Agent": "skill-source-audit/1"}
    )
    try:
        with urllib.request.urlopen(req, timeout=timeout) as resp:  # noqa: S310 (trusted URLs)
            code = resp.getcode()
            lm = resp.headers.get("Last-Modified", "")
            detail = f"HTTP {code}" + (f", Last-Modified: {lm}" if lm else "")
            return ("ok" if code and code < 400 else "error", detail)
    except urllib.error.HTTPError as e:
        return ("missing" if e.code == 404 else "error", f"HTTP {e.code}")
    except (urllib.error.URLError, TimeoutError, ValueError) as e:
        return ("error", f"unreachable: {e}")


def audit_skill(
    sources_file: Path, max_age: int, do_fetch: bool, today: dt.date
) -> dict:
    skill_dir = sources_file.parent
    text = sources_file.read_text(encoding="utf-8")
    sources = parse_sources(text)

    existing_refs = {p.name for p in (skill_dir / "references").glob("*.md")} \
        if (skill_dir / "references").is_dir() else set()
    covered: set[str] = set()

    src_reports: list[dict] = []
    problems = 0
    for s in sources:
        age = (today - s.reviewed).days if s.reviewed else None
        stale = age is not None and age > max_age
        dangling = sorted(r for r in s.refs if r not in existing_refs)
        covered.update(s.refs)
        report: dict = {
            "title": s.title,
            "url": s.url,
            "last_reviewed": s.reviewed_raw,
            "age_days": age,
            "stale": stale,
            "refs": sorted(s.refs),
            "dangling_refs": dangling,
        }
        if stale or dangling:
            problems += 1
        if do_fetch and s.url:
            status, detail = fetch_status(s.url)
            report["fetch"] = {"status": status, "detail": detail}
            if status != "ok":
                problems += 1
        src_reports.append(report)

    uncovered = sorted(existing_refs - covered)
    return {
        "skill": skill_dir.resolve().name,
        "sources_file": str(sources_file),
        "source_count": len(sources),
        "sources": src_reports,
        "uncovered_refs": uncovered,  # repo-original or provenance gap (informational)
        "problems": problems,
    }


def mark_reviewed(sources_file: Path, title_filter: str | None, today: dt.date) -> list[str]:
    """Stamp 'Last reviewed' to today for matching source blocks. Returns titles edited."""
    lines = sources_file.read_text(encoding="utf-8").splitlines(keepends=True)
    needle = title_filter.lower() if title_filter else None
    edited: list[str] = []
    current_title = ""
    matched = False
    for i, line in enumerate(lines):
        header = HEADER_RE.match(line.rstrip("\n"))
        if header:
            current_title = header.group(1) or ""
            matched = needle is None or needle in current_title.lower()
            continue
        if matched and REVIEWED_RE.search(line):
            lines[i] = REVIEWED_RE.sub(f"**Last reviewed:** {today.isoformat()}", line)
            if current_title and current_title not in edited:
                edited.append(current_title)
    if edited:
        sources_file.write_text("".join(lines), encoding="utf-8")
    return edited


def print_text_report(rep: dict, max_age: int) -> None:
    print(f"skill: {rep['skill']}  ({rep['sources_file']})")
    print(f"sources: {rep['source_count']}  max-age: {max_age}d")
    for s in rep["sources"]:
        flags = []
        if s["stale"]:
            flags.append(f"STALE ({s['age_days']}d)")
        if s["dangling_refs"]:
            flags.append(f"DANGLING: {', '.join(s['dangling_refs'])}")
        fetch = s.get("fetch")
        if fetch and fetch["status"] != "ok":
            flags.append(f"URL {fetch['status']}")
        marker = "  ".join(flags) if flags else "ok"
        print(f"\n  [{marker}] {s['title']}")
        print(f"    url           : {s['url']}")
        print(f"    last reviewed : {s['last_reviewed'] or '(none)'}"
              + (f"  ({s['age_days']}d ago)" if s["age_days"] is not None else ""))
        if fetch:
            print(f"    fetch         : {fetch['detail']}")
        if s["refs"]:
            print(f"    feeds         : {', '.join('references/' + r for r in s['refs'])}")
    if rep["uncovered_refs"]:
        print("\n  reference files with no upstream source (repo-original or provenance gap):")
        for r in rep["uncovered_refs"]:
            print(f"    references/{r}")
    print("\nNote: distilled prose is not auto-rewritten. Review the reference files a "
          "stale/changed source feeds, then re-run with --mark-reviewed to stamp the date.")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(
        description="Audit a skill's upstream sources (SOURCES.md) for drift.",
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    p.add_argument("target", nargs="?", help="skill dir or SOURCES.md path")
    p.add_argument("--all", nargs="?", const=".", metavar="ROOT",
                   help="audit every */SOURCES.md under ROOT (default: cwd)")
    p.add_argument("--max-age", type=int, default=180, metavar="DAYS",
                   help="staleness threshold in days (default: 180)")
    p.add_argument("--fetch", action="store_true",
                   help="HTTP-check each URL still resolves")
    p.add_argument("--mark-reviewed", nargs="?", const="", metavar="TITLE",
                   help="stamp 'Last reviewed' to today (optionally only titles containing TITLE)")
    p.add_argument("--json", action="store_true", help="emit a JSON report")
    return p


def collect_targets(args: argparse.Namespace) -> list[Path]:
    if args.all is not None:
        root = Path(args.all)
        if not root.is_dir():
            sys.stderr.write(f"error: --all root not a directory: {root}\n")
            raise SystemExit(2)
        found = sorted(
            p for p in root.rglob("SOURCES.md") if "node_modules" not in p.parts
        )
        if not found:
            sys.stderr.write(f"error: no SOURCES.md found under {root}\n")
            raise SystemExit(2)
        return found
    if not args.target:
        sys.stderr.write("error: provide a skill dir / SOURCES.md path, or --all\n")
        raise SystemExit(2)
    sf = resolve_sources_file(Path(args.target))
    if sf is None:
        sys.stderr.write(f"error: no SOURCES.md at {args.target}\n")
        raise SystemExit(2)
    return [sf]


def main(argv: list[str]) -> int:
    args = build_parser().parse_args(argv)
    today = dt.date.today()
    targets = collect_targets(args)

    # --mark-reviewed is a single-skill write action; refuse it for --all.
    if args.mark_reviewed is not None:
        if len(targets) != 1:
            sys.stderr.write("error: --mark-reviewed operates on a single skill\n")
            return 2
        edited = mark_reviewed(targets[0], args.mark_reviewed or None, today)
        if edited:
            sys.stderr.write(f"stamped 'Last reviewed' = {today.isoformat()} for: "
                             + ", ".join(edited) + "\n")
        else:
            sys.stderr.write("no matching source blocks to stamp\n")
        return 0

    reports = [audit_skill(sf, args.max_age, args.fetch, today) for sf in targets]
    total_problems = sum(r["problems"] for r in reports)

    if args.json:
        out = reports[0] if len(reports) == 1 else reports
        print(json.dumps(out, indent=2))
    else:
        for i, rep in enumerate(reports):
            if i:
                print("\n" + "=" * 60)
            print_text_report(rep, args.max_age)
        if len(reports) > 1:
            stale = sum(1 for r in reports for s in r["sources"] if s["stale"])
            sys.stderr.write(f"\n--- {len(reports)} skills, {total_problems} problem(s), "
                             f"{stale} stale source(s) ---\n")

    return 1 if total_problems else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
