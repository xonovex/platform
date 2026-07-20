import {describe, expect, it, vi} from "vitest";
import {
  childExitCode,
  findBinary,
  getBinaryName,
  getPlatformPackage,
  launchAgentCli,
} from "./launcher.js";

describe("getPlatformPackage", () => {
  it("selects the package for a supported platform", () => {
    expect(getPlatformPackage("linux", "x64")).toBe(
      "@xonovex/agent-cli-go-linux-x64",
    );
  });

  it("returns undefined for an unsupported architecture", () => {
    expect(getPlatformPackage("linux", "riscv64")).toBeUndefined();
  });
});

describe("getBinaryName", () => {
  it("selects the Windows executable suffix only for Windows", () => {
    expect(getBinaryName("win32")).toBe("agent-cli-go.exe");
    expect(getBinaryName("linux")).toBe("agent-cli-go");
  });
});

describe("findBinary", () => {
  it("resolves an installed platform binary", () => {
    const result = findBinary("linux", "x64", {
      resolvePackageJson: () => "/packages/linux-x64/package.json",
      binaryExists: () => true,
    });

    expect(result).toEqual({
      ok: true,
      binaryPath: "/packages/linux-x64/bin/agent-cli-go",
    });
  });

  it("reports unsupported platforms without resolving a package", () => {
    const result = findBinary("aix", "ppc64", {
      resolvePackageJson: () => {
        throw new Error("should not resolve");
      },
      binaryExists: () => false,
    });

    expect(result).toEqual({
      ok: false,
      error: "No binary available for aix-ppc64",
    });
  });

  it("reports a missing optional package", () => {
    const result = findBinary("linux", "x64", {
      resolvePackageJson: () => {
        throw new Error("package missing");
      },
      binaryExists: () => false,
    });

    expect(result.ok).toBe(false);
    if (result.ok) throw new Error("expected binary resolution to fail");
    expect(result.error).toContain("npm install -g");
  });

  it("reports a missing binary inside an installed package", () => {
    const result = findBinary("win32", "x64", {
      resolvePackageJson: () => String.raw`C:\packages\win32-x64\package.json`,
      binaryExists: () => false,
    });

    expect(result.ok).toBe(false);
    if (result.ok) throw new Error("expected binary resolution to fail");
    expect(result.error).toContain("Binary not found");
  });
});

describe("childExitCode", () => {
  it("preserves a normal child exit code", () => {
    expect(childExitCode(7, null)).toBe(7);
  });

  it("maps signal termination to a non-zero shell exit code", () => {
    expect(childExitCode(null, "SIGTERM")).toBe(143);
  });

  it("fails when the child closes without a code or signal", () => {
    expect(childExitCode(null, null)).toBe(1);
  });
});

describe("launchAgentCli", () => {
  it("starts the resolved binary and forwards its exit code", () => {
    const exit = vi.fn();
    let closeHandler:
      | ((code: number | null, signal: NodeJS.Signals | null) => void)
      | undefined;
    const startBinary = vi.fn(
      (
        _binaryPath: string,
        _arguments: readonly string[],
        _environment: NodeJS.ProcessEnv,
        _onError: (error: Error) => void,
        onClose: (code: number | null, signal: NodeJS.Signals | null) => void,
      ) => {
        closeHandler = onClose;
      },
    );

    launchAgentCli({
      currentPlatform: "linux",
      architecture: "x64",
      arguments: ["run", "task"],
      environment: {TEST_ENV: "enabled"},
      resolvePackageJson: () => "/packages/linux-x64/package.json",
      binaryExists: () => true,
      startBinary,
      reportError: vi.fn(),
      exit,
    });
    closeHandler?.(7, null);

    expect(startBinary).toHaveBeenCalledWith(
      "/packages/linux-x64/bin/agent-cli-go",
      ["run", "task"],
      {TEST_ENV: "enabled"},
      expect.any(Function),
      expect.any(Function),
    );
    expect(exit).toHaveBeenCalledWith(7);
  });

  it("reports child startup failures", () => {
    const reportError = vi.fn();
    const exit = vi.fn();

    launchAgentCli({
      currentPlatform: "linux",
      architecture: "x64",
      arguments: [],
      environment: {},
      resolvePackageJson: () => "/packages/linux-x64/package.json",
      binaryExists: () => true,
      startBinary: (_binaryPath, _arguments, _environment, onError) => {
        onError(new Error("permission denied"));
      },
      reportError,
      exit,
    });

    expect(reportError).toHaveBeenCalledWith(
      "Failed to start: permission denied",
    );
    expect(exit).toHaveBeenCalledWith(1);
  });

  it("reports binary resolution failures without starting a child", () => {
    const startBinary = vi.fn();
    const reportError = vi.fn();
    const exit = vi.fn();

    launchAgentCli({
      currentPlatform: "aix",
      architecture: "ppc64",
      arguments: [],
      environment: {},
      resolvePackageJson: vi.fn(),
      binaryExists: vi.fn(),
      startBinary,
      reportError,
      exit,
    });

    expect(startBinary).not.toHaveBeenCalled();
    expect(reportError).toHaveBeenCalledWith(
      "No binary available for aix-ppc64",
    );
    expect(exit).toHaveBeenCalledWith(1);
  });
});
