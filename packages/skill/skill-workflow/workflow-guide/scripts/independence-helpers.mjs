// The declared authority for an independence requirement is the profile, never
// the record under scrutiny: a record cannot elect, waive, or weaken the check
// applied to it. checkIndependence therefore reads `required` from the profile
// and ignores any independence field the record carries.
const independenceLevels = [
  "none",
  "distinct-identity",
  "distinct-team",
  "distinct-organization",
];

const isIdentity = (value) => typeof value === "string" && value.length > 0;

// Every level above `none` requires the deciding actor to differ from the
// subject's author by identity. `distinct-team` and `distinct-organization`
// additionally require reporting-line or legal-entity separation, which the
// actor record cannot express; checkIndependence enforces only the identity
// component of every level and leaves the rest to the profile's provider.
export const checkIndependence = ({required, decider, author, failureCode}) => {
  if (!independenceLevels.includes(required)) {
    return "independence-requirement-undeclared";
  }
  if (required === "none") return null;
  if (!isIdentity(decider) || !isIdentity(author)) {
    return "independence-unverifiable";
  }
  return decider === author ? failureCode : null;
};
