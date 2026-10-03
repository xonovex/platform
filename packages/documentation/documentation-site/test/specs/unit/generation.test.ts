import {existsSync, rmSync} from "node:fs";
import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {createCatalog} from "../../../src/catalog.js";
import {generateSite} from "../../../src/site.js";
import {repository} from "../../support/repository.js";

describe("documentation generation", () => {
  it("preserves authored README text, creates harness installs, and stays idempotent", () => {
    const fixture = repository();
    const first = generateSite(fixture.root, fixture.site, false);
    const readme = fixture.read(fixture.plugin + "/README.md");
    expect(readme).toContain("Authored introduction.");
    expect(readme).toContain("## Usage\n\nKeep this procedure.");
    expect(readme).toContain(
      "claude plugin install xonovex-workflow@xonovex-marketplace",
    );
    expect(readme).toContain(
      "codex plugin add xonovex-workflow@xonovex-marketplace",
    );
    expect(readme).toContain("Plugin dependencies: `xonovex-core`.");
    expect(first.changedReadmes).toEqual([fixture.plugin + "/README.md"]);
    expect(
      generateSite(fixture.root, fixture.site, true).changedReadmes,
    ).toEqual([]);
    expect(fixture.read(fixture.plugin + "/README.md")).toBe(readme);
    const command = fixture.read(
      "packages/documentation/documentation-site/content/commands/xonovex-workflow/plan.md",
    );
    expect(command).toContain("### Claude Code");
    expect(command).not.toContain("### Codex");
  });

  it("fails stale checks without rewriting the README or generated pages", () => {
    const fixture = repository();
    generateSite(fixture.root, fixture.site, false);
    const readme = fixture.read(fixture.plugin + "/README.md");
    const pagePath =
      "packages/documentation/documentation-site/content/plugins/xonovex-workflow.md";
    const page = fixture.read(pagePath);
    fixture.write(
      fixture.plugin + "/plugin.json",
      fixture.read(fixture.plugin + "/plugin.json").replace("1.2.3", "1.2.4"),
    );
    expect(() => generateSite(fixture.root, fixture.site, true)).toThrow(
      "npm run docs:generate",
    );
    expect(fixture.read(fixture.plugin + "/README.md")).toBe(readme);
    expect(fixture.read(pagePath)).toBe(page);
    generateSite(fixture.root, fixture.site, false);
    expect(fixture.read(fixture.plugin + "/README.md")).toContain(
      "bundle version is `1.2.4`",
    );
  });

  it("discovers added components and removes obsolete generated pages", () => {
    const fixture = repository();
    generateSite(fixture.root, fixture.site, false);
    fixture.write(
      fixture.plugin + "/skills/review-guide/SKILL.md",
      "---\nname: review-guide\ndescription: Review a plan.\n---\n# Review guide\n",
    );
    expect(
      generateSite(fixture.root, fixture.site, true).catalog.entries.map(
        (entry) => entry.name,
      ),
    ).toContain("review-guide");
    const generated = join(
      fixture.site,
      "content/skills/xonovex-workflow/review-guide.md",
    );
    expect(existsSync(generated)).toBe(true);
    rmSync(join(fixture.root, fixture.plugin, "skills/review-guide"), {
      recursive: true,
    });
    generateSite(fixture.root, fixture.site, true);
    expect(existsSync(generated)).toBe(false);
  });

  it("copies referenced diagrams and derives dark exports from DESIGN.md", () => {
    const fixture = repository();
    generateSite(fixture.root, fixture.site, false);
    const prefix =
      "packages/documentation/documentation-site/content/public/repository-assets/" +
      fixture.plugin +
      "/diagrams/workflow";
    expect(fixture.read(prefix + ".svg")).toBe(
      fixture.read(fixture.plugin + "/diagrams/workflow.svg"),
    );
    expect(fixture.read(prefix + ".dark.svg")).toContain('fill="#0b0f19"');
    expect(fixture.read(prefix + ".dark.svg")).toContain('fill="#f9fafb"');
    expect(
      fixture.read(
        "packages/documentation/documentation-site/content/theme.css",
      ),
    ).toContain(".dark {\n  --xonovex-canvas: #0b0f19;");
    const publicPath =
      "packages/documentation/documentation-site/content/public/";
    expect(fixture.read(publicPath + "xonovex-logo.svg")).toBe(
      fixture.read("packages/asset/asset-images/xonovex-logo.svg"),
    );
    expect(fixture.read(publicPath + "xonovex-logo.dark.svg")).toContain(
      'fill="#f9fafb"',
    );
    expect(fixture.read(publicPath + "favicon.svg")).toContain(
      "@media (prefers-color-scheme: dark) { g { fill: #f9fafb; } }",
    );
    expect(
      fixture.read(
        "packages/documentation/documentation-site/content/index.md",
      ),
    ).toContain("dark: /xonovex-logo.dark.svg");
  });
});

describe("component discovery", () => {
  it("declares Codex skills and Claude Code commands with their actual harnesses", () => {
    const fixture = repository();
    const catalog = createCatalog(fixture.root);
    expect(
      catalog.entries.find((entry) => entry.type === "skills")?.harnesses,
    ).toEqual(["Claude Code", "Codex"]);
    expect(
      catalog.entries.find((entry) => entry.type === "commands")?.harnesses,
    ).toEqual(["Claude Code"]);
    expect(catalog.documents.map((document) => document.source)).toContain(
      fixture.plugin + "/skills/plan-guide/references/steps.md",
    );
    expect(catalog.entries.some((entry) => entry.type === "agents")).toBe(
      false,
    );
  });

  it("discovers agents, hooks, and MCP servers when manifests declare them", () => {
    const fixture = repository();
    const manifest = {
      name: "xonovex-workflow",
      version: "1.2.3",
      hooks: "./hooks/hooks.json",
      mcpServers: "./.mcp.json",
    };
    fixture.write(
      fixture.plugin + "/.claude-plugin/plugin.json",
      JSON.stringify(manifest),
    );
    fixture.write(
      fixture.plugin + "/.codex-plugin/plugin.json",
      JSON.stringify(manifest),
    );
    fixture.write(
      fixture.plugin + "/agents/researcher.md",
      "---\nname: researcher\ndescription: Research a change.\n---\n# Researcher\n",
    );
    fixture.write(
      fixture.plugin + "/hooks/hooks.json",
      JSON.stringify({
        hooks: {
          SessionStart: [
            {hooks: [{type: "command", command: "check-workspace"}]},
          ],
        },
      }),
    );
    fixture.write(
      fixture.plugin + "/.mcp.json",
      JSON.stringify({mcpServers: {workspace: {command: "workspace-server"}}}),
    );
    const result = generateSite(fixture.root, fixture.site, false);
    expect(
      result.catalog.entries.find((entry) => entry.type === "hooks")?.harnesses,
    ).toEqual(["Claude Code", "Codex"]);
    expect(
      result.catalog.entries.find((entry) => entry.type === "mcp-servers")
        ?.name,
    ).toBe("workspace");
    for (const type of ["agents", "hooks", "mcp-servers"])
      expect(
        existsSync(join(fixture.site, "content/catalog", type + ".md")),
      ).toBe(true);
    expect(
      fixture.read(
        "packages/documentation/documentation-site/content/hooks/xonovex-workflow/SessionStart.md",
      ),
    ).toContain("check-workspace");
  });
});
