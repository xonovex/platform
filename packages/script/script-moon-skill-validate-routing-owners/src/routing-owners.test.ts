import {mkdirSync, mkdtempSync, rmSync, writeFileSync} from "node:fs";
import {tmpdir} from "node:os";
import {join} from "node:path";
import {afterEach, describe, expect, it, vi} from "vitest";
import {main} from "./routing-owners.js";

const temporaryDirectories: string[] = [];

afterEach(() => {
  vi.restoreAllMocks();
  for (const directory of temporaryDirectories) {
    rmSync(directory, {recursive: true, force: true});
  }
  temporaryDirectories.length = 0;
});

interface CatalogSkill {
  readonly name: string;
  readonly queries: readonly unknown[];
}

const writeCatalog = (skills: readonly CatalogSkill[]): string => {
  const root = mkdtempSync(join(tmpdir(), "routing-owners-"));
  temporaryDirectories.push(root);
  for (const skill of skills) {
    const guide = join(root, `skill-${skill.name}`, `${skill.name}-guide`);
    mkdirSync(guide, {recursive: true});
    writeFileSync(
      join(guide, "SKILL.md"),
      `---\nname: ${skill.name}-guide\n---\n`,
    );
    writeFileSync(
      join(guide, "eval-queries.json"),
      JSON.stringify(skill.queries),
    );
  }
  return root;
};

const query = (text: string, shouldTrigger: boolean) => ({
  query: text,
  rationale: "pairing fixture",
  should_trigger: shouldTrigger,
  split: "validation" as const,
});

describe("routing owners check", () => {
  it("renders help without reading a catalog", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => undefined);

    expect(main(["--help"])).toBe(0);
    expect(log.mock.calls.at(0)?.at(0)).toContain(
      "moon-skill-validate-routing-owners",
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
    const log = vi.spyOn(console, "log").mockImplementation(() => undefined);
    const root = writeCatalog([
      {
        name: "owner",
        queries: [query("owner's query", true), query("rival's query", false)],
      },
      {
        name: "rival",
        queries: [query("rival's query", true), query("owner's query", false)],
      },
    ]);

    expect(main([root])).toBe(0);
    expect(log.mock.calls.at(-1)?.at(0)).toContain("every skill owns");
  });

  it("reports a skill whose only pairing was stripped", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => undefined);
    const root = writeCatalog([
      {name: "owner", queries: [query("unpaired routing query", true)]},
      {name: "rival", queries: [query("a query nobody else carries", true)]},
    ]);

    expect(main([root])).toBe(1);
    expect(
      log.mock.calls.map((call) => String(call.at(0))).join("\n"),
    ).toContain("owns no validation-split routing scenario");
  });

  it("reports an unreadable catalog root", () => {
    const write = vi
      .spyOn(process.stderr, "write")
      .mockImplementation(() => true);

    expect(main([join(tmpdir(), "routing-owners-absent")])).toBe(2);
    expect(write.mock.calls.at(0)?.at(0)).toContain("catalog root not found");
  });
});
