import {describe, expect, it} from "vitest";
import {memoryOutput} from "../../../../src/adapters/output-memory.js";

describe("memoryOutput", () => {
  it("keeps the lines of each stream apart and in order", () => {
    const output = memoryOutput();

    output.stdout("one");
    output.stderr("problem");
    output.stdout("two");

    expect(output.stdoutLines()).toEqual(["one", "two"]);
    expect(output.stderrLines()).toEqual(["problem"]);
  });

  it("returns a copy that a later write does not change", () => {
    const output = memoryOutput();
    output.stdout("one");
    const before = output.stdoutLines();

    output.stdout("two");

    expect(before).toEqual(["one"]);
  });
});
