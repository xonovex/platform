import {execFileSync} from "node:child_process";
import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {dirname, join, resolve} from "node:path";
import {resolveExecutable} from "@xonovex/script-moon-common/executable";
import {describe, expect, it} from "vitest";

describe("portable evaluation plugin loading", () => {
  it("discovers grouped plugins and stages skills with their resources and dependencies", () => {
    const root = mkdtempSync(join(tmpdir(), "portable-eval-plugin-"));
    const write = (path: string, content: string): void => {
      const target = join(root, path);
      mkdirSync(dirname(target), {recursive: true});
      writeFileSync(target, content);
    };
    try {
      write(
        "plugin-native/.claude-plugin/plugin.json",
        JSON.stringify({name: "xonovex-native", skills: "./skills/"}),
      );
      write("plugin-native/skills/c99-guide/SKILL.md", "# C99");
      write(
        "plugin-core/.claude-plugin/plugin.json",
        JSON.stringify({
          name: "xonovex-core",
          skills: "./skills/",
          dependencies: ["xonovex-native"],
        }),
      );
      write("plugin-core/commands/git-commit.md", "Commit the change.");
      write("plugin-core/skills/alpha-guide/SKILL.md", "# Alpha");
      write(
        "plugin-core/skills/alpha-guide/references/details.md",
        "Alpha details.",
      );
      write("plugin-core/skills/beta-guide/SKILL.md", "# Beta");
      const scripts = resolve(
        import.meta.dirname,
        "../../../../../../runtime/plugin/plugin-agentic/skills/skill-guide/scripts",
      );

      const result = execFileSync(
        resolveExecutable("python3"),
        [
          "-c",
          `
import json
import subprocess
import sys
from pathlib import Path
sys.path.insert(0, sys.argv[1])
from _eval_plugin import find_plugin_directory, read_plugin_manifest, resolve_plugin_directories, staged_plugin_directories
root = Path(sys.argv[2])
core = root / "plugin-core"
native = root / "plugin-native"
assert find_plugin_directory(core / "skills" / "alpha-guide") == core
assert resolve_plugin_directories(core) == [native, core]
with staged_plugin_directories([native, core], "alpha-guide") as staged:
    assert (staged[0] / "skills/c99-guide/SKILL.md").is_file()
    assert (staged[1] / "skills/alpha-guide/references/details.md").read_text() == "Alpha details."
    assert not (staged[1] / "skills/beta-guide").exists()
    assert not (staged[1] / "commands").exists()
    assert json.loads((staged[1] / ".claude-plugin/plugin.json").read_text())["skills"] == ["./skills/alpha-guide"]
assert all(not directory.exists() for directory in staged)
assert (core / "commands/git-commit.md").is_file()
(core / "hooks").mkdir()
try:
    read_plugin_manifest(core)
except ValueError as error:
    assert "contains hooks" in str(error)
else:
    raise AssertionError("grouped plugins must reject active components")
for script in ("eval-triggers.py", "eval-outputs.py"):
    subprocess.run([sys.executable, str(Path(sys.argv[1]) / script), "--help"], check=True, capture_output=True)
print("portable plugin loading passed")
`,
          scripts,
          root,
        ],
        {encoding: "utf8", env: {...process.env, PYTHONDONTWRITEBYTECODE: "1"}},
      );

      expect(result.trim()).toBe("portable plugin loading passed");
    } finally {
      rmSync(root, {recursive: true, force: true});
    }
  });
});
