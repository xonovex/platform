"""Discover and stage skill-only Claude plugins for portable evaluations."""

from __future__ import annotations

import json
import shutil
from collections.abc import Iterator, Sequence
from contextlib import contextmanager
from pathlib import Path
from tempfile import TemporaryDirectory

PLUGIN_FIELDS = {"author", "dependencies", "description", "name", "skills", "version"}
EXECUTABLE_COMPONENTS = (
    "commands", "agents", "hooks", ".mcp.json", ".lsp.json", "settings.json",
    ".claude/settings.json",
)


def read_plugin_manifest(directory: Path) -> tuple[str, list[str], bool]:
    manifest_path = directory / ".claude-plugin" / "plugin.json"
    try:
        value = json.loads(manifest_path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError) as error:
        raise ValueError(f"invalid Claude plugin manifest: {manifest_path}") from error
    if not isinstance(value, dict) or not isinstance(value.get("name"), str) or not value["name"]:
        raise ValueError(f"Claude plugin manifest has no name: {manifest_path}")
    unsupported = sorted(set(value) - PLUGIN_FIELDS)
    if unsupported:
        raise ValueError(
            f"Claude eval plugin is not skill-only: {manifest_path} has {unsupported[0]}"
        )
    skills = value.get("skills")
    grouped = skills == "./skills/"
    if not grouped and (not isinstance(skills, list) or not skills or not all(
        isinstance(skill, str) for skill in skills
    )):
        raise ValueError(f"Claude plugin skills are invalid: {manifest_path}")
    dependencies = value.get("dependencies", [])
    if not isinstance(dependencies, list) or not all(
        isinstance(dependency, str) for dependency in dependencies
    ):
        raise ValueError(f"Claude plugin dependencies are invalid: {manifest_path}")
    executable = next(
        (entry for entry in EXECUTABLE_COMPONENTS
         if not (grouped and entry == "commands") and (directory / entry).exists()),
        None,
    )
    if executable is not None:
        raise ValueError(
            f"Claude eval plugin is not skill-only: {directory} contains {executable}"
        )
    return value["name"], dependencies, grouped


def find_plugin_directory(guide: Path) -> Path:
    grouped = guide.parent.parent
    return grouped if (grouped / ".claude-plugin" / "plugin.json").is_file() else guide.parent


def resolve_plugin_directories(target: Path) -> list[Path]:
    target = target.resolve()
    directories: dict[str, Path] = {}
    for directory in target.parent.iterdir():
        if not (directory / ".claude-plugin" / "plugin.json").is_file():
            continue
        name, _, _ = read_plugin_manifest(directory)
        directories[name] = directory

    ordered: list[Path] = []
    visited: set[str] = set()

    def visit(directory: Path) -> None:
        name, dependencies, _ = read_plugin_manifest(directory)
        if name in visited:
            return
        visited.add(name)
        for dependency in dependencies:
            dependency_directory = directories.get(dependency)
            if dependency_directory is None:
                raise ValueError(
                    f"local Claude plugin dependency not found: {name} -> {dependency}"
                )
            visit(dependency_directory)
        ordered.append(directory)

    visit(target)
    return ordered


@contextmanager
def staged_plugin_directories(
    directories: Sequence[Path], target_skill: str,
) -> Iterator[list[Path]]:
    with TemporaryDirectory(prefix="xonovex-claude-eval-") as workspace:
        staged: list[Path] = []
        for index, directory in enumerate(directories):
            name, _, grouped = read_plugin_manifest(directory)
            if not grouped:
                staged.append(directory)
                continue
            guides = sorted(
                guide for guide in (directory / "skills").iterdir()
                if (guide / "SKILL.md").is_file()
            )
            selected = [guide for guide in guides if guide.name == target_skill]
            guides = selected or guides
            target = Path(workspace) / f"plugin-{index}"
            manifest = target / ".claude-plugin" / "plugin.json"
            manifest.parent.mkdir(parents=True)
            for guide in guides:
                shutil.copytree(guide, target / "skills" / guide.name)
            manifest.write_text(json.dumps({
                "name": name,
                "skills": [f"./skills/{guide.name}" for guide in guides],
            }), encoding="utf-8")
            staged.append(target)
        yield staged
