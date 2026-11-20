import {skillNameMentions} from "@xonovex/script-moon-common/skill-name";

const NEGATIVE_ROUTING_CLAUSE_RE =
  /(?:^|[.!?]\s+)(?:skip\b|do not use\b|don't use\b|out[- ]of[- ]scope\b)/i;

export const descriptionRoutingErrors = (
  description: string,
  ownSkillName: string | undefined,
): readonly string[] => {
  const errors: string[] = [];
  const normalizedOwnSkillName = ownSkillName?.toLowerCase();
  if (NEGATIVE_ROUTING_CLAUSE_RE.test(description)) {
    errors.push(
      "description: routing must use positive triggers only; move skip/out-of-scope guidance to the body",
    );
  }
  const namedSkills = [
    ...new Set(
      skillNameMentions(description).filter(
        (name) => name !== normalizedOwnSkillName,
      ),
    ),
  ].toSorted();
  if (namedSkills.length > 0) {
    errors.push(
      `description: names other skill(s) ${namedSkills.join(", ")}; put by-name handoffs in the body`,
    );
  }
  return errors;
};
