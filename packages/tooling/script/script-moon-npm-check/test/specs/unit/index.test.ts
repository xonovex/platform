import {afterEach, describe, expect, it, vi} from "vitest";

const {logErrorMock, mainMock} = vi.hoisted(() => ({
  logErrorMock: vi.fn(),
  mainMock: vi.fn<(...args: readonly unknown[]) => number>(),
}));

vi.mock("../../../src/npm-check.js", () => ({main: mainMock}));
vi.mock("@xonovex/script-moon-common/logging", () => ({
  logError: logErrorMock,
}));

const originalExitCode = process.exitCode;

// The entry point runs on import, so each case resets the module registry and
// restores the exit code it sets.
const runEntryPoint = async (): Promise<void> => {
  await import("../../../src/index.js");
};

afterEach(() => {
  process.exitCode = originalExitCode;
  vi.resetModules();
  vi.clearAllMocks();
});

describe("npm check entry point", () => {
  it("checks the working directory and reports the exit code", async () => {
    mainMock.mockReturnValue(1);

    await runEntryPoint();

    expect(mainMock.mock.calls[0]?.slice(0, 2)).toEqual([
      process.argv.slice(2),
      process.cwd(),
    ]);
    expect(process.exitCode).toBe(1);
  });

  it("logs a thrown error and exits non-zero", async () => {
    mainMock.mockImplementation(() => {
      throw new Error("broken");
    });

    await runEntryPoint();

    expect(logErrorMock).toHaveBeenCalledWith("broken");
    expect(process.exitCode).toBe(1);
  });

  it("stringifies a thrown value that is not an error", async () => {
    const thrown: unknown = "broken";
    mainMock.mockImplementation(() => {
      throw thrown;
    });

    await runEntryPoint();

    expect(logErrorMock).toHaveBeenCalledWith("broken");
    expect(process.exitCode).toBe(1);
  });
});
