import {
  afterEach,
  beforeEach,
  describe,
  expect,
  it,
  vi,
  type MockInstance,
} from "vitest";
import {Colors} from "../../../src/colors.js";

describe("Logging", () => {
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
    it("should log with INFO prefix and blue color", async () => {
      const {logInfo} = await import("../../../src/logging.js");
      logInfo("test message");

      expect(consoleLogSpy).toHaveBeenCalledWith(
        `${Colors.BLUE}[INFO]${Colors.NC}`,
        "test message",
      );
    });

    it("should handle multiple arguments", async () => {
      const {logInfo} = await import("../../../src/logging.js");
      logInfo("test", "multiple", "args");

      expect(consoleLogSpy).toHaveBeenCalledWith(
        `${Colors.BLUE}[INFO]${Colors.NC}`,
        "test",
        "multiple",
        "args",
      );
    });
  });

  describe("logSuccess", () => {
    it("should log with SUCCESS prefix and green color", async () => {
      const {logSuccess} = await import("../../../src/logging.js");
      logSuccess("test success");

      expect(consoleLogSpy).toHaveBeenCalledWith(
        `${Colors.GREEN}[SUCCESS]${Colors.NC}`,
        "test success",
      );
    });
  });

  describe("logWarning", () => {
    it("should log with WARNING prefix and yellow color", async () => {
      const {logWarning} = await import("../../../src/logging.js");
      logWarning("test warning");

      expect(consoleLogSpy).toHaveBeenCalledWith(
        `${Colors.YELLOW}[WARNING]${Colors.NC}`,
        "test warning",
      );
    });
  });

  describe("logError", () => {
    it("should log to stderr with ERROR prefix and red color", async () => {
      const {logError} = await import("../../../src/logging.js");
      logError("test error");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        `${Colors.RED}[ERROR]${Colors.NC}`,
        "test error",
      );
    });
  });

  describe("logDebug", () => {
    it("should log when DEBUG env var is set", async () => {
      const originalDebug = process.env.DEBUG;
      process.env.DEBUG = "1";

      // Need to reimport to get fresh module with new env var
      const {logDebug} = await import("../../../src/logging.js");
      logDebug("debug message");

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        `${Colors.PURPLE}[DEBUG]${Colors.NC}`,
        "debug message",
      );

      if (originalDebug === undefined) {
        delete process.env.DEBUG;
      } else {
        process.env.DEBUG = originalDebug;
      }
    });

    it("should not log when DEBUG env var is not set", async () => {
      const originalDebug = process.env.DEBUG;
      delete process.env.DEBUG;

      const {logDebug} = await import("../../../src/logging.js");
      consoleErrorSpy.mockClear();
      logDebug("debug message");

      expect(consoleErrorSpy).not.toHaveBeenCalled();

      if (originalDebug === undefined) {
        delete process.env.DEBUG;
      } else {
        process.env.DEBUG = originalDebug;
      }
    });
  });

  describe("printSection", () => {
    it("should print section with title and separator", async () => {
      const {printSection} = await import("../../../src/logging.js");
      printSection("Test Section");

      expect(consoleLogSpy).toHaveBeenCalledWith(
        `\n${Colors.CYAN}Test Section${Colors.NC}`,
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        `${Colors.CYAN}${"=".repeat(12)}${Colors.NC}`,
      );
    });

    it("should print content if provided", async () => {
      const {printSection} = await import("../../../src/logging.js");
      printSection("Test", "Some content");

      expect(consoleLogSpy).toHaveBeenCalledWith("Some content");
    });
  });

  describe("printSubsection", () => {
    it("should print subsection with title and dashes", async () => {
      const {printSubsection} = await import("../../../src/logging.js");
      printSubsection("Sub Title");

      expect(consoleLogSpy).toHaveBeenCalledWith(
        `\n${Colors.BLUE}Sub Title${Colors.NC}`,
      );
      expect(consoleLogSpy).toHaveBeenCalledWith(
        `${Colors.BLUE}${"-".repeat(9)}${Colors.NC}`,
      );
    });

    it("should omit the separator for an empty title", async () => {
      const {printSubsection} = await import("../../../src/logging.js");
      printSubsection("");

      expect(consoleLogSpy).toHaveBeenCalledExactlyOnceWith(
        `\n${Colors.BLUE}${Colors.NC}`,
      );
    });
  });

  describe("checkResult", () => {
    it.each([
      ["PASS", "✓", Colors.GREEN],
      ["OK", "✓", Colors.GREEN],
      ["SUCCESS", "✓", Colors.GREEN],
      ["FAIL", "✗", Colors.RED],
      ["ERROR", "✗", Colors.RED],
      ["FAILED", "✗", Colors.RED],
      ["WARN", "⚠", Colors.YELLOW],
      ["WARNING", "⚠", Colors.YELLOW],
      ["INFO", "ℹ", Colors.BLUE],
    ] as const)(
      "should show %s as %s in its status color",
      async (status, symbol, color) => {
        const {checkResult} = await import("../../../src/logging.js");
        checkResult("Test check", status);

        expect(consoleLogSpy).toHaveBeenCalledWith(
          `${color}${symbol} Test check${Colors.NC}`,
        );
      },
    );

    it("should print details if provided", async () => {
      const {checkResult} = await import("../../../src/logging.js");
      checkResult("Test check", "SUCCESS", "Additional details");

      expect(consoleLogSpy).toHaveBeenCalledWith(
        expect.stringContaining("Additional details"),
      );
    });
  });
});
