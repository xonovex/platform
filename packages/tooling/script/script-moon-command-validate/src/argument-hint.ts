export interface CommandArgument {
  readonly kind: "flag" | "positional";
  readonly name: string;
  readonly required: boolean;
  readonly repeatable: boolean;
  readonly valueName?: string;
}

interface HintToken {
  readonly depth: number;
  readonly end: number;
  readonly start: number;
  readonly text: string;
}

const TOKEN_RE =
  /\[|\]|--[a-z0-9]+(?:-[a-z0-9]+)*|<[^>\s]+>(?:\.\.\.)?|[a-z0-9]+(?:-[a-z0-9]+)*(?:\.\.\.)?/giu;

const tokenize = (hint: string): readonly HintToken[] => {
  const tokens: HintToken[] = [];
  let depth = 0;
  for (const match of hint.matchAll(TOKEN_RE)) {
    const text = match[0];
    const start = match.index;
    if (text === "[") {
      depth += 1;
      continue;
    }
    if (text === "]") {
      depth = Math.max(0, depth - 1);
      continue;
    }
    tokens.push({depth, end: start + text.length, start, text});
  }
  return tokens;
};

const placeholderName = (token: string): string =>
  token
    .replace(/^</u, "")
    .replace(/>(?:\.\.\.)?$/u, "")
    .replace(/\.\.\.$/u, "");

export const parseArgumentHint = (hint: string): readonly CommandArgument[] => {
  const tokens = tokenize(hint);
  const argumentsFound: CommandArgument[] = [];
  const consumed = new Set<number>();

  for (const [index, token] of tokens.entries()) {
    if (consumed.has(index)) continue;
    if (token.text.startsWith("--")) {
      const next = tokens[index + 1];
      const hasValue =
        next !== undefined &&
        next.text.startsWith("<") &&
        !hint.slice(token.end, next.start).includes("]");
      if (hasValue) consumed.add(index + 1);
      argumentsFound.push({
        kind: "flag",
        name: token.text.slice(2),
        required: token.depth === 0,
        repeatable:
          (hasValue && next.text.endsWith("...")) ||
          hint
            .slice(token.end, hasValue ? next.end : token.end + 3)
            .includes("..."),
        ...(hasValue ? {valueName: placeholderName(next.text)} : {}),
      });
      continue;
    }
    if (token.text.startsWith("<") || /^[a-z0-9]/iu.test(token.text)) {
      argumentsFound.push({
        kind: "positional",
        name: placeholderName(token.text),
        required: token.depth === 0,
        repeatable: token.text.endsWith("..."),
        valueName: placeholderName(token.text),
      });
    }
  }

  return argumentsFound;
};
