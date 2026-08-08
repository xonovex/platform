import {memoryFileSystem} from "@xonovex/script-moon-common/file-system-memory";
import {describe, expect, it} from "vitest";
import {
  coinedTerms,
  definedTerms,
  evaluateVocabulary,
  readVocabularyManifest,
} from "../../../src/drift-vocabulary.js";

describe("drift vocabulary", () => {
  it("treats headings and bold lead-ins as definitions", () => {
    const terms = definedTerms(
      "## Cold boundary\n\n**Handoff** — placement into a native system.\n",
    );

    expect([...terms].toSorted()).toEqual(["cold boundary", "handoff"]);
  });

  it("ignores definitions inside fenced code", () => {
    const terms = definedTerms("```md\n## Handoff\n```\n\nplain prose\n");

    expect([...terms]).toEqual([]);
  });

  it("lets the owner define its own term", () => {
    const findings = evaluateVocabulary(
      [{path: "contract.md", text: "## Handoff\n"}],
      {Handoff: "contract.md"},
    );

    expect(findings).toEqual([]);
  });

  it("flags another file redefining an owned term", () => {
    const findings = evaluateVocabulary(
      [{path: "handoffs.md", text: "## Handoff\n"}],
      {Handoff: "contract.md"},
    );

    expect(findings).toHaveLength(1);
    expect(findings[0]?.message).toContain("contract.md owns");
  });

  it("allows using an owned term without defining it", () => {
    const findings = evaluateVocabulary(
      [{path: "execute.md", text: "Return a handoff to the next role.\n"}],
      {Handoff: "contract.md"},
    );

    expect(findings).toEqual([]);
  });

  it("warns about a coined term that the manifest does not declare", () => {
    const text =
      "`warm path` is a sequence of operations.\n\nStay on the warm path.\n";

    expect([...coinedTerms(text)]).toEqual(["warm path"]);

    const findings = evaluateVocabulary([{path: "execute.md", text}], {});

    expect(findings).toHaveLength(1);
    expect(findings[0]?.message).toContain("coined term 'warm path'");
  });

  it("ignores a phrase mentioned once, which is an example not vocabulary", () => {
    // Real cases: a shell command and a code snippet, each shown a single time.
    expect([
      ...coinedTerms("`git remote get-url origin` is the way to read it.\n"),
    ]).toEqual([]);
    expect([
      ...coinedTerms("`square extends rectangle` is a classic violation.\n"),
    ]).toEqual([]);
  });

  it("accepts a coined term once it is declared", () => {
    const findings = evaluateVocabulary(
      [
        {
          path: "contract.md",
          text: "`warm path` is a sequence of operations.\n\nUse the warm path.\n",
        },
      ],
      {"warm path": "contract.md"},
    );

    expect(findings).toEqual([]);
  });

  it("reads a manifest and reports malformed content", () => {
    const fs = memoryFileSystem({
      files: {
        "/repo/valid.json": JSON.stringify({Handoff: "contract.md"}),
        "/repo/malformed.json": "{",
        "/repo/shape.json": JSON.stringify({Handoff: 3}),
      },
    });

    expect(readVocabularyManifest("/repo/valid.json", fs).manifest).toEqual({
      Handoff: "contract.md",
    });
    expect(readVocabularyManifest("/repo/missing.json", fs).manifest).toEqual(
      {},
    );
    expect(readVocabularyManifest("/repo/malformed.json", fs).error).toContain(
      "not valid JSON",
    );
    expect(readVocabularyManifest("/repo/shape.json", fs).error).toContain(
      "owning file",
    );
  });
});
