import {afterEach, describe, expect, it, vi} from "vitest";

const {mainMock} = vi.hoisted(() => ({
  mainMock: vi.fn<(...args: readonly unknown[]) => number>(),
}));

vi.mock("../../../src/cli.js", () => ({main: mainMock}));

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

describe("plugin validate entry point", () => {
  it("runs the validator from the working directory", async () => {
    mainMock.mockReturnValue(0);

    await runEntryPoint();

    expect(mainMock.mock.calls[0]?.slice(0, 2)).toEqual([
      process.argv.slice(2),
      process.cwd(),
    ]);
    expect(process.exitCode).toBe(0);
  });

  it("reports a validation failure", async () => {
    mainMock.mockReturnValue(1);

    await runEntryPoint();

    expect(process.exitCode).toBe(1);
  });

  it("reports an input error", async () => {
    mainMock.mockReturnValue(2);

    await runEntryPoint();

    expect(process.exitCode).toBe(2);
  });
});
