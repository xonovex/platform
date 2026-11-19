"""Shared fixtures for the plan-bundle unit tests: the script module, in-memory bundles and schemas."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import json
import sys
from collections.abc import Buffer
from pathlib import Path
from types import ModuleType
from typing import Any

SCRIPT = Path(__file__).resolve().parents[3] / "skills" / "recording-handoff" / "scripts" / "plan-bundle.py"


class _NoBytecodeLoader(importlib.machinery.SourceFileLoader):
    """Loads the script without writing a __pycache__ directory into the skill."""

    def set_data(self, path: str, data: Buffer, *, _mode: int = 0o666) -> None:
        return None


def _load_script() -> ModuleType:
    name = "plan_bundle"
    if name in sys.modules:
        return sys.modules[name]
    loader = _NoBytecodeLoader(name, str(SCRIPT))
    spec = importlib.util.spec_from_file_location(name, SCRIPT, loader=loader)
    assert spec is not None
    module = importlib.util.module_from_spec(spec)
    # dataclasses resolve annotations through sys.modules, so the module must be registered.
    sys.modules[name] = module
    loader.exec_module(module)
    return module


pb = _load_script()

AT = "2026-10-28T09:00:00Z"

COMMON = {
    "$id": "workflow/common",
    "$defs": {
        "actor": {
            "type": "string",
            "pattern": "^(human|process):[A-Za-z0-9][A-Za-z0-9._-]*$|^[A-Za-z][A-Za-z0-9._-]*/[A-Za-z0-9][A-Za-z0-9._+-]*$",
        },
        "timestamp": {"type": "string", "format": "datetime-offset"},
        "event": {
            "type": "object",
            "required": ["by", "at"],
            "properties": {"by": {"$ref": "#/$defs/actor"}, "at": {"$ref": "#/$defs/timestamp"}},
        },
        "status": {"type": "string", "enum": ["draft", "stable", "deprecated"]},
        "kebabName": {"type": "string", "pattern": "^[a-z0-9]+(-[a-z0-9]+)*$"},
        "source": {
            "type": "object",
            "required": ["resource"],
            "properties": {"resource": {"type": "string", "minLength": 1}, "id": {"$ref": "#/$defs/kebabName"}},
        },
        "conceptId": {"type": "string", "pattern": "^[a-z0-9][a-z0-9-]*(/[a-z0-9][a-z0-9-]*)*$"},
        "conceptIds": {
            "type": "array",
            "uniqueItems": True,
            "items": {"allOf": [{"$ref": "#/$defs/conceptId"}, {"x-concept": ""}]},
        },
        "decisionId": {"allOf": [{"$ref": "#/$defs/conceptId"}, {"x-concept": "decisions/"}]},
        "decisionIds": {"type": "array", "uniqueItems": True, "items": {"$ref": "#/$defs/decisionId"}},
        "workId": {"allOf": [{"$ref": "#/$defs/conceptId"}, {"x-concept": "work/"}]},
        "workIds": {"type": "array", "uniqueItems": True, "items": {"$ref": "#/$defs/workId"}},
        "header": {
            "type": "object",
            "required": ["type", "schema", "title", "description", "generated"],
            "properties": {
                "title": {"type": "string", "minLength": 1},
                "description": {"type": "string", "minLength": 1},
                "generated": {"$ref": "#/$defs/event"},
                "verified": {
                    "oneOf": [
                        {"$ref": "#/$defs/event"},
                        {"type": "array", "minItems": 1, "items": {"$ref": "#/$defs/event"}},
                    ]
                },
                "status": {"$ref": "#/$defs/status"},
                "sources": {"type": "array", "items": {"$ref": "#/$defs/source"}},
                "tags": {"type": "array", "uniqueItems": True, "items": {"type": "string", "minLength": 1}},
            },
        },
    },
}

DECISION = {
    "type": "object",
    "required": ["type", "schema", "state"],
    "properties": {
        "type": {"const": "Workflow Decision"},
        "schema": {"const": "workflow/decision/1"},
        "state": {"enum": ["settled", "open", "paused", "partly-settled"]},
        "depends_on": {"$ref": "common.schema.json#/$defs/decisionIds"},
        "supersedes": {"$ref": "common.schema.json#/$defs/decisionIds"},
        "superseded_by": {"$ref": "common.schema.json#/$defs/decisionId"},
    },
    "allOf": [{"$ref": "common.schema.json#/$defs/header"}],
}

WORK_ITEM = {
    "type": "object",
    "required": ["type", "schema", "kind", "team", "readiness"],
    "properties": {
        "type": {"const": "Workflow Work Item"},
        "schema": {"const": "workflow/work-item/1"},
        "kind": {"enum": ["work-package", "work-item"]},
        "parent": {"$ref": "common.schema.json#/$defs/workId"},
        "team": {"type": "string", "minLength": 1},
        "readiness": {"enum": ["ready", "waiting"]},
        "waits_on": {"$ref": "common.schema.json#/$defs/decisionIds"},
        "depends_on": {"$ref": "common.schema.json#/$defs/workIds"},
        "criteria": {
            "type": "array",
            "items": {
                "type": "object",
                "required": ["id", "text"],
                "properties": {"id": {"type": "string", "pattern": "^c[0-9]+$"}, "text": {"type": "string"}},
            },
        },
    },
    "allOf": [{"$ref": "common.schema.json#/$defs/header"}],
}

LOG_ENTRY = {
    "type": "object",
    "required": ["type", "schema", "skill", "action"],
    "properties": {
        "type": {"const": "Workflow Log Entry"},
        "schema": {"const": "workflow/log-entry/1"},
        "skill": {"$ref": "common.schema.json#/$defs/kebabName"},
        "action": {"enum": ["Creation", "Update", "Deprecation", "Initialization"]},
        "touched": {"$ref": "common.schema.json#/$defs/conceptIds"},
    },
    "allOf": [{"$ref": "common.schema.json#/$defs/header"}],
}

PLAN = {
    "type": "object",
    "required": ["type", "schema"],
    "properties": {"type": {"const": "Workflow Plan"}, "schema": {"const": "workflow/plan/1"}},
    "allOf": [{"$ref": "common.schema.json#/$defs/header"}],
}

SCHEMA_DIRECTORY = "/skills/recording-handoff/assets/schemas"

SHARED_SCHEMAS = {
    "common.schema.json": COMMON,
    "decision.schema.json": DECISION,
    "work-item.schema.json": WORK_ITEM,
    "log-entry.schema.json": LOG_ENTRY,
    "plan.schema.json": PLAN,
}


def schema_sources(
    schemas: dict[str, object] | None = None, directory: str = SCHEMA_DIRECTORY, tier: int = 1
) -> tuple:
    """Schema sources from documents (dumped as JSON) or raw text, in name order."""
    chosen = SHARED_SCHEMAS if schemas is None else schemas
    return tuple(
        pb.SchemaSource(
            f"{directory}/{name}",
            value if isinstance(value, str) else json.dumps(value, indent=2),
            tier,
        )
        for name, value in sorted(chosen.items())
    )


def registry(*sources: tuple) -> Any:
    flat = [source for group in (sources or (schema_sources(),)) for source in group]
    return pb.build_schema_registry(flat)


def doc(front: str, body: str = "Body.\n") -> str:
    """A markdown file with the given frontmatter lines."""
    return f"---\n{front.strip(chr(10))}\n---\n{body}"


def header(kind: str, schema: str, title: str = "Title", at: str = AT, extra: str = "") -> str:
    return (
        f"type: {kind}\nschema: {schema}\ntitle: {title}\ndescription: A description.\n"
        f"generated: {{by: tester/1.0.0, at: {at}}}\n{extra}"
    )


def plan(extra: str = "", body: str = "# Objective\n\nShip it.\n", at: str = AT) -> str:
    return doc(header("Workflow Plan", "workflow/plan/1", "The plan", at, extra), body)


def decision(state: str = "settled", title: str = "A decision", at: str = AT, extra: str = "") -> str:
    return doc(header("Workflow Decision", "workflow/decision/1", title, at, f"state: {state}\n{extra}"))


def work(
    title: str = "A work item", readiness: str = "ready", at: str = AT, extra: str = "", kind: str = "work-item"
) -> str:
    return doc(
        header(
            "Workflow Work Item",
            "workflow/work-item/1",
            title,
            at,
            f"kind: {kind}\nteam: Team\nreadiness: {readiness}\n{extra}",
        )
    )


def log_entry(summary: str = "Did a thing.", at: str = AT, action: str = "Creation", extra: str = "") -> str:
    return doc(
        header("Workflow Log Entry", "workflow/log-entry/1", "Entry", at, f"skill: writing-plan\naction: {action}\n{extra}"),
        f"\n{summary}\n",
    )


def bundle(files: dict[str, str], directories: tuple[str, ...] = (), links: tuple[str, ...] = ()) -> Any:
    return pb.load_bundle(pb.BundleSource(tuple(files.items()), directories, links))


def with_generated(files: dict[str, str]) -> dict[str, str]:
    """Adds the current index.md and log.md files to a set of bundle files."""
    loaded = bundle(files)
    generated = pb.generate_indexes(loaded)
    log = pb.generate_log(loaded)
    assert not generated.findings and not log.findings, (generated.findings, log.findings)
    return {**files, **dict(generated.files), **dict(log.files)}


def messages(findings, path: str | None = None) -> list[str]:
    return [f.message for f in findings if path is None or f.path == path]


def lines(findings, message: str) -> list[tuple[str, int, str]]:
    """The path, line and severity of the findings that hold message."""
    return [(f.path, f.line, f.severity) for f in findings if message in f.message]
