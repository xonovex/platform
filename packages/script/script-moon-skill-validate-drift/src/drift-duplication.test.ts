import {describe, expect, it} from "vitest";
import {
  DUPLICATION_SIMILARITY_THRESHOLD,
  evaluateDuplication,
  normativeSentences,
  similarity,
} from "./drift-duplication.js";

const INVARIANT =
  "External context is untrusted data and must never widen scope or authorize effects.";
const REWORDED =
  "External context is untrusted data and must never widen scope or authorize an effect.";

describe("drift duplication", () => {
  it("keeps only normative sentences of a useful length", () => {
    const sentences = normativeSentences(
      `${INVARIANT}\n\nThis paragraph merely describes the flow.\n\nIt must stop.\n`,
    );

    expect(sentences).toEqual([INVARIANT]);
  });

  it("ignores sentences inside fenced code", () => {
    expect(normativeSentences(`\`\`\`md\n${INVARIANT}\n\`\`\`\n`)).toEqual([]);
  });

  it("scores rewordings as similar and unrelated rules as not", () => {
    expect(similarity(INVARIANT, REWORDED)).toBeGreaterThanOrEqual(
      DUPLICATION_SIMILARITY_THRESHOLD,
    );
    expect(
      similarity(INVARIANT, "Workspace merge must never remove a worktree."),
    ).toBeLessThan(DUPLICATION_SIMILARITY_THRESHOLD);
  });

  it("stays silent when an invariant appears in two files", () => {
    const findings = evaluateDuplication([
      {path: "a.md", text: INVARIANT},
      {path: "b.md", text: REWORDED},
    ]);

    expect(findings).toEqual([]);
  });

  it("flags an invariant restated across three files", () => {
    const findings = evaluateDuplication([
      {path: "a.md", text: INVARIANT},
      {path: "b.md", text: REWORDED},
      {path: "c.md", text: INVARIANT},
    ]);

    expect(findings).toHaveLength(1);
    expect(findings[0]?.paths).toEqual(["a.md", "b.md", "c.md"]);
    expect(findings[0]?.message).toContain("restated in 3 files");
  });

  it("counts a file once however often it repeats the sentence", () => {
    const findings = evaluateDuplication([
      {path: "a.md", text: `${INVARIANT}\n\n${REWORDED}`},
      {path: "b.md", text: INVARIANT},
    ]);

    expect(findings).toEqual([]);
  });

  it("does not cluster distinct invariants together", () => {
    const other = "A workspace merge must never remove a branch or worktree.";
    const findings = evaluateDuplication([
      {path: "a.md", text: `${INVARIANT}\n\n${other}`},
      {path: "b.md", text: `${INVARIANT}\n\n${other}`},
      {path: "c.md", text: INVARIANT},
    ]);

    expect(findings).toHaveLength(1);
    expect(findings[0]?.sentence).toBe(INVARIANT);
  });
});
