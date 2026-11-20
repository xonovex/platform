import {afterEach, describe, expect, it, vi} from "vitest";
import {processOutput} from "../../../../src/adapters/output-node.js";

describe("processOutput", () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("writes a stdout line to the process stdout with a newline", () => {
    const write = vi.spyOn(process.stdout, "write").mockReturnValue(true);

    processOutput.stdout("report");

    expect(write).toHaveBeenCalledWith("report\n");
  });

  it("writes a stderr line to the process stderr with a newline", () => {
    const write = vi.spyOn(process.stderr, "write").mockReturnValue(true);

    processOutput.stderr("problem");

    expect(write).toHaveBeenCalledWith("problem\n");
  });
});
