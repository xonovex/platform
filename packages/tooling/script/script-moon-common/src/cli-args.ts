import {parseArgs} from "node:util";

interface ArgOption {
  readonly type: "string" | "boolean";
  readonly short?: string;
  readonly default?: string | boolean;
  readonly description?: string;
}

interface CliSpec {
  readonly name: string;
  readonly description?: string;
  readonly options?: Readonly<Record<string, ArgOption>>;
}

interface ParsedArgs {
  readonly kind: "parsed";
  readonly values: Record<string, string | boolean | undefined>;
  readonly positionals: readonly string[];
}

interface HelpRequest {
  readonly kind: "help";
  readonly text: string;
}

// Help is a result rather than a side effect, so the entry point owns the
// output stream and the exit code.
type CliArgs = ParsedArgs | HelpRequest;

const formatHelp = (spec: CliSpec): string => {
  const lines = [
    `Usage: ${spec.name} [options]`,
    ...(spec.description ? [`\n${spec.description}`] : []),
    "\nOptions:",
    "  -h, --help  Show this help message",
    ...Object.entries(spec.options ?? {}).map(([name, opt]) => {
      const shortFlag = opt.short ? `-${opt.short}, ` : " ".repeat(4);
      const desc = opt.description ?? "";
      return `  ${shortFlag}--${name}  ${desc}`;
    }),
  ];
  return lines.join("\n");
};

const parseCliArgs = (spec: CliSpec, args: readonly string[]): CliArgs => {
  if (args.includes("--help") || args.includes("-h")) {
    return {kind: "help", text: formatHelp(spec)};
  }

  const options: Record<
    string,
    {type: "string" | "boolean"; short?: string; default?: string | boolean}
  > = {};
  if (spec.options) {
    for (const [name, opt] of Object.entries(spec.options)) {
      options[name] = {
        type: opt.type,
        ...(opt.short ? {short: opt.short} : {}),
        default: opt.default,
      };
    }
  }

  const result = parseArgs({
    args: [...args],
    options,
    strict: true,
    allowPositionals: true,
  });

  return {
    kind: "parsed",
    values: result.values,
    positionals: result.positionals,
  };
};

export {parseCliArgs, formatHelp};
export type {CliArgs, CliSpec, HelpRequest, ParsedArgs};
