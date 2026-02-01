import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  vi,
  type MockInstance,
} from "vitest";
import {formatHelp, parseCliArgs, type CliSpec} from "./cli-args.js";

const spec: CliSpec = {
  name: "test-cli",
  description: "A test CLI tool",
  options: {
    name: {type: "string", short: "n", default: "world", description: "Name"},
    verbose: {type: "boolean", short: "v", description: "Verbose output"},
  },
};

describe("parseCliArgs", () => {
  let consoleLogSpy: MockInstance<typeof console.log>;
  let processExitSpy: MockInstance<typeof process.exit>;

  beforeEach(() => {
    consoleLogSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    processExitSpy = vi
      .spyOn(process, "exit")
      .mockImplementation(() => undefined as never);
  });

  afterEach(() => {
    consoleLogSpy.mockRestore();
    processExitSpy.mockRestore();
  });

  it("should parse named string and boolean options", () => {
    const result = parseCliArgs(spec, ["--name", "alice", "--verbose"]);
    expect(result.values.name).toBe("alice");
    expect(result.values.verbose).toBe(true);
  });

  it("should use defaults when options are not provided", () => {
    const result = parseCliArgs(spec, []);
    expect(result.values.name).toBe("world");
    expect(result.values.verbose).toBeUndefined();
  });

  it("should return positionals", () => {
    const result = parseCliArgs(spec, ["foo", "bar"]);
    expect(result.positionals).toEqual(["foo", "bar"]);
  });

  it("should print help and exit on --help", () => {
    parseCliArgs(spec, ["--help"]);
    expect(consoleLogSpy).toHaveBeenCalledWith(
      expect.stringContaining("test-cli"),
    );
    expect(processExitSpy).toHaveBeenCalledWith(0);
  });

  it("should print help and exit on -h", () => {
    parseCliArgs(spec, ["-h"]);
    expect(processExitSpy).toHaveBeenCalledWith(0);
  });

  it("should throw on unknown flags", () => {
    expect(() => parseCliArgs(spec, ["--unknown"])).toThrow();
  });

  it("should use custom argv parameter", () => {
    const result = parseCliArgs(spec, ["-n", "custom"]);
    expect(result.values.name).toBe("custom");
  });
});

describe("formatHelp", () => {
  it("should include spec name and description", () => {
    const help = formatHelp(spec);
    expect(help).toContain("test-cli");
    expect(help).toContain("A test CLI tool");
  });

  it("should include option details", () => {
    const help = formatHelp(spec);
    expect(help).toContain("--name");
    expect(help).toContain("-n");
    expect(help).toContain("--verbose");
  });
});
