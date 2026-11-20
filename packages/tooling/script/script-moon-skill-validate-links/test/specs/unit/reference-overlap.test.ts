import {describe, expect, it} from "vitest";
import {checkReferenceOverlap} from "../../../src/reference-overlap.js";

const words = (count: number, prefix: string): string =>
  Array.from({length: count}, (_, index) => `${prefix}${String(index)}`).join(
    " ",
  );

const references = (name: string, text: string) => ({
  name,
  references: [{path: `skills/using-${name}/references/r.md`, text}],
});

describe("checkReferenceOverlap", () => {
  it("fails a reference that copies most of a dependency reference", () => {
    const shared = words(40, "shared");

    expect(
      checkReferenceOverlap(
        references("xonovex-child", `${shared} ${words(10, "own")}`),
        references("xonovex-base", shared),
      ),
    ).toEqual([
      {
        level: "fail",
        message:
          "skill ownership: xonovex-child skills/using-xonovex-child/references/r.md duplicates xonovex-base skills/using-xonovex-base/references/r.md (100% five-word-shingle containment); keep only the dependent skill's specialization",
      },
    ]);
  });

  it("ignores code blocks, distinct prose, and files too short to compare", () => {
    const fenced = `\`\`\`\n${words(40, "shared")}\n\`\`\``;

    expect(
      checkReferenceOverlap(
        references("a", `${fenced} ${words(40, "a")}`),
        references("b", `${fenced} ${words(40, "b")}`),
      ),
    ).toEqual([]);
    expect(
      checkReferenceOverlap(
        references("a", words(10, "x")),
        references("b", words(10, "x")),
      ),
    ).toEqual([]);
  });
});
