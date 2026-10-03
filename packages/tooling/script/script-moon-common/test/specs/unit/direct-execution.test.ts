import {resolve} from "node:path";
import {pathToFileURL} from "node:url";
import {describe, expect, it} from "vitest";
import {isDirectExecution} from "../../../src/direct-execution.js";

describe("isDirectExecution", () => {
  it("matches the module file to the process entrypoint", () => {
    const entrypoint = resolve("example-cli.js");
    expect(isDirectExecution(pathToFileURL(entrypoint).href, entrypoint)).toBe(
      true,
    );
  });

  it("rejects imports and missing process entrypoints", () => {
    const moduleUrl = pathToFileURL(resolve("example-cli.js")).href;
    expect(isDirectExecution(moduleUrl, resolve("test.js"))).toBe(false);
    expect(isDirectExecution(moduleUrl, undefined)).toBe(false);
  });
});
