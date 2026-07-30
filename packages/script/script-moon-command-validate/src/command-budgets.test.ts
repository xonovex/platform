import {mkdirSync, mkdtempSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {describe, expect, it} from "vitest";
import {checkCommandBudgets, commandFiles} from "./command-budgets.js";

const words = (count: number) =>
  Array.from({length: count}, () => "word").join(" ");

const packageWith = (
  commands: Readonly<Record<string, string>>,
  budgets?: Readonly<Record<string, number>>,
): {readonly packageDir: string; readonly repositoryRoot: string} => {
  const repositoryRoot = mkdtempSync(join(tmpdir(), "command-budgets-"));
  const packageDir = join(repositoryRoot, "packages", "command", "pkg");
  mkdirSync(join(packageDir, "commands"), {recursive: true});
  for (const [name, text] of Object.entries(commands)) {
    writeFileSync(join(packageDir, "commands", name), text, "utf8");
  }
  if (budgets !== undefined) {
    writeFileSync(
      join(repositoryRoot, "budgets.json"),
      JSON.stringify(budgets),
      "utf8",
    );
  }
  return {packageDir, repositoryRoot};
};

describe("command budgets", () => {
  it("keys command files by repository-relative path", () => {
    const {packageDir, repositoryRoot} = packageWith({"a.md": "one two"});

    expect(commandFiles(packageDir, repositoryRoot)).toEqual([
      {
        kind: "command",
        path: "packages/command/pkg/commands/a.md",
        text: "one two",
      },
    ]);
  });

  it("returns nothing for a package without a commands directory", () => {
    const repositoryRoot = mkdtempSync(join(tmpdir(), "command-budgets-"));

    expect(commandFiles(repositoryRoot, repositoryRoot)).toEqual([]);
  });

  it("stays silent when every command is within budget", () => {
    const {packageDir, repositoryRoot} = packageWith(
      {"a.md": "one two three"},
      {"packages/command/pkg/commands/a.md": 3},
    );

    expect(checkCommandBudgets(packageDir, repositoryRoot)).toEqual([]);
  });

  it("warns in warn mode and fails in enforce mode", () => {
    const {packageDir, repositoryRoot} = packageWith(
      {"a.md": words(300)},
      {"packages/command/pkg/commands/a.md": 10},
    );

    const warned = checkCommandBudgets(packageDir, repositoryRoot, "warn");
    const enforced = checkCommandBudgets(packageDir, repositoryRoot, "enforce");

    expect(warned).toHaveLength(1);
    expect(warned[0]?.severity).toBe("warning");
    expect(warned[0]?.code).toBe("budget");
    expect(enforced[0]?.severity).toBe("error");
  });

  it("applies the command cap when the manifest omits the file", () => {
    const {packageDir, repositoryRoot} = packageWith({"a.md": words(300)});

    const issues = checkCommandBudgets(packageDir, repositoryRoot, "warn");

    expect(issues).toHaveLength(1);
    expect(issues[0]?.message).toContain("250-word cap");
  });

  it("reports a malformed manifest as an error regardless of mode", () => {
    const {packageDir, repositoryRoot} = packageWith({"a.md": "one"});
    writeFileSync(join(repositoryRoot, "budgets.json"), "{", "utf8");

    const issues = checkCommandBudgets(packageDir, repositoryRoot, "warn");

    expect(issues).toHaveLength(1);
    expect(issues[0]?.severity).toBe("error");
    expect(issues[0]?.code).toBe("budget-manifest");
  });
});
