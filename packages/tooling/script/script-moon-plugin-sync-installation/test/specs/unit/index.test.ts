import {afterEach, describe, expect, it, vi} from "vitest";

const {mainMock, rootMock, stderrMock} = vi.hoisted(() => ({
  mainMock: vi.fn<(...args: readonly unknown[]) => number>(),
  rootMock: vi.fn<(...args: readonly unknown[]) => string>(),
  stderrMock: vi.fn(),
}));

vi.mock("../../../src/cli.js", () => ({main: mainMock}));
vi.mock("@xonovex/script-moon-common/workspace", () => ({
  findWorkspaceRoot: rootMock,
}));
vi.mock("@xonovex/script-moon-common/adapters/output-node", () => ({
  processOutput: {stdout: vi.fn(), stderr: stderrMock},
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

describe("installation sync entry point", () => {
  it("runs on the workspace above the script and reports the exit code", async () => {
    rootMock.mockReturnValue("/repo");
    mainMock.mockReturnValue(1);

    await runEntryPoint();

    expect(rootMock.mock.calls[0]?.[0]).toMatch(/src$/u);
    expect(mainMock.mock.calls[0]?.slice(0, 2)).toEqual([
      process.argv.slice(2),
      "/repo",
    ]);
    expect(process.exitCode).toBe(1);
  });

  it("reports a missing workspace and exits non-zero", async () => {
    rootMock.mockImplementation(() => {
      throw new Error("Could not find workspace root");
    });

    await runEntryPoint();

    expect(stderrMock).toHaveBeenCalledWith(
      expect.stringContaining("Could not find workspace root"),
    );
    expect(mainMock).not.toHaveBeenCalled();
    expect(process.exitCode).toBe(1);
  });

  it("stringifies a thrown value that is not an error", async () => {
    const thrown: unknown = "broken";
    rootMock.mockReturnValue("/repo");
    mainMock.mockImplementation(() => {
      throw thrown;
    });

    await runEntryPoint();

    expect(stderrMock).toHaveBeenCalledWith(expect.stringContaining("broken"));
  });
});
