import {afterEach, describe, expect, it, vi} from "vitest";

const {mainMock, stderrMock} = vi.hoisted(() => ({
  mainMock: vi.fn<(...args: readonly unknown[]) => Promise<number>>(),
  stderrMock: vi.fn(),
}));

vi.mock("../../../src/audit.js", () => ({main: mainMock}));
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

describe("source audit entry point", () => {
  it("audits from the working directory and reports the exit code", async () => {
    mainMock.mockResolvedValue(1);

    await runEntryPoint();

    expect(mainMock.mock.calls[0]?.slice(0, 2)).toEqual([
      process.argv.slice(2),
      process.cwd(),
    ]);
    expect(mainMock.mock.calls[0]?.[2]).toMatchObject({
      today: expect.any(Date) as unknown,
    });
    expect(process.exitCode).toBe(1);
  });

  it.each([
    [new Error("broken"), "error: broken"],
    ["refused", "error: refused"],
  ])("reports a thrown %s and exits 2", async (thrown, message) => {
    mainMock.mockRejectedValue(thrown);

    await runEntryPoint();

    expect(stderrMock).toHaveBeenCalledWith(message);
    expect(process.exitCode).toBe(2);
  });
});
