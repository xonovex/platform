// splitLines accepts every common newline and omits a terminal empty line.
export const splitLines = (text: string): string[] => {
  if (text.length === 0) return [];
  const parts = text.split(/\r\n|\r|\n/);
  if (parts.length > 0 && parts.at(-1) === "") parts.pop();
  return parts;
};

// A skill name and a reference file name share the kebab-case rule.
export const KEBAB_CASE_RE = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
