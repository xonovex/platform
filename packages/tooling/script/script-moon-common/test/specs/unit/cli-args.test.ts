import {describe, expect, it} from "vitest";
import {
  formatHelp,
  parseCliArgs,
  type CliArgs,
  type CliSpec,
  type ParsedArgs,
} from "../../../src/cli-args.js";

const spec: CliSpec = {
  name: "test-cli",
  description: "A test CLI tool",
  options: {
    name: {type: "string", short: "n", default: "world", description: "Name"},
    verbose: {type: "boolean", short: "v", description: "Verbose output"},
  },
};

const parsed = (result: CliArgs): ParsedArgs => {
  if (result.kind !== "parsed") throw new Error("expected parsed arguments");
  return result;
};

describe("parseCliArgs", () => {
  it("should parse named string and boolean options", () => {
    const result = parsed(parseCliArgs(spec, ["--name", "alice", "--verbose"]));
    expect(result.values.name).toBe("alice");
    expect(result.values.verbose).toBe(true);
  });

  it("should use defaults when options are not provided", () => {
    const result = parsed(parseCliArgs(spec, []));
    expect(result.values.name).toBe("world");
    expect(result.values.verbose).toBeUndefined();
  });

  it("should return positionals", () => {
    const result = parsed(parseCliArgs(spec, ["foo", "bar"]));
    expect(result.positionals).toEqual(["foo", "bar"]);
  });

  it("should return the help text on --help", () => {
    expect(parseCliArgs(spec, ["--help"])).toEqual({
      kind: "help",
      text: formatHelp(spec),
    });
  });

  it("should return the help text on -h", () => {
    expect(parseCliArgs(spec, ["-h"]).kind).toBe("help");
  });

  it("should throw on unknown flags", () => {
    expect(() => parseCliArgs(spec, ["--unknown"])).toThrow();
  });

  it("should parse a short option", () => {
    const result = parsed(parseCliArgs(spec, ["-n", "custom"]));
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
