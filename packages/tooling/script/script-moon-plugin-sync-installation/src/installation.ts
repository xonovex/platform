import type {Harness, Plugin} from "./model.js";

const fence = "```";

const harnessSection = (harness: Harness, selector: string): string => {
  const command =
    harness === "Claude Code" ? "claude plugin install " : "codex plugin add ";
  const add =
    harness === "Claude Code"
      ? "claude plugin marketplace add "
      : "codex plugin marketplace add ";
  const contents =
    harness === "Claude Code"
      ? "Install this bundle in Claude Code. The harness discovers its bundled skills."
      : "Install this bundle in Codex. Codex loads the skills declared in its plugin manifest.";
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
};

export const installation = (plugin: Plugin, marketplace: string): string => {
  const selector = plugin.name + "@" + marketplace;
  const sections = plugin.harnesses.map((harness) =>
    harnessSection(harness, selector),
  );
  const dependencies =
    plugin.dependencies.length === 0
      ? ""
      : "\n\nPlugin dependencies: " +
        plugin.dependencies.map((name) => "`" + name + "`").join(", ") +
        ".";
  const companions =
    plugin.companions.length === 0
      ? ""
      : "\n\nOptional companions: " +
        plugin.companions.map((name) => "`" + name + "`").join(", ") +
        ".";
  return (
    "## Install\n\nAdd the Xonovex marketplace once, then install `" +
    plugin.name +
    "` in the selected harness. The bundle version is `" +
    plugin.version +
    "`.\n\n" +
    sections.join("\n\n") +
    dependencies +
    companions
  );
};
