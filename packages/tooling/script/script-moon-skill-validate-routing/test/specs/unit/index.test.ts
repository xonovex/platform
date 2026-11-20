import {afterEach, describe, expect, it, vi} from "vitest";

const {mainMock} = vi.hoisted(() => ({
  mainMock: vi.fn<(...args: readonly unknown[]) => number>(),
}));

vi.mock("../../../src/routing.js", () => ({main: mainMock}));

const originalExitCode = process.exitCode;

afterEach(() => {
  process.exitCode = originalExitCode;
  vi.resetModules();
  vi.clearAllMocks();
});

describe("routing validator entry point", () => {
  it("runs from the working directory and reports the exit code", async () => {
    mainMock.mockReturnValue(1);

    await import("../../../src/index.js");

    expect(mainMock.mock.calls[0]?.slice(0, 2)).toEqual([
      process.argv.slice(2),
      process.cwd(),
    ]);
    expect(process.exitCode).toBe(1);
  });
});
