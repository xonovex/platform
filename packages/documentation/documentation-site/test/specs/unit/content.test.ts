import MarkdownIt from "markdown-it";
import {describe, expect, it} from "vitest";
import {createCatalog} from "../../../src/catalog.js";
import {inside} from "../../../src/files.js";
import {filterComponents} from "../../../src/filter.js";
import {replaceInstallation} from "../../../src/install.js";
import {resolveLink} from "../../../src/links.js";
import {repositoryMarkdown} from "../../../src/markdown-renderer.js";
import {designPalette} from "../../../src/palette.js";
import {repository} from "../../support/repository.js";

describe("installation ownership", () => {
  it("replaces a legacy install section while preserving the following procedure", () => {
    const before =
      "# Plugin\n\nIntro.\n\n## Installation\n\nOld instructions.\n\n## Usage\n\nKeep this.\n";
    const updated = replaceInstallation(
      before,
      "## Install\n\nNew instructions.",
    );
    expect(updated).not.toContain("Old instructions.");
    expect(updated).toContain("Intro.");
    expect(updated).toContain("## Usage\n\nKeep this.");
    expect(
      replaceInstallation(updated, "## Install\n\nNew instructions."),
    ).toBe(updated);
  });

  it.each([
    "<!-- xonovex:installation:start -->",
    "<!-- xonovex:installation:end -->\n<!-- xonovex:installation:start -->",
    "<!-- xonovex:installation:start -->\n<!-- xonovex:installation:start -->\n<!-- xonovex:installation:end -->",
  ])("rejects an incomplete or duplicated marker block", (text) => {
    expect(() => replaceInstallation(text, "## Install")).toThrow(
      "one complete block",
    );
  });
});

describe("source links and Markdown", () => {
  it("hides installation markers while preserving instructions and literal code examples", () => {
    const fixture = repository();
    const md = new MarkdownIt({html: false});
    repositoryMarkdown(md, fixture.root, createCatalog(fixture.root));
    const marker = "<!-- xonovex:installation:start -->";
    const end = "<!-- xonovex:installation:end -->";
    const rendered = md.render(
      "# Plugin\n\n" +
        marker +
        "\n## Install\n\nKeep these instructions.\n" +
        end,
    );
    expect(rendered).toContain("<h2>Install</h2>");
    expect(rendered).toContain("Keep these instructions.");
    expect(rendered).not.toContain("xonovex:installation");
    expect(rendered).not.toContain("<p></p>");
    const examples = md.render(
      "`" + marker + "`\n\n```markdown\n" + marker + "\n```\n\n    " + end,
    );
    expect(examples.match(/xonovex:installation/gu)).toHaveLength(3);
    expect(md.render(marker)).toBe("");
  });

  it("routes documents and assets while retaining external links and source links", () => {
    const fixture = repository();
    const catalog = createCatalog(fixture.root);
    expect(
      resolveLink(
        fixture.root,
        catalog,
        "README.md",
        fixture.plugin + "/README.md#skills",
      ).href,
    ).toBe("/plugins/xonovex-workflow.html#skills");
    expect(
      resolveLink(
        fixture.root,
        catalog,
        fixture.plugin + "/README.md",
        "diagrams/workflow.svg",
      ).dark,
    ).toContain("workflow.dark.svg");
    expect(
      resolveLink(
        fixture.root,
        catalog,
        "README.md",
        fixture.plugin + "/plugin.json",
      ).href,
    ).toBe(
      catalog.repository + "/blob/main/" + fixture.plugin + "/plugin.json",
    );
    expect(
      resolveLink(
        fixture.root,
        catalog,
        "README.md",
        "https://example.com/guide?mode=dark#steps",
      ).href,
    ).toBe("https://example.com/guide?mode=dark#steps");
    expect(() => inside(fixture.root, "../outside.md")).toThrow("escapes");
  });

  it("rewrites rendered links, leaves fenced examples intact, and escapes Vue placeholders", () => {
    const fixture = repository();
    const md = new MarkdownIt({html: false});
    repositoryMarkdown(md, fixture.root, createCatalog(fixture.root));
    const rendered = md.render(
      "[Plan](skills/plan-guide/SKILL.md)\n\n![Steps](diagrams/workflow.svg)\n\nUse <model> and {{ task }} and `{{args}}`.\n\n```{language}\n[Plan](skills/plan-guide/SKILL.md)\n```",
      {frontmatter: {sourcePath: fixture.plugin + "/README.md"}},
    );
    expect(rendered).toContain(
      'href="/skills/xonovex-workflow/plan-guide.html"',
    );
    expect(rendered).toContain("<ThemeDiagram");
    expect(rendered).toContain("workflow.dark.svg");
    expect(rendered).toContain("[Plan](skills/plan-guide/SKILL.md)");
    expect(rendered).toContain("&lt;model&gt;");
    expect(rendered).toContain("&#123;&#123; task &#125;&#125;");
    expect(rendered).toContain('<code v-pre="">{{args}}</code>');
    expect(rendered).toContain('class="language-text"');
  });

  it("serves Graphviz sources as downloads under a host prefix", () => {
    const fixture = repository();
    fixture.write(
      fixture.plugin + "/diagrams/workflow.dot",
      "digraph workflow {}\n",
    );
    const md = new MarkdownIt();
    repositoryMarkdown(
      md,
      fixture.root,
      createCatalog(fixture.root),
      "/platform/",
    );
    const rendered = md.render("[Source](diagrams/workflow.dot)", {
      frontmatter: {sourcePath: fixture.plugin + "/README.md"},
    });
    expect(rendered).toContain('download="workflow.dot"');
    expect(rendered).toContain('href="/platform/repository-assets/');
    expect(rendered).not.toContain(".dot.html");
  });
});

describe("catalog filtering and theme validation", () => {
  it("combines keywords, ownership, and supported harness filters", () => {
    const fixture = repository();
    const entries = createCatalog(fixture.root).entries;
    expect(
      filterComponents(entries, {
        type: "skills",
        query: "  PLAN research ",
        plugin: "xonovex-workflow",
        harness: "Codex",
      }).map((entry) => entry.name),
    ).toEqual(["plan-guide"]);
    expect(
      filterComponents(entries, {
        type: "commands",
        query: "",
        plugin: "",
        harness: "Codex",
      }),
    ).toEqual([]);
    expect(
      filterComponents(entries, {
        type: "skills",
        query: "unknown",
        plugin: "",
        harness: "",
      }),
    ).toEqual([]);
  });

  it("rejects missing colors in the shared light and dark design contract", () => {
    const fixture = repository();
    expect(designPalette(fixture.read("DESIGN.md"))).toHaveLength(13);
    expect(() =>
      designPalette(
        fixture.read("DESIGN.md").replaceAll(/^\| Role \| `canvas`.*\n/gmu, ""),
      ),
    ).toThrow("canvas");
  });
});
