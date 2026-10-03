import {readText} from "./files.js";
import type {Catalog, Plugin, ReadmeUpdate} from "./model.js";

const start = "<!-- xonovex:installation:start -->";
const end = "<!-- xonovex:installation:end -->";
const fence = "```";

export const installation = (plugin: Plugin, marketplace: string): string => {
  const selector = plugin.name + "@" + marketplace;
  const sections = plugin.harnesses.map((harness) => {
    const command =
      harness === "Claude Code"
        ? "claude plugin install "
        : "codex plugin add ";
    const add =
      harness === "Claude Code"
        ? "claude plugin marketplace add "
        : "codex plugin marketplace add ";
    const contents =
      harness === "Claude Code"
        ? "Install this bundle in Claude Code. The harness discovers its bundled skills and commands."
        : "Install this bundle in Codex. Codex loads the skills declared in its plugin manifest; Claude Code slash commands are separate components.";
    return (
      "### " +
      harness +
      "\n\n" +
      contents +
      "\n\n" +
      fence +
      "bash\n" +
      add +
      "xonovex/platform\n" +
      command +
      selector +
      "\n" +
      fence
    );
  });
  const dependencies =
    plugin.dependencies.length === 0
      ? ""
      : "\n\nPlugin dependencies: " +
        plugin.dependencies.map((name) => "`" + name + "`").join(", ") +
        ".";
  return (
    "## Install\n\nAdd the Xonovex marketplace once, then install `" +
    plugin.name +
    "` in the selected harness. The bundle version is `" +
    plugin.version +
    "`.\n\n" +
    sections.join("\n\n") +
    dependencies
  );
};

export const replaceInstallation = (text: string, block: string): string => {
  const marked = start + "\n\n" + block + "\n\n" + end;
  const begins = text.indexOf(start);
  const ends = text.indexOf(end);
  if (
    (begins === -1) !== (ends === -1) ||
    (begins !== -1 &&
      (ends < begins ||
        text.includes(start, begins + start.length) ||
        text.includes(end, ends + end.length)))
  ) {
    throw new Error("Installation markers must form one complete block.");
  }
  if (begins !== -1)
    return text.slice(0, begins) + marked + text.slice(ends + end.length);
  const heading = /^## (?:Install|Installation)\s*$/mu.exec(text);
  if (heading?.index !== undefined) {
    const remainder = text.slice(heading.index + heading[0].length);
    const nextHeading = /^## /mu.exec(remainder);
    const next =
      nextHeading?.index === undefined
        ? text.length
        : heading.index + heading[0].length + nextHeading.index;
    return text.slice(0, heading.index) + marked + "\n\n" + text.slice(next);
  }
  return text.trimEnd() + "\n\n" + marked + "\n";
};

export const readmeUpdates = (
  root: string,
  catalog: Catalog,
): readonly ReadmeUpdate[] =>
  catalog.plugins.map((plugin) => {
    const path = plugin.directory + "/README.md";
    const previous = readText(root, path);
    return {
      path,
      previous,
      content: replaceInstallation(
        previous,
        installation(plugin, catalog.marketplace),
      ),
    };
  });
