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
  readonly values: Record<string, string | boolean | undefined>;
  readonly positionals: readonly string[];
}

const formatHelp = (spec: CliSpec): string => {
  const lines = [
    `Usage: ${spec.name} [options]`,
    ...(spec.description ? [`\n${spec.description}`] : []),
    "\nOptions:",
    "  -h, --help  Show this help message",
    ...Object.entries(spec.options ?? {}).map(([name, opt]) => {
      const shortFlag = opt.short ? `-${opt.short}, ` : "    ";
      const desc = opt.description ?? "";
      return `  ${shortFlag}--${name}  ${desc}`;
    }),
  ];
  return lines.join("\n");
};

const parseCliArgs = (spec: CliSpec, argv?: readonly string[]): ParsedArgs => {
  const args = argv ?? process.argv.slice(2);

  if (args.includes("--help") || args.includes("-h")) {
    console.log(formatHelp(spec));
    // eslint-disable-next-line unicorn/no-process-exit
    process.exit(0);
    return {values: {}, positionals: []} as ParsedArgs;
  }

  const options: Record<
    string,
    {type: "string" | "boolean"; short?: string; default?: string | boolean}
  > = {};
  if (spec.options) {
    for (const [name, opt] of Object.entries(spec.options)) {
      options[name] = {type: opt.type, short: opt.short, default: opt.default};
    }
  }

  const result = parseArgs({
    args: [...args],
    options,
    strict: true,
    allowPositionals: true,
  });

  return {
    values: result.values as Record<string, string | boolean | undefined>,
    positionals: result.positionals,
  };
};

export {parseCliArgs, formatHelp};
export type {ArgOption, CliSpec, ParsedArgs};
