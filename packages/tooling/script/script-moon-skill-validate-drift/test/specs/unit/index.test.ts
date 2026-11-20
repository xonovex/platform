import {afterEach, describe, expect, it, vi} from "vitest";

const {mainMock, stderrMock} = vi.hoisted(() => ({
  mainMock: vi.fn<(...args: readonly unknown[]) => number>(),
  stderrMock: vi.fn(),
}));

vi.mock("../../../src/validate-drift.js", () => ({main: mainMock}));
vi.mock("@xonovex/script-moon-common/adapters/output-node", () => ({
  processOutput: {stdout: vi.fn(), stderr: stderrMock},
}));

const originalExitCode = process.exitCode;
const originalLintMode = process.env.XONOVEX_LINT_MODE;

// The entry point runs on import, so each case resets the module registry and
// restores the exit code and the environment it reads.
const runEntryPoint = async (): Promise<void> => {
  await import("../../../src/index.js");
};

afterEach(() => {
  process.exitCode = originalExitCode;
  if (originalLintMode === undefined) delete process.env.XONOVEX_LINT_MODE;
  else process.env.XONOVEX_LINT_MODE = originalLintMode;
  vi.resetModules();
  vi.clearAllMocks();
});

describe("drift validator entry point", () => {
  it("runs from the working directory and reports the exit code", async () => {
    process.env.XONOVEX_LINT_MODE = "enforce";
    mainMock.mockReturnValue(1);

    await runEntryPoint();

    expect(mainMock.mock.calls[0]?.slice(0, 2)).toEqual([
      process.argv.slice(2),
      process.cwd(),
    ]);
    expect(mainMock.mock.calls[0]?.[2]).toMatchObject({
      driftMode: "enforce",
      scriptDirectory: expect.stringMatching(/src$/u) as unknown,
    });
    expect(process.exitCode).toBe(1);
  });

  it("reads an unset lint mode as warn", async () => {
    delete process.env.XONOVEX_LINT_MODE;
    mainMock.mockReturnValue(0);

    await runEntryPoint();

    expect(mainMock.mock.calls[0]?.[2]).toMatchObject({driftMode: "warn"});
  });

  it("reports a thrown error and exits non-zero", async () => {
    mainMock.mockImplementation(() => {
      throw new Error("broken");
    });

    await runEntryPoint();

    expect(stderrMock).toHaveBeenCalledWith("Error: broken");
    expect(process.exitCode).toBe(1);
  });

  it("stringifies a thrown value that is not an error", async () => {
    const thrown: unknown = "broken";
    mainMock.mockImplementation(() => {
      throw thrown;
    });

    await runEntryPoint();

    expect(stderrMock).toHaveBeenCalledWith("Error: broken");
  });
});
