export interface Checker {
  check: (condition: boolean, message: string) => void;
  checks: () => number;
  failures: () => readonly string[];
}

export const createChecker = (): Checker => {
  const failures: string[] = [];
  let checks = 0;
  return {
    check: (condition, message) => {
      checks += 1;
      if (!condition) failures.push(message);
    },
    checks: () => checks,
    failures: () => failures,
  };
};

export const tableIds = (content: string, prefix: string): Set<string> =>
  new Set(
    [
      ...content.matchAll(
        new RegExp(String.raw`^\|\s*(${prefix}[^|\s]+)\s*\|`, "gm"),
      ),
    ].flatMap((match) => (match[1] === undefined ? [] : [match[1]])),
  );

export const markdownLinkTargets = (content: string): string[] =>
  [...content.matchAll(/\[[^\]]+\]\(([^)]+)\)/g)]
    .flatMap((match) => {
      const target = match[1]?.split("#", 1)[0];
      return target === undefined ? [] : [target];
    })
    .filter((target) => target !== "" && !/^(?:https?:|mailto:)/.test(target));

export const forbiddenClaims: readonly RegExp[] = [
  /all harnesses (?:have|support|provide)/i,
  /skills? (?:are|is|provide|provides) enforcement/i,
  /install(?:ing|ed)? (?:a )?skills? (?:enforces|proves)/i,
  /workflow ya?ml (?:is|are) required/i,
  /(?:provides?|ensures?|establishes?|achieves?) automatic compliance/i,
  /silently launch(?:es|ing)? (?:a |an )?(?:child )?agent/i,
];
