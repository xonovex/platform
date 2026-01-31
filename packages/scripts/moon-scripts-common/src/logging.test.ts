import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  vi,
  type MockInstance,
} from "vitest";
import {logError, logInfo, logSuccess, logWarning} from "./logging.js";

describe("logging", () => {
  let consoleLogSpy: MockInstance<typeof console.log>;
  let consoleErrorSpy: MockInstance<typeof console.error>;

  beforeEach(() => {
    consoleLogSpy = vi.spyOn(console, "log").mockImplementation(() => {});
    consoleErrorSpy = vi.spyOn(console, "error").mockImplementation(() => {});
  });

  afterEach(() => {
    consoleLogSpy.mockRestore();
    consoleErrorSpy.mockRestore();
  });

  describe("logInfo", () => {
    it("should log with INFO prefix to stdout", () => {
      logInfo("test message");
      expect(consoleLogSpy).toHaveBeenCalledWith(
        expect.stringContaining("[INFO]"),
        "test message",
      );
    });

    it("should handle multiple arguments", () => {
      logInfo("a", "b", "c");
      expect(consoleLogSpy).toHaveBeenCalledWith(
        expect.stringContaining("[INFO]"),
        "a",
        "b",
        "c",
      );
    });
  });

  describe("logSuccess", () => {
    it("should log with SUCCESS prefix to stdout", () => {
      logSuccess("done");
      expect(consoleLogSpy).toHaveBeenCalledWith(
        expect.stringContaining("[SUCCESS]"),
        "done",
      );
    });
  });

  describe("logWarning", () => {
    it("should log with WARNING prefix to stdout", () => {
      logWarning("caution");
      expect(consoleLogSpy).toHaveBeenCalledWith(
        expect.stringContaining("[WARNING]"),
        "caution",
      );
    });
  });

  describe("logError", () => {
    it("should log with ERROR prefix to stderr", () => {
      logError("failure");
      expect(consoleErrorSpy).toHaveBeenCalledWith(
        expect.stringContaining("[ERROR]"),
        "failure",
      );
    });
  });
});
