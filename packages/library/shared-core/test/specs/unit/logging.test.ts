import {describe, expect, it} from "vitest";
import {Colors} from "../../../src/colors.js";
import {
  createLogger,
  type CheckStatus,
  type Logger,
} from "../../../src/logging.js";

interface Recorded {
  readonly stream: "out" | "err";
  readonly args: readonly unknown[];
}

function recordingLogger(debug = true): {
  logger: Logger;
  lines: Recorded[];
} {
  const lines: Recorded[] = [];
  const logger = createLogger({
    sink: {
      out: (...args) => {
        lines.push({stream: "out", args});
      },
      err: (...args) => {
        lines.push({stream: "err", args});
      },
    },
    debug,
  });
  return {logger, lines};
}

describe("createLogger", () => {
  it.each([
    ["info", Colors.BLUE, "INFO", "out"],
    ["success", Colors.GREEN, "SUCCESS", "out"],
    ["warning", Colors.YELLOW, "WARNING", "out"],
    ["error", Colors.RED, "ERROR", "err"],
    ["debug", Colors.PURPLE, "DEBUG", "err"],
  ] as const)(
    "writes %s with its colored tag to the %s stream",
    (method, color, label, stream) => {
      const {logger, lines} = recordingLogger();
      logger[method]("test", "multiple", 1);

      expect(lines).toEqual([
        {
          stream,
          args: [`${color}[${label}]${Colors.NC}`, "test", "multiple", 1],
        },
      ]);
    },
  );

  it("drops debug messages when the debug gate is closed", () => {
    const {logger, lines} = recordingLogger(false);
    logger.debug("hidden");

    expect(lines).toEqual([]);
  });

  it("prints a section with its content", () => {
    const {logger, lines} = recordingLogger();
    logger.printSection("Title", "content");

    expect(lines.map((line) => line.args)).toEqual([
      [`\n${Colors.CYAN}Title${Colors.NC}`],
      [`${Colors.CYAN}=====${Colors.NC}`],
      ["content"],
    ]);
  });

  it("prints a section without content", () => {
    const {logger, lines} = recordingLogger();
    logger.printSection("Title");

    expect(lines).toHaveLength(2);
  });

  it("prints a subsection with an underline", () => {
    const {logger, lines} = recordingLogger();
    logger.printSubsection("Sub");

    expect(lines.map((line) => line.args)).toEqual([
      [`\n${Colors.BLUE}Sub${Colors.NC}`],
      [`${Colors.BLUE}---${Colors.NC}`],
    ]);
  });

  it("prints an empty subsection title without an underline", () => {
    const {logger, lines} = recordingLogger();
    logger.printSubsection("");

    expect(lines).toHaveLength(1);
  });

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
  ] as const satisfies readonly (readonly [CheckStatus, string, string])[])(
    "styles a %s check result",
    (status, symbol, color) => {
      const {logger, lines} = recordingLogger();
      logger.checkResult("my check", status);

      expect(lines).toEqual([
        {stream: "out", args: [`${color}${symbol} my check${Colors.NC}`]},
      ]);
    },
  );

  it("prints check details in purple", () => {
    const {logger, lines} = recordingLogger();
    logger.checkResult("my check", "PASS", "some details");

    expect(lines[1]?.args).toEqual([
      `${Colors.PURPLE}    some details${Colors.NC}`,
    ]);
  });
});
