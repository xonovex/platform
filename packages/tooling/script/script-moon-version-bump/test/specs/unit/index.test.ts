import {errorLine} from "@xonovex/script-moon-common/logging";
import {afterEach, describe, expect, it, vi} from "vitest";

const {stderrMock, mainMock} = vi.hoisted(() => ({
  stderrMock: vi.fn<(line: string) => void>(),
  mainMock: vi.fn<(...args: readonly unknown[]) => number>(),
}));

vi.mock("../../../src/version-bump.js", () => ({main: mainMock}));
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

describe("version bump entry point", () => {
  it("reports the exit code the command returns", async () => {
    mainMock.mockReturnValue(0);

    await runEntryPoint();

    expect(mainMock).toHaveBeenCalledOnce();
    expect(mainMock.mock.calls[0]?.slice(0, 2)).toEqual([
      process.argv.slice(2),
      process.cwd(),
    ]);
    expect(process.exitCode).toBe(0);
    expect(stderrMock).not.toHaveBeenCalled();
  });

  it("logs a thrown error and exits non-zero", async () => {
    mainMock.mockImplementation(() => {
      throw new Error("no package.json found in current directory");
    });

    await runEntryPoint();

    expect(stderrMock).toHaveBeenCalledWith(
      errorLine("no package.json found in current directory"),
    );
    expect(process.exitCode).toBe(1);
  });

  it("stringifies a thrown value that is not an error", async () => {
    const thrown: unknown = "broken";
    mainMock.mockImplementation(() => {
      throw thrown;
    });

    await runEntryPoint();

    expect(stderrMock).toHaveBeenCalledWith(errorLine("broken"));
    expect(process.exitCode).toBe(1);
  });
});
