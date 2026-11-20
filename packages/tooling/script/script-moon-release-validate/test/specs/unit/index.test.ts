import {afterEach, describe, expect, it, vi} from "vitest";

const {mainMock} = vi.hoisted(() => ({
  mainMock: vi.fn<(...args: readonly unknown[]) => number>(),
}));

vi.mock("../../../src/release-validate.js", () => ({main: mainMock}));

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
  vi.restoreAllMocks();
});

describe("release validate entry point", () => {
  it("reports the exit code main returns", async () => {
    mainMock.mockReturnValue(1);

    await runEntryPoint();

    expect(mainMock).toHaveBeenCalledOnce();
    expect(process.exitCode).toBe(1);
  });

  it("reports an unexpected error and exits non-zero", async () => {
    const write = vi
      .spyOn(process.stderr, "write")
      .mockImplementation(() => true);
    mainMock.mockImplementation(() => {
      throw new Error("broken tree");
    });

    await runEntryPoint();

    expect(write).toHaveBeenCalledWith(
      "Release validation failed unexpectedly: broken tree\n",
    );
    expect(process.exitCode).toBe(1);
  });

  it("stringifies a thrown value that is not an error", async () => {
    const write = vi
      .spyOn(process.stderr, "write")
      .mockImplementation(() => true);
    const thrown: unknown = "broken";
    mainMock.mockImplementation(() => {
      throw thrown;
    });

    await runEntryPoint();

    expect(write).toHaveBeenCalledWith(
      "Release validation failed unexpectedly: broken\n",
    );
    expect(process.exitCode).toBe(1);
  });
});
