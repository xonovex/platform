import {describe, expect, it} from "vitest";
import {childExitCode, getBinaryName, getPlatformPackage} from "./launcher.js";

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
