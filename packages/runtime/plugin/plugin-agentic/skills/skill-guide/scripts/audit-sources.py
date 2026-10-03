#!/usr/bin/env -S uv run --script
# /// script
# requires-python = ">=3.11"
# dependencies = []
# ///
"""Audit Agent Skill source provenance, live URLs, and upstream drift.

This portable implementation mirrors ``moon-skill-audit-sources`` so a shipped
skill and the repository CI enforce the same source schema and exit behavior.
It is read-only unless ``--mark-reviewed`` is supplied.
"""

from __future__ import annotations

import argparse
import datetime as dt
import hashlib
import json
import re
import subprocess
import sys
import urllib.error
import urllib.request
from dataclasses import dataclass, field
from pathlib import Path

HEADER_RE = re.compile(r"^##\s+(.*\S)\s*$")
URL_FIELD_RE = re.compile(r"\*\*URLs?:\*\*")
HTTP_URL_RE = re.compile(r"https?://\S+")
REVIEWED_RE = re.compile(r"\*\*Last reviewed:\*\*\s*(\d{4}-\d{2}-\d{2})")
PROVENANCE_RE = re.compile(r"\*\*Provenance:\*\*\s*(.+\S)\s*$")
REF_RE = re.compile(r"references/([a-z0-9][a-z0-9-]*\.md)")
ARROW_REF_RE = re.compile(r"(?:->|→)\s*`?([a-z0-9][a-z0-9-]*\.md)`?")
REFERENCES_RE = re.compile(r"\*\*References:\*\*\s*(.*\S)?\s*$")
LEGACY_ALL_REFERENCES_RE = re.compile(
    r"\*\*Used for:\*\*.*\b(?:all `references/`|all references|every [^\n]* reference)",
    re.IGNORECASE,
)
CHECKOUT_RE = re.compile(r"\*\*Checkout:\*\*\s*(\S+)")
VERSION_RE = re.compile(r"\*\*Version:\*\*\s*(.+\S)\s*$")
COMMIT_RE = re.compile(r"\*\*Commit:\*\*\s*([0-9a-f]{7,40})")
WATCH_RE = re.compile(r"\*\*Watch:\*\*\s*(.+?)\s*(?:->|→)\s*(.+\S)\s*$")
CONTENT_SHA256_RE = re.compile(
    r"\*\*Content SHA256:\*\*\s*([0-9a-f]{64})", re.IGNORECASE
)


@dataclass(frozen=True)
class Watch:
    path: str
    refs: tuple[str, ...]


@dataclass
class Source:
    title: str
    line_no: int
    urls: list[str] = field(default_factory=list)
    provenance: str | None = None
    reviewed: dt.date | None = None
    reviewed_raw: str | None = None
    refs: set[str] = field(default_factory=set)
    covers_all_references: bool = False
    checkout: str | None = None
    version: str | None = None
    commit: str | None = None
    watches: list[Watch] = field(default_factory=list)
    content_sha256: str | None = None


def find_urls(line: str) -> list[str]:
    return [match.group(0).rstrip(".,;:·") for match in HTTP_URL_RE.finditer(line)]


def parse_reference_field(value: str) -> tuple[bool, tuple[str, ...]]:
    normalized = re.sub(r"^[-:]\s*", "", value.strip())
    if re.fullmatch(r"all(?:\s+references)?", normalized, re.IGNORECASE):
        return True, ()
    refs = tuple(
        Path(part.strip().replace("`", "")).name
        for part in normalized.split(",")
        if re.fullmatch(
            r"[a-z0-9][a-z0-9-]*\.md",
            Path(part.strip().replace("`", "")).name,
        )
    )
    return False, refs


def collect_metadata(source: Source, line: str) -> None:
    if match := PROVENANCE_RE.search(line):
        source.provenance = match.group(1)
    if match := REVIEWED_RE.search(line):
        source.reviewed_raw = match.group(1)
        try:
            source.reviewed = dt.date.fromisoformat(match.group(1))
        except ValueError:
            source.reviewed = None
    if match := CHECKOUT_RE.search(line):
        source.checkout = match.group(1)
    if match := VERSION_RE.search(line):
        value = match.group(1).strip()
        source.version = value[1 : value.find("`", 1)] if value.startswith("`") else value
    if match := COMMIT_RE.search(line):
        source.commit = match.group(1)
    if match := CONTENT_SHA256_RE.search(line):
        source.content_sha256 = match.group(1).lower()
    if match := WATCH_RE.search(line):
        refs = tuple(
            Path(part.strip().replace("`", "")).name
            for part in match.group(2).split(",")
            if Path(part.strip().replace("`", "")).name.endswith(".md")
        )
        source.watches.append(Watch(match.group(1), refs))
        source.refs.update(refs)
    source.refs.update(REF_RE.findall(line))
    source.refs.update(ARROW_REF_RE.findall(line))


def parse_sources(text: str) -> list[Source]:
    sources: list[Source] = []
    current: Source | None = None
    reading_urls = False
    reading_references = False
    for index, line in enumerate(text.splitlines()):
        if header := HEADER_RE.match(line):
            current = Source(header.group(1), index)
            sources.append(current)
            reading_urls = False
            reading_references = False
            continue
        if current is None:
            continue

        if URL_FIELD_RE.search(line):
            current.urls.extend(find_urls(line))
            reading_urls = True
        elif reading_urls and re.match(r"^\s{2,}-\s+", line):
            current.urls.extend(find_urls(line))
        else:
            reading_urls = False

        reference_match = REFERENCES_RE.search(line)
        reference_item = reading_references and re.match(r"^\s{2,}-\s+", line)
        if reference_match or reference_item:
            all_references, refs = parse_reference_field(
                reference_match.group(1) if reference_match and reference_match.group(1) else line
            )
            current.covers_all_references |= all_references
            current.refs.update(refs)
            reading_references = True
        else:
            reading_references = False
        if LEGACY_ALL_REFERENCES_RE.search(line):
            current.covers_all_references = True
        collect_metadata(current, line)

    for source in sources:
        source.urls = list(dict.fromkeys(source.urls))
    return [source for source in sources if source.urls or source.provenance]


def resolve_guide_directory(base: Path) -> Path:
    if (base / "SKILL.md").is_file():
        return base
    nested = sorted(path.parent for path in base.glob("*/SKILL.md") if path.is_file())
    if len(nested) > 1:
        raise ValueError(f"multiple SKILL.md found under {base}; pass one explicitly")
    return nested[0] if nested else base


def resolve_sources_file(target: Path) -> Path | None:
    if target.is_file() and target.name == "SOURCES.md":
        return target
    if target.is_dir():
        direct = target / "SOURCES.md"
        if direct.is_file():
            return direct
        nested = resolve_guide_directory(target) / "SOURCES.md"
        if nested.is_file():
            return nested
    return None


def fetch_status_once(url: str, timeout: int) -> dict[str, str]:
    request = urllib.request.Request(
        url, method="GET", headers={"User-Agent": "skill-source-audit/1"}
    )
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:  # noqa: S310
            code = response.getcode()
            modified = response.headers.get("Last-Modified", "")
            detail = f"HTTP {code}" + (
                f", Last-Modified: {modified}" if modified else ""
            )
            if code == 404:
                return {"status": "missing", "detail": f"HTTP {code}"}
            if code in {401, 403, 429}:
                return {"status": "restricted", "detail": detail}
            if not code or code >= 400:
                return {"status": "error", "detail": detail}
            return {
                "status": "ok",
                "detail": detail,
                "content_sha256": hashlib.sha256(response.read()).hexdigest(),
            }
    except urllib.error.HTTPError as error:
        if error.code == 404:
            return {"status": "missing", "detail": "HTTP 404"}
        if error.code in {401, 403, 429}:
            return {"status": "restricted", "detail": f"HTTP {error.code}"}
        return {"status": "error", "detail": f"HTTP {error.code}"}
    except (urllib.error.URLError, TimeoutError, ValueError) as error:
        return {"status": "error", "detail": f"unreachable: {error}"}


def fetch_status(url: str, timeout: int = 15) -> dict[str, str]:
    report = fetch_status_once(url, timeout)
    if report["status"] == "error":
        report = fetch_status_once(url, timeout)
    return report


def aggregate_content_sha256(urls: list[str], fetches: list[dict[str, str]]) -> str | None:
    if len(urls) != len(fetches) or any(
        fetch.get("status") != "ok" or "content_sha256" not in fetch
        for fetch in fetches
    ):
        return None
    digest = hashlib.sha256()
    for url, fetch in zip(urls, fetches, strict=True):
        digest.update(url.encode())
        digest.update(b"\0")
        digest.update(fetch["content_sha256"].encode())
        digest.update(b"\n")
    return digest.hexdigest()


def find_workspace_root(start: Path) -> Path:
    directory = start.resolve()
    while directory.parent != directory:
        if (directory / ".moon").is_dir():
            return directory
        directory = directory.parent
    return start.resolve()


def git(directory: Path, *arguments: str) -> str | None:
    try:
        result = subprocess.run(
            ["git", *arguments],
            cwd=directory,
            check=True,
            capture_output=True,
            text=True,
        )
    except (OSError, subprocess.CalledProcessError):
        return None
    return result.stdout


def semver(value: str) -> tuple[int, int, int] | None:
    match = re.search(r"(\d+)\.(\d+)\.(\d+)", value)
    if match is None:
        return None
    return (int(match.group(1)), int(match.group(2)), int(match.group(3)))


def latest_semver_tag(checkout: Path) -> tuple[tuple[int, int, int], str] | None:
    output = git(checkout, "tag", "--list")
    if output is None:
        return None
    tags = [tag for tag in output.splitlines() if re.fullmatch(r"v?\d+\.\d+\.\d+", tag)]
    parsed = [(version, tag) for tag in tags if (version := semver(tag)) is not None]
    return max(parsed, default=None)


def compute_drift(workspace_root: Path, source: Source, pull: bool) -> dict:
    checkout = (workspace_root / (source.checkout or "")).resolve()
    report = {
        "checkout": str(checkout),
        "resolved": False,
        "pull_failed": False,
        "pinned_version": source.version,
        "latest_version": None,
        "behind": False,
        "commit_count": 0,
        "commits": [],
        "review_refs": [],
    }
    if not checkout.is_dir() or not (checkout / ".git").is_dir():
        return {**report, "note": "checkout not found or not a git repo"}
    report["resolved"] = True
    if pull and git(checkout, "fetch", "--tags", "--quiet") is None:
        return {**report, "pull_failed": True, "note": "failed to fetch upstream tags"}
    latest = latest_semver_tag(checkout)
    pinned_version = semver(source.version or "")
    if latest:
        report["latest_version"] = latest[1]
        report["behind"] = pinned_version is not None and latest[0] > pinned_version
    reference = source.commit or source.version
    if reference is None:
        return {**report, "note": "no **Commit:** or **Version:** to diff from"}
    if git(checkout, "rev-parse", "--verify", "--quiet", f"{reference}^{{commit}}") is None:
        return {**report, "note": f"pinned ref '{reference}' not found in checkout"}
    upper = latest[1] if latest else "HEAD"
    paths = [watch.path for watch in source.watches]
    range_value = f"{reference}..{upper}"
    scoped_paths = ["--", *paths] if paths else []
    log = git(checkout, "log", "--oneline", range_value, *scoped_paths) or ""
    commits = [line for line in log.splitlines() if line.strip()]
    report["commit_count"] = len(commits)
    report["commits"] = commits[:20]
    changed = (
        git(checkout, "diff", "--name-only", range_value, *scoped_paths) or ""
    )
    review_refs = {
        reference
        for filename in changed.splitlines()
        for watch in source.watches
        if filename.strip().startswith(watch.path)
        for reference in watch.refs
    }
    report["review_refs"] = sorted(review_refs)
    return report


def audit_skill(
    sources_file: Path,
    max_age: int,
    version_max_age: int,
    do_fetch: bool,
    pull: bool,
    today: dt.date,
) -> dict:
    skill_dir = sources_file.parent
    workspace_root = find_workspace_root(skill_dir)
    sources = parse_sources(sources_file.read_text(encoding="utf-8"))
    references_dir = skill_dir / "references"
    existing_refs = (
        {path.name for path in references_dir.glob("*.md")}
        if references_dir.is_dir()
        else set()
    )
    covered: set[str] = set()
    reports: list[dict] = []
    problems = 0
    for source in sources:
        age = (today - source.reviewed).days if source.reviewed else None
        review_max_age = version_max_age if source.version else max_age
        stale = age is not None and age > review_max_age
        dangling = sorted(source.refs - existing_refs)
        reference_mapping_missing = not (
            source.covers_all_references or source.refs
        )
        drift_anchor_missing = bool(
            source.version
            and source.urls
            and not source.content_sha256
            and not (source.checkout and source.commit and source.watches)
        )
        covered.update(existing_refs if source.covers_all_references else source.refs)
        report = {
            "title": source.title,
            "url": source.urls[0] if source.urls else None,
            "urls": source.urls,
            "provenance": source.provenance,
            "last_reviewed": source.reviewed_raw,
            "age_days": age,
            "review_max_age_days": review_max_age,
            "stale": stale,
            "refs": sorted(source.refs),
            "covers_all_references": source.covers_all_references,
            "reference_mapping_missing": reference_mapping_missing,
            "dangling_refs": dangling,
            "version": source.version,
            "commit": source.commit,
            "watch_count": len(source.watches),
            "content_sha256": source.content_sha256,
            "fetched_content_sha256": None,
            "content_changed": False,
            "drift_anchor_missing": drift_anchor_missing,
        }
        if age is None or stale or dangling:
            problems += 1
        if reference_mapping_missing:
            problems += 1
        if drift_anchor_missing:
            problems += 1
        if do_fetch and source.urls:
            fetches = [fetch_status(url) for url in source.urls]
            report["fetch"] = fetches[0]
            report["fetches"] = fetches
            problems += sum(
                fetch["status"] not in {"ok", "restricted"} for fetch in fetches
            )
            fetched_hash = aggregate_content_sha256(source.urls, fetches)
            report["fetched_content_sha256"] = fetched_hash
            if source.content_sha256:
                if fetched_hash is None:
                    problems += 1
                elif fetched_hash != source.content_sha256:
                    report["content_changed"] = True
                    problems += 1
        if source.checkout:
            drift = compute_drift(workspace_root, source, pull)
            report["drift"] = drift
            if drift["pull_failed"] or drift["behind"] or drift["commit_count"] > 0:
                problems += 1
        reports.append(report)

    uncovered = sorted(existing_refs - covered)
    problems += len(uncovered)
    if not sources:
        problems += 1
    return {
        "skill": skill_dir.resolve().name,
        "sources_file": str(sources_file),
        "source_count": len(sources),
        "sources": reports,
        "uncovered_refs": uncovered,
        "problems": problems,
    }


def mark_reviewed(
    sources_file: Path, title_filter: str | None, today: dt.date
) -> list[str]:
    lines = sources_file.read_text(encoding="utf-8").splitlines(keepends=True)
    needle = title_filter.lower() if title_filter else None
    edited: list[str] = []
    title = ""
    matched = False
    for index, line in enumerate(lines):
        if header := HEADER_RE.match(line.rstrip("\n")):
            title = header.group(1)
            matched = needle is None or needle in title.lower()
        elif matched and REVIEWED_RE.search(line):
            lines[index] = REVIEWED_RE.sub(
                f"**Last reviewed:** {today.isoformat()}", line
            )
            if title and title not in edited:
                edited.append(title)
    if edited:
        sources_file.write_text("".join(lines), encoding="utf-8")
    return edited


def print_text_report(report: dict, max_age: int, version_max_age: int) -> None:
    print(f"skill: {report['skill']}  ({report['sources_file']})")
    print(
        f"sources: {report['source_count']}  max-age: {max_age}d  "
        f"version-max-age: {version_max_age}d"
    )
    for source in report["sources"]:
        flags: list[str] = []
        if source["stale"]:
            flags.append(f"STALE ({source['age_days']}d)")
        if source["reference_mapping_missing"]:
            flags.append("MISSING REFERENCE MAPPING")
        if source["drift_anchor_missing"]:
            flags.append("MISSING DRIFT ANCHOR")
        if source["content_changed"]:
            flags.append("CONTENT CHANGED")
        if source["dangling_refs"]:
            flags.append(f"DANGLING: {', '.join(source['dangling_refs'])}")
        for fetch in source.get("fetches", []):
            if fetch["status"] != "ok":
                flags.append(f"URL {fetch['status']}")
                break
        print(f"\n  [{'  '.join(flags) if flags else 'ok'}] {source['title']}")
        print(f"    urls          : {', '.join(source['urls']) or '(none)'}")
        print(f"    last reviewed : {source['last_reviewed'] or '(none)'}")
        print(
            "    feeds         : "
            + (
                "all references"
                if source["covers_all_references"]
                else ", ".join(f"references/{ref}" for ref in source["refs"])
                or "(missing)"
            )
        )


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Audit a skill's upstream sources (SOURCES.md) for drift."
    )
    parser.add_argument("target", nargs="?", help="skill dir or SOURCES.md path")
    parser.add_argument("--all", nargs="?", const=".", metavar="ROOT")
    parser.add_argument("--max-age", type=int, default=180, metavar="DAYS")
    parser.add_argument("--version-max-age", type=int, default=90, metavar="DAYS")
    parser.add_argument("--fetch", action="store_true")
    parser.add_argument("--pull", action="store_true")
    parser.add_argument("--mark-reviewed", nargs="?", const="", metavar="TITLE")
    parser.add_argument("--json", action="store_true")
    return parser


def collect_targets(args: argparse.Namespace) -> list[Path]:
    if args.all is not None:
        root = Path(args.all)
        if not root.is_dir():
            raise ValueError(f"--all root not a directory: {root}")
        found = sorted(
            path
            for path in root.rglob("SOURCES.md")
            if "node_modules" not in path.parts and (path.parent / "SKILL.md").is_file()
        )
        if not found:
            raise ValueError(f"no SOURCES.md found under {root}")
        return found
    target = Path(args.target or Path.cwd())
    sources_file = resolve_sources_file(target)
    if sources_file is None:
        raise ValueError(f"no SOURCES.md at {target}")
    return [sources_file]


def main(argv: list[str]) -> int:
    args = build_parser().parse_args(argv)
    try:
        targets = collect_targets(args)
    except ValueError as error:
        sys.stderr.write(f"error: {error}\n")
        return 2
    today = dt.date.today()
    if args.mark_reviewed is not None:
        if len(targets) != 1:
            sys.stderr.write("error: --mark-reviewed operates on a single skill\n")
            return 2
        edited = mark_reviewed(targets[0], args.mark_reviewed or None, today)
        message = (
            f"stamped 'Last reviewed' = {today.isoformat()} for: {', '.join(edited)}\n"
            if edited
            else "no matching source blocks to stamp\n"
        )
        sys.stderr.write(message)
        return 0

    reports = [
        audit_skill(
            source,
            args.max_age,
            args.version_max_age,
            args.fetch,
            args.pull,
            today,
        )
        for source in targets
    ]
    total_problems = sum(report["problems"] for report in reports)
    if args.json:
        print(json.dumps(reports[0] if len(reports) == 1 else reports, indent=2))
    else:
        for index, report in enumerate(reports):
            if index:
                print("\n" + "=" * 60)
            print_text_report(report, args.max_age, args.version_max_age)
    return 1 if total_problems else 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
