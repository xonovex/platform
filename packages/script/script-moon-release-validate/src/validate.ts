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
  /(?:provides?|ensures?|establishes?|achieves?) automatic compliance/i,
  /silently launch(?:es|ing)? (?:a |an )?(?:child )?agent/i,
];

export const releaseWorkflowFailures = (content: string): string[] => {
  const failures: string[] = [];
  if (/^\s{2}push:\s*$/m.test(content) || content.includes("refs/tags/")) {
    failures.push("release workflow must not publish from pushed tags");
  }
  if (/inputs\.dry_run\s*!=\s*true/.test(content)) {
    failures.push(
      "release workflow must not derive publish authority from a dispatch input",
    );
  }

  const publishStep =
    /\n\s+- name: Publish\n([\s\S]*?)(?=\n\s+- (?:name:|uses:))/u.exec(
      content,
    )?.[1];
  if (
    publishStep === undefined ||
    !/if:\s*github\.event_name == 'pull_request'/.test(publishStep)
  ) {
    failures.push("publish step must be restricted to the pull_request event");
  }

  const dryRunStep =
    /\n\s+- name: Publish \(dry run\)\n([\s\S]*?)(?=\n\s+- (?:name:|uses:))/u.exec(
      content,
    )?.[1];
  if (
    dryRunStep === undefined ||
    !/if:\s*github\.event_name == 'workflow_dispatch'/.test(dryRunStep)
  ) {
    failures.push(
      "manual dispatch must be restricted to the dry-run publish step",
    );
  }

  return failures;
};
