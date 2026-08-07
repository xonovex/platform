import {afterEach, describe, expect, it, vi} from "vitest";

const {mainMock} = vi.hoisted(() => ({mainMock: vi.fn()}));

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

describe("command validate entry point", () => {
  it("passes the arguments through to the command", async () => {
    mainMock.mockReturnValue(0);

    await runEntryPoint();

    expect(mainMock).toHaveBeenCalledWith(process.argv.slice(2));
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
