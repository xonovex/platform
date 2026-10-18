import {describe, expect, it} from "vitest";
import {
  installationEnd,
  installationStart,
  replaceInstallation,
} from "../../../src/installation-section.js";

const block = "## Install\n\nNew instructions.";

describe("replaceInstallation", () => {
  it("replaces the text between the markers and keeps the prose around them", () => {
    const before = [
      "# Plugin",
      "",
      "Intro.",
      "",
      installationStart,
      "",
      "Old instructions.",
      "",
      installationEnd,
      "",
      "## Usage",
      "",
      "Keep this.",
      "",
    ].join("\n");
    expect(replaceInstallation(before, block)).toBe(
      [
        "# Plugin",
        "",
        "Intro.",
        "",
        installationStart,
        "",
        block,
        "",
        installationEnd,
        "",
        "## Usage",
        "",
        "Keep this.",
        "",
      ].join("\n"),
    );
  });

  it("replaces a legacy install heading and keeps the following section", () => {
    const before =
      "# Plugin\n\nIntro.\n\n## Installation\n\nOld instructions.\n\n## Usage\n\nKeep this.\n";
    const updated = replaceInstallation(before, block);
    expect(updated).not.toContain("Old instructions.");
    expect(updated).toContain("Intro.");
    expect(updated).toContain("## Usage\n\nKeep this.\n");
    expect(updated).toContain(installationStart + "\n\n" + block);
  });

  it("appends the block when the text has no marker and no install heading", () => {
    const before = "# Plugin\n\nIntro.\n\n\n";
    expect(replaceInstallation(before, block)).toBe(
      "# Plugin\n\nIntro.\n\n" +
        installationStart +
        "\n\n" +
        block +
        "\n\n" +
        installationEnd +
        "\n",
    );
  });

  it("is idempotent", () => {
    const once = replaceInstallation("# Plugin\n\n## Install\n\nOld.\n", block);
    expect(replaceInstallation(once, block)).toBe(once);
  });

  it.each([
    installationStart,
    installationEnd + "\n" + installationStart,
    installationStart + "\n" + installationStart + "\n" + installationEnd,
    installationStart + "\n" + installationEnd + "\n" + installationEnd,
  ])("rejects an incomplete or duplicated marker block: %s", (text) => {
    expect(() => replaceInstallation(text, block)).toThrow(
      "one complete block",
    );
  });
});
