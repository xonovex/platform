interface ParsedArgs {
  readonly target: string | undefined;
  readonly all: string | undefined;
  readonly maxAge: number;
  readonly fetch: boolean;
  readonly markReviewed: string | undefined;
  readonly json: boolean;
  readonly pull: boolean;
  readonly help: boolean;
}

const isOptionToken = (token: string | undefined): boolean =>
  token !== undefined && token.startsWith("-") && token !== "-";

export const parseArgs = (argv: readonly string[]): ParsedArgs => {
  let all: string | undefined;
  let maxAge = 180;
  let fetch = false;
  let markReviewed: string | undefined;
  let json = false;
  let pull = false;
  let help = false;
  const positionals: string[] = [];

  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i];
    if (arg === undefined) continue;
    switch (arg) {
      case "-h":
      case "--help": {
        help = true;
        break;
      }
      case "--fetch": {
        fetch = true;
        break;
      }
      case "--pull": {
        pull = true;
        break;
      }
      case "--json": {
        json = true;
        break;
      }
      default: {
        if (arg === "--max-age" || arg.startsWith("--max-age=")) {
          const inline = arg.startsWith("--max-age=")
            ? arg.slice("--max-age=".length)
            : undefined;
          const raw = inline ?? argv[i + 1];
          if (inline === undefined) i += 1;
          if (raw === undefined || isOptionToken(raw)) {
            throw new Error("argument --max-age: expected one argument");
          }
          const value = Number(raw);
          if (!Number.isInteger(value) || value < 0) {
            throw new Error(
              `argument --max-age: invalid non-negative int value: '${raw}'`,
            );
          }
          maxAge = value;
        } else if (arg === "--all" || arg.startsWith("--all=")) {
          if (arg.startsWith("--all=")) {
            all = arg.slice("--all=".length);
          } else {
            const next = argv[i + 1];
            if (next !== undefined && !isOptionToken(next)) {
              all = next;
              i += 1;
            } else {
              all = ".";
            }
          }
        } else if (
          arg === "--mark-reviewed" ||
          arg.startsWith("--mark-reviewed=")
        ) {
          if (arg.startsWith("--mark-reviewed=")) {
            markReviewed = arg.slice("--mark-reviewed=".length);
          } else {
            const next = argv[i + 1];
            if (next !== undefined && !isOptionToken(next)) {
              markReviewed = next;
              i += 1;
            } else {
              markReviewed = "";
            }
          }
        } else if (isOptionToken(arg)) {
          throw new Error(`unrecognized arguments: ${arg}`);
        } else {
          positionals.push(arg);
        }
      }
    }
  }

  if (positionals.length > 1) {
    throw new Error(
      `unrecognized arguments: ${positionals.slice(1).join(" ")}`,
    );
  }
  return {
    target: positionals[0],
    all,
    maxAge,
    fetch,
    markReviewed,
    json,
    pull,
    help,
  };
};

export type {ParsedArgs};
