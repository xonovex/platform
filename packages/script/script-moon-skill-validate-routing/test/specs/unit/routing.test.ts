import {join} from "node:path";
import {type FileSystem} from "@xonovex/script-moon-common/file-system";
import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {afterEach, describe, expect, it, vi} from "vitest";
import {main} from "../../../src/routing.js";

const ROOT = "/catalog";

afterEach(() => {
  vi.restoreAllMocks();
});

interface CatalogSkill {
  readonly name: string;
  readonly queries: readonly unknown[];
}

const writeCatalog = (skills: readonly CatalogSkill[]): FileSystem => {
  const files: Record<string, string> = {};
  for (const skill of skills) {
    const guide = join(ROOT, `skill-${skill.name}`, `${skill.name}-guide`);
    files[join(guide, "SKILL.md")] = `---\nname: ${skill.name}-guide\n---\n`;
    files[join(guide, "eval-queries.json")] = JSON.stringify(skill.queries);
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
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    expect(main(["--help"])).toBe(0);
    expect(log.mock.calls.at(0)?.at(0)).toContain(
      "moon-skill-validate-routing",
    );
  });

  it("rejects an unrecognized option", () => {
    const write = vi
      .spyOn(process.stderr, "write")
      .mockImplementation(() => true);

    expect(main(["--unknown"])).toBe(2);
    expect(write.mock.calls.at(0)?.at(0)).toContain("unrecognized argument");
  });

  it("passes when every skill owns a validation routing scenario", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => {});
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

    expect(main([ROOT], fs)).toBe(0);
    expect(log.mock.calls.at(-1)?.at(0)).toContain("every skill owns");
  });

  it("reports a skill whose only pairing was stripped", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => {});
    const fs = writeCatalog([
      {name: "owner", queries: [query("unpaired routing query", true)]},
      {name: "rival", queries: [query("a query nobody else carries", true)]},
    ]);

    expect(main([ROOT], fs)).toBe(1);
    expect(
      log.mock.calls.map((call) => String(call.at(0))).join("\n"),
    ).toContain("owns no validation-split routing scenario");
  });

  it("reports a query two skills both claim", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => {});
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

    expect(main([ROOT], fs)).toBe(1);
    expect(
      log.mock.calls.map((call) => String(call.at(0))).join("\n"),
    ).toContain(
      'owner-guide and rival-guide both mark "contested query" should_trigger',
    );
  });

  it("reports a rationale citing an operation no reference file backs", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => {});
    const fs = writeCatalog([
      {
        name: "owner",
        queries: [
          {
            query: "settle the option",
            rationale: "settling it is the owner-decide operation",
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

    expect(main([ROOT], fs)).toBe(1);
    expect(
      log.mock.calls.map((call) => String(call.at(0))).join("\n"),
    ).toContain("owner-guide cites the 'owner-decide' operation");
  });

  it("reports an unreadable catalog root", () => {
    const write = vi
      .spyOn(process.stderr, "write")
      .mockImplementation(() => true);

    expect(main(["/absent-catalog"], memoryFileSystem())).toBe(2);
    expect(write.mock.calls.at(0)?.at(0)).toContain("catalog root not found");
  });
});
