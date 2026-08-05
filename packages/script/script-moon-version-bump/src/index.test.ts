import {afterEach, describe, expect, it, vi} from "vitest";

const {logErrorMock, mainMock} = vi.hoisted(() => ({
  logErrorMock: vi.fn(),
  mainMock: vi.fn(),
}));

vi.mock("./version-bump.js", () => ({main: mainMock}));
vi.mock("@xonovex/script-moon-common/logging", () => ({
  logError: logErrorMock,
}));

const originalExitCode = process.exitCode;

// The entry point runs on import, so each case resets the module registry and
// restores the exit code it sets.
const runEntryPoint = async (): Promise<void> => {
  await import("./index.js");
};

afterEach(() => {
  process.exitCode = originalExitCode;
  vi.resetModules();
  vi.clearAllMocks();
});

describe("version bump entry point", () => {
  it("reports the exit code the command returns", async () => {
    mainMock.mockReturnValue(0);

    await runEntryPoint();

    expect(mainMock).toHaveBeenCalledWith(process.argv.slice(2));
    expect(process.exitCode).toBe(0);
    expect(logErrorMock).not.toHaveBeenCalled();
  });

  it("logs a thrown error and exits non-zero", async () => {
    mainMock.mockImplementation(() => {
      throw new Error("no package.json found in current directory");
    });

    await runEntryPoint();

    expect(logErrorMock).toHaveBeenCalledWith(
      "no package.json found in current directory",
    );
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
