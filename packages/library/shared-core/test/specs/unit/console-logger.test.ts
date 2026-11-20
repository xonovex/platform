import {describe, expect, it} from "vitest";
import {Colors} from "../../../src/colors.js";
import {createConsoleLogger} from "../../../src/console-logger.js";

function fakeConsole(): {
  console: {
    log: (...args: unknown[]) => void;
    error: (...args: unknown[]) => void;
  };
  logged: unknown[][];
  errors: unknown[][];
} {
  const logged: unknown[][] = [];
  const errors: unknown[][] = [];
  return {
    console: {
      log: (...args) => {
        logged.push(args);
      },
      error: (...args) => {
        errors.push(args);
      },
    },
    logged,
    errors,
  };
}

describe("createConsoleLogger", () => {
  it("writes out lines to console.log and error lines to console.error", () => {
    const {console, logged, errors} = fakeConsole();
    const logger = createConsoleLogger(console, {});

    logger.info("a");
    logger.error("b");

    expect(logged).toEqual([[`${Colors.BLUE}[INFO]${Colors.NC}`, "a"]]);
    expect(errors).toEqual([[`${Colors.RED}[ERROR]${Colors.NC}`, "b"]]);
  });

  it("enables debug output when DEBUG is set", () => {
    const {console, errors} = fakeConsole();
    createConsoleLogger(console, {DEBUG: "1"}).debug("on");

    expect(errors).toHaveLength(1);
  });

  it("drops debug output when DEBUG is unset or empty", () => {
    const {console, errors} = fakeConsole();
    createConsoleLogger(console, {}).debug("off");
    createConsoleLogger(console, {DEBUG: ""}).debug("off");

    expect(errors).toEqual([]);
  });
});
