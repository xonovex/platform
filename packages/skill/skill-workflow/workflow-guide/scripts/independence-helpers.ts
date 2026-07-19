// The declared authority for an independence requirement is the profile, never
// the record under scrutiny: a record cannot elect, waive, or weaken the check
// applied to it. checkIndependence therefore reads `required` from the profile
// and ignores any independence field the record carries.
export const independenceLevels = [
  "none",
  "distinct-identity",
  "distinct-team",
  "distinct-organization",
];

// `distinct-team` and `distinct-organization` demand reporting-line or
// legal-entity separation on top of identity. That separation is not expressible
// in the actor record — it carries no team or organization field — and is not
// resolvable here: membership and legal entity are provider state. For these
// levels the identity inequality is insufficient by construction, so the record
// must carry a reference to the provider evidence proving the provider checked
// the separation. Absent or empty, the check fails closed rather than silently
// downgrading to the identity comparison alone.
const providerEnforcedLevels = ["distinct-team", "distinct-organization"];

const isNonEmptyString = (value) =>
  typeof value === "string" && value.length > 0;

// Every level above `none` requires the deciding actor to differ from the
// subject's author by identity. `distinct-team` and `distinct-organization`
// additionally require a recorded provider-evidence reference that the provider
// checked the reporting-line or legal-entity separation; checkIndependence
// enforces the identity component of every level and, for the stronger two, that
// the provider's separation evidence was recorded — never the separation itself,
// which stays provider state. The self-grant inequality is evaluated before the
// evidence requirement, so an author clearing their own subject is caught at any
// level regardless of whether evidence is present.
export const checkIndependence = ({
  required,
  decider,
  author,
  providerEvidence,
  failureCode,
}) => {
  if (!independenceLevels.includes(required)) {
    return "independence-requirement-undeclared";
  }
  if (required === "none") return null;
  if (!isNonEmptyString(decider) || !isNonEmptyString(author)) {
    return "independence-unverifiable";
  }
  if (decider === author) return failureCode;
  if (
    providerEnforcedLevels.includes(required) &&
    !isNonEmptyString(providerEvidence)
  ) {
    return "independence-provider-evidence-required";
  }
  return null;
};
