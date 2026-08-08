import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  checkCommandBudgets,
  commandFiles,
} from "../../../src/command-budgets.js";

const words = (count: number) =>
  Array.from({length: count}, () => "word").join(" ");

const REPOSITORY_ROOT = "/repo";
const PACKAGE_DIR = join(REPOSITORY_ROOT, "packages", "command", "pkg");

const packageWith = (
  commands: Readonly<Record<string, string>>,
  budgets?: Readonly<Record<string, number>>,
): FileSystem =>
  memoryFileSystem({
    directories: [join(PACKAGE_DIR, "commands")],
    files: {
      ...Object.fromEntries(
        Object.entries(commands).map(([name, text]) => [
          join(PACKAGE_DIR, "commands", name),
          text,
        ]),
      ),
      ...(budgets === undefined
        ? {}
        : {[join(REPOSITORY_ROOT, "budgets.json")]: JSON.stringify(budgets)}),
    },
  });

describe("command budgets", () => {
  it("keys command files by repository-relative path", () => {
    const fs = packageWith({"a.md": "one two"});

    expect(commandFiles(PACKAGE_DIR, REPOSITORY_ROOT, fs)).toEqual([
      {
        kind: "command",
        path: "packages/command/pkg/commands/a.md",
        text: "one two",
      },
    ]);
  });

  it("returns nothing for a package without a commands directory", () => {
    const fs = memoryFileSystem({directories: [REPOSITORY_ROOT]});

    expect(commandFiles(REPOSITORY_ROOT, REPOSITORY_ROOT, fs)).toEqual([]);
  });

  it("stays silent when every command is within budget", () => {
    const fs = packageWith(
      {"a.md": "one two three"},
      {"packages/command/pkg/commands/a.md": 3},
    );

    expect(
      checkCommandBudgets(PACKAGE_DIR, REPOSITORY_ROOT, "warn", fs),
    ).toEqual([]);
  });

  it("warns in warn mode and fails in enforce mode", () => {
    const fs = packageWith(
      {"a.md": words(300)},
      {"packages/command/pkg/commands/a.md": 10},
    );

    const warned = checkCommandBudgets(
      PACKAGE_DIR,
      REPOSITORY_ROOT,
      "warn",
      fs,
    );
    const enforced = checkCommandBudgets(
      PACKAGE_DIR,
      REPOSITORY_ROOT,
      "enforce",
      fs,
    );

    expect(warned).toHaveLength(1);
    expect(warned[0]?.severity).toBe("warning");
    expect(warned[0]?.code).toBe("budget");
    expect(enforced[0]?.severity).toBe("error");
  });

  it("applies the command cap when the manifest omits the file", () => {
    const fs = packageWith({"a.md": words(300)});

    const issues = checkCommandBudgets(
      PACKAGE_DIR,
      REPOSITORY_ROOT,
      "warn",
      fs,
    );

    expect(issues).toHaveLength(1);
    expect(issues[0]?.message).toContain("250-word cap");
  });

  it("reports a malformed manifest as an error regardless of mode", () => {
    const fs = packageWith({"a.md": "one"});
    fs.writeFile(join(REPOSITORY_ROOT, "budgets.json"), "{");

    const issues = checkCommandBudgets(
      PACKAGE_DIR,
      REPOSITORY_ROOT,
      "warn",
      fs,
    );

    expect(issues).toHaveLength(1);
    expect(issues[0]?.severity).toBe("error");
    expect(issues[0]?.code).toBe("budget-manifest");
  });
});
