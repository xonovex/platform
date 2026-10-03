interface ParsedArgs {
  readonly target: string | undefined;
  readonly all: string | undefined;
  readonly maxAge: number;
  readonly versionMaxAge: number;
  readonly fetch: boolean;
  readonly markReviewed: string | undefined;
  readonly json: boolean;
  readonly pull: boolean;
  readonly help: boolean;
}

interface ValueOption {
  readonly name: "all" | "markReviewed" | "maxAge" | "versionMaxAge";
  readonly value: string;
  readonly nextIndex: number;
}

const isOptionToken = (token: string | undefined): boolean =>
  token !== undefined && token !== "-" && token.startsWith("-");

const optionalValue = (
  argv: readonly string[],
  index: number,
  inline: string | undefined,
  fallback: string,
): {readonly value: string; readonly nextIndex: number} => {
  if (inline !== undefined) return {value: inline, nextIndex: index};
  const next = argv[index + 1];
  return next !== undefined && !isOptionToken(next)
    ? {value: next, nextIndex: index + 1}
    : {value: fallback, nextIndex: index};
};

const valueOption = (
  argv: readonly string[],
  index: number,
  argument: string,
): ValueOption | undefined => {
  const equals = argument.indexOf("=");
  const flag = equals === -1 ? argument : argument.slice(0, equals);
  const inline = equals === -1 ? undefined : argument.slice(equals + 1);
  switch (flag) {
    case "--all": {
      const option = optionalValue(argv, index, inline, ".");
      return {name: "all", ...option};
    }
    case "--mark-reviewed": {
      const option = optionalValue(argv, index, inline, "");
      return {name: "markReviewed", ...option};
    }
    case "--max-age": {
      const raw = inline ?? argv[index + 1];
      if (raw === undefined || isOptionToken(raw)) {
        throw new Error("argument --max-age: expected one argument");
      }
      return {
        name: "maxAge",
        value: raw,
        nextIndex: inline === undefined ? index + 1 : index,
      };
    }
    case "--version-max-age": {
      const raw = inline ?? argv[index + 1];
      if (raw === undefined || isOptionToken(raw)) {
        throw new Error("argument --version-max-age: expected one argument");
      }
      return {
        name: "versionMaxAge",
        value: raw,
        nextIndex: inline === undefined ? index + 1 : index,
      };
    }
    default: {
      return undefined;
    }
  }
};

const parseMaxAge = (flag: string, raw: string): number => {
  const value = Number(raw);
  if (!Number.isInteger(value) || value < 0) {
    throw new Error(
      `argument ${flag}: invalid non-negative int value: '${raw}'`,
    );
  }
  return value;
};

export const parseArgs = (argv: readonly string[]): ParsedArgs => {
  let all: string | undefined;
  let maxAge = 180;
  let versionMaxAge = 90;
  let fetch = false;
  let markReviewed: string | undefined;
  let json = false;
  let pull = false;
  let help = false;
  const positionals: string[] = [];
  const consumed = new Set<number>();

  for (const [index, argument] of argv.entries()) {
    if (consumed.has(index)) continue;
    switch (argument) {
      case "-h":
      case "--help": {
        help = true;
        continue;
      }
      case "--fetch": {
        fetch = true;
        continue;
      }
      case "--pull": {
        pull = true;
        continue;
      }
      case "--json": {
        json = true;
        continue;
      }
    }

    const option = valueOption(argv, index, argument);
    if (option !== undefined) {
      if (option.nextIndex > index) consumed.add(option.nextIndex);
      if (option.name === "all") all = option.value;
      if (option.name === "markReviewed") markReviewed = option.value;
      if (option.name === "maxAge") {
        maxAge = parseMaxAge("--max-age", option.value);
      }
      if (option.name === "versionMaxAge") {
        versionMaxAge = parseMaxAge("--version-max-age", option.value);
      }
      continue;
    }
    if (isOptionToken(argument)) {
      throw new Error(`unrecognized arguments: ${argument}`);
    }
    positionals.push(argument);
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
    versionMaxAge,
    fetch,
    markReviewed,
    json,
    pull,
    help,
  };
};

export type {ParsedArgs};
