import {join} from "node:path";
import {memoryFileSystem} from "@xonovex/script-moon-common/adapters/file-system-memory";
import {memoryOutput} from "@xonovex/script-moon-common/adapters/output-memory";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {describe, expect, it} from "vitest";
import {main} from "../../../src/routing.js";

const ROOT = "/catalog";

// Runs the check from the catalog root and returns the exit code with its output.
const run = (argv: readonly string[], fs: FileSystem = memoryFileSystem()) => {
  const output = memoryOutput();
  const code = main(argv, ROOT, {fs, output});
  return {
    code,
    stdout: output.stdoutLines().join("\n"),
    stderr: output.stderrLines().join("\n"),
  };
};

interface CatalogSkill {
  readonly name: string;
  readonly queries: readonly unknown[];
}

const writeCatalog = (skills: readonly CatalogSkill[]): FileSystem => {
  const files: Record<string, string> = {};
  for (const skill of skills) {
    const directory = join(
      ROOT,
      `plugin-${skill.name}`,
      "skills",
      `using-${skill.name}`,
    );
    files[join(directory, "SKILL.md")] =
      `---\nname: using-${skill.name}\n---\n`;
    files[join(directory, "eval-queries.json")] = JSON.stringify(skill.queries);
  }
  return memoryFileSystem({files});
};

const query = (text: string, shouldTrigger: boolean) => ({
  query: text,
  rationale: "pairing fixture",
  should_trigger: shouldTrigger,
  split: "validation" as const,
});

describe("routing owners check", () => {
  it("renders help without reading a catalog", () => {
    const result = run(["--help"]);

    expect(result.code).toBe(0);
    expect(result.stdout).toContain("moon-skill-validate-routing");
  });

  it("rejects an unrecognized option", () => {
    expect(run(["--unknown"])).toMatchObject({
      code: 2,
      stderr: "Error: unrecognized argument: --unknown",
    });
  });

  it("passes when every skill owns a validation routing scenario", () => {
    const fs = writeCatalog([
      {
        name: "owner",
        queries: [query("owner's query", true), query("rival's query", false)],
      },
      {
        name: "rival",
        queries: [query("rival's query", true), query("owner's query", false)],
      },
    ]);

    const result = run(["."], fs);

    expect(result.code).toBe(0);
    expect(result.stdout).toContain("every skill owns");
  });

  it("reports a skill whose only pairing was stripped", () => {
    const fs = writeCatalog([
      {name: "owner", queries: [query("unpaired routing query", true)]},
      {name: "rival", queries: [query("a query nobody else carries", true)]},
    ]);

    const result = run(["."], fs);

    expect(result.code).toBe(1);
    expect(result.stdout).toContain(
      "owns no validation-split routing scenario",
    );
  });

  it("reports a query two skills both claim", () => {
    const fs = writeCatalog([
      {
        name: "owner",
        queries: [
          query("contested query", true),
          query("rival's query", false),
        ],
      },
      {
        name: "rival",
        queries: [
          query("contested query", true),
          query("owner's query", false),
        ],
      },
    ]);

    const result = run(["."], fs);

    expect(result.code).toBe(1);
    expect(result.stdout).toContain(
      'using-owner and using-rival both mark "contested query" should_trigger',
    );
    expect(result.stdout.split("\n").slice(-2)).toEqual([
      "",
      "Routing: 3 finding(s)",
    ]);
  });

  it("reports a rationale citing an operation no reference file backs", () => {
    const fs = writeCatalog([
      {
        name: "owner",
        queries: [
          {
            query: "settle the option",
            rationale: "settling it is the `owner-decide` operation",
            should_trigger: true,
            split: "validation" as const,
          },
          query("rival's query", false),
        ],
      },
      {
        name: "rival",
        queries: [
          query("rival's query", true),
          query("settle the option", false),
        ],
      },
    ]);

    const result = run(["."], fs);

    expect(result.code).toBe(1);
    expect(result.stdout).toContain(
      "using-owner cites the 'owner-decide' operation",
    );
  });

  it("reports an unreadable catalog root", () => {
    const result = run(["/absent-catalog"]);

    expect(result.code).toBe(2);
    expect(result.stderr).toContain("catalog root not found");
  });
});
