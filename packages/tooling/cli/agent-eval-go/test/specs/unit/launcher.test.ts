import {describe, expect, it, vi} from "vitest";
import {
  childExitCode,
  findBinary,
  getBinaryName,
  getPlatformPackage,
  launchAgentEval,
} from "../../../src/launcher.js";

const signalNumber = (signal: NodeJS.Signals): number =>
  signal === "SIGTERM" ? 15 : 2;

describe("getPlatformPackage", () => {
  it("selects the package for a supported platform", () => {
    expect(getPlatformPackage("linux", "x64")).toBe(
      "@xonovex/agent-eval-go-linux-x64",
    );
  });

  it("returns undefined for an unsupported architecture", () => {
    expect(getPlatformPackage("linux", "riscv64")).toBeUndefined();
  });
});

describe("getBinaryName", () => {
  it("selects the Windows executable suffix only for Windows", () => {
    expect(getBinaryName("win32")).toBe("agent-eval-go.exe");
    expect(getBinaryName("linux")).toBe("agent-eval-go");
  });
});

describe("findBinary", () => {
  it("asks the lookup for the platform package and binary name", () => {
    const locateBinary = vi.fn(
      () => "/packages/win32-x64/bin/agent-eval-go.exe",
    );

    const result = findBinary("win32", "x64", {locateBinary});

    expect(locateBinary).toHaveBeenCalledWith(
      "@xonovex/agent-eval-go-win32-x64",
      "agent-eval-go.exe",
    );
    expect(result).toEqual({
      ok: true,
      binaryPath: "/packages/win32-x64/bin/agent-eval-go.exe",
    });
  });

  it("reports unsupported platforms without consulting the lookup", () => {
    const locateBinary = vi.fn();

    const result = findBinary("aix", "ppc64", {locateBinary});

    expect(locateBinary).not.toHaveBeenCalled();
    expect(result).toEqual({
      ok: false,
      error: "No binary available for aix-ppc64",
    });
  });

  it("reports a binary that is not installed", () => {
    const result = findBinary("linux", "x64", {
      locateBinary: () => {},
    });

    expect(result).toEqual({
      ok: false,
      error:
        "Binary not found. Install with: npm install -g @xonovex/agent-eval-go",
    });
  });
});

describe("childExitCode", () => {
  it("preserves a normal child exit code", () => {
    expect(childExitCode(7, null, signalNumber)).toBe(7);
  });

  it("maps signal termination to a non-zero shell exit code", () => {
    expect(childExitCode(null, "SIGTERM", signalNumber)).toBe(143);
  });

  it("fails when the child closes without a code or signal", () => {
    expect(childExitCode(null, null, signalNumber)).toBe(1);
  });
});

describe("launchAgentEval", () => {
  const baseRuntime = {
    currentPlatform: "linux",
    architecture: "x64",
    arguments: [] as readonly string[],
    environment: {},
    signalNumber,
    locateBinary: () => "/packages/linux-x64/bin/agent-eval-go",
    reportError: vi.fn(),
  } as const;

  it("starts the resolved binary and returns its exit code", async () => {
    const runBinary = vi.fn(() =>
      Promise.resolve({kind: "closed", code: 7, signal: null} as const),
    );

    const code = await launchAgentEval({
      ...baseRuntime,
      arguments: ["run", "task"],
      environment: {TEST_ENV: "enabled"},
      runBinary,
    });

    expect(runBinary).toHaveBeenCalledWith(
      "/packages/linux-x64/bin/agent-eval-go",
      ["run", "task"],
      {TEST_ENV: "enabled"},
    );
    expect(code).toBe(7);
  });

  it("maps a signal that ends the child to a shell exit code", async () => {
    const code = await launchAgentEval({
      ...baseRuntime,
      runBinary: () =>
        Promise.resolve({kind: "closed", code: null, signal: "SIGTERM"}),
    });

    expect(code).toBe(143);
  });

  it("reports child startup failures", async () => {
    const reportError = vi.fn();

    const code = await launchAgentEval({
      ...baseRuntime,
      reportError,
      runBinary: () =>
        Promise.resolve({
          kind: "failed-to-start",
          error: new Error("permission denied"),
        }),
    });

    expect(reportError).toHaveBeenCalledWith(
      "Failed to start: permission denied",
    );
    expect(code).toBe(1);
  });

  it("reports binary resolution failures without starting a child", async () => {
    const runBinary = vi.fn();
    const reportError = vi.fn();

    const code = await launchAgentEval({
      ...baseRuntime,
      currentPlatform: "aix",
      architecture: "ppc64",
      runBinary,
      reportError,
    });

    expect(runBinary).not.toHaveBeenCalled();
    expect(reportError).toHaveBeenCalledWith(
      "No binary available for aix-ppc64",
    );
    expect(code).toBe(1);
  });
});
