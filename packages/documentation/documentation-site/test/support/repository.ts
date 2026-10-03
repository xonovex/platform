import {
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import {tmpdir} from "node:os";
import {dirname, join} from "node:path";
import {onTestFinished} from "vitest";

const colors = [
  ["canvas", "#ffffff", "#0b0f19"],
  ["surface", "#f9fafb", "#111827"],
  ["text", "#000000", "#f9fafb"],
  ["secondary", "#374151", "#d1d5db"],
  ["muted", "#6b7280", "#9ca3af"],
  ["border", "#d1d5db", "#374151"],
  ["strong", "#111827", "#f9fafb"],
  ["blue", "#3b82f6", "#60a5fa"],
  ["blue-strong", "#2563eb", "#93c5fd"],
  ["blue-soft", "#eff6ff", "#172554"],
  ["green", "#22c55e", "#4ade80"],
  ["green-strong", "#15803d", "#86efac"],
  ["green-soft", "#f0fdf4", "#052e16"],
] as const;

export const repository = (): {
  readonly root: string;
  readonly site: string;
  readonly plugin: string;
  readonly write: (path: string, content: string) => void;
  readonly read: (path: string) => string;
} => {
  const root = mkdtempSync(join(tmpdir(), "xonovex-docsite-"));
  onTestFinished(() => {
    rmSync(root, {recursive: true, force: true});
  });
  const write = (path: string, content: string): void => {
    const destination = join(root, path);
    mkdirSync(dirname(destination), {recursive: true});
    writeFileSync(destination, content);
  };
  const read = (path: string): string => readFileSync(join(root, path), "utf8");
  const plugin = "packages/runtime/plugin/plugin-workflow";
  const manifest = {
    name: "xonovex-workflow",
    version: "1.2.3",
    description: "Research and deliver changes.",
    skills: "./skills/",
    dependencies: ["xonovex-core"],
  };
  const marketplace = JSON.stringify({
    name: "xonovex-marketplace",
    plugins: [{name: manifest.name, source: "./" + plugin}],
  });
  write(".claude-plugin/marketplace.json", marketplace);
  write(".agents/plugins/marketplace.json", marketplace);
  write(plugin + "/plugin.json", JSON.stringify(manifest));
  write(plugin + "/.claude-plugin/plugin.json", JSON.stringify(manifest));
  write(plugin + "/.codex-plugin/plugin.json", JSON.stringify(manifest));
  write(
    plugin + "/README.md",
    "# Workflow\n\nAuthored introduction.\n\n## Skills\n\n[Plan](skills/plan-guide/SKILL.md)\n\n![Steps](diagrams/workflow.svg)\n\n## Usage\n\nKeep this procedure.\n",
  );
  write(
    plugin + "/skills/plan-guide/SKILL.md",
    "---\nname: plan-guide\ndescription: Research and plan a change.\n---\n# Plan guide\n\nRead the [steps](references/steps.md).\n",
  );
  write(
    plugin + "/skills/plan-guide/references/steps.md",
    "# Plan steps\n\nUse <model> and {{ task }} as literal placeholders.\n",
  );
  write(
    plugin + "/commands/plan.md",
    "---\ndescription: Create a researched plan.\n---\n# Plan command\n\nUse the planning skill.\n",
  );
  write(
    plugin + "/diagrams/workflow.svg",
    '<svg xmlns="http://www.w3.org/2000/svg"><rect fill="white" stroke="#d1d5db"/><text fill="black">Steps</text></svg>',
  );
  write(
    "README.md",
    "# Platform\n\nUse [Workflow](" + plugin + "/README.md).\n",
  );
  write(
    "packages/asset/asset-images/xonovex-logo.svg",
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 559 838"><title>Xonovex</title><g fill="#000000"><path d="M0 0h559v838z"/></g></svg>\n',
  );
  write(
    "DESIGN.md",
    "# Design\n\n| Role | Token | Light | Dark | Use |\n| --- | --- | --- | --- | --- |\n" +
      colors
        .map(
          ([token, light, dark]) =>
            "| Role | `" +
            token +
            "` | `" +
            light +
            "` | `" +
            dark +
            "` | Use |\n",
        )
        .join(""),
  );
  return {
    root,
    site: join(root, "packages/documentation/documentation-site"),
    plugin,
    write,
    read,
  };
};
