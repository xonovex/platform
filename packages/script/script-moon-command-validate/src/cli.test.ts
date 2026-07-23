import {describe, expect, it, vi} from "vitest";
import {main} from "./cli.js";

describe("command validator CLI", () => {
  it("renders help", () => {
    const log = vi.spyOn(console, "log").mockImplementation(() => {});

    expect(main(["--help"])).toBe(0);
    expect(log).toHaveBeenCalledWith(expect.stringContaining("usage:"));
  });

  it("rejects invalid input", () => {
    const stderr = vi
      .spyOn(process.stderr, "write")
      .mockImplementation(() => true);

    expect(main(["--unknown"])).toBe(2);
    expect(main(["one", "two"])).toBe(2);
    expect(stderr).toHaveBeenCalled();
  });
});
