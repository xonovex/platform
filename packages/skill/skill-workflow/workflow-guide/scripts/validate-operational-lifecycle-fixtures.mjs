import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";
import {
  validateAcceptance,
  validateAgentAssistance,
  validateAuthorization,
  validateCorrectiveAction,
  validateEmergencyAccess,
  validateIncident,
  validateObservation,
  validatePrivilegedOperation,
  validateRelease,
  validateRetirement,
  validateTransition,
} from "./operational-lifecycle-helpers.ts";

const fixtureUrl = new URL(
  "../assets/operational-lifecycle-fixtures.json",
  import.meta.url,
);

const isRecord = (value) =>
  value !== null && typeof value === "object" && !Array.isArray(value);

const mergeFixture = (base, override) => {
  if (!isRecord(base) || !isRecord(override)) {
    return override === undefined ? base : override;
  }
  return Object.fromEntries(
    [...new Set([...Object.keys(base), ...Object.keys(override)])].map(
      (key) => [key, mergeFixture(base[key], override[key])],
    ),
  );
};

// The profile is resolved from the case's own `profileTemplate`/`profile` keys
// and handed to each validator beside the record, never out of it: a case that
// hides a profile inside `input` is passing a record field, and no validator
// reads one. A case that declares no profile resolves to `{}`, which every
// independence check rejects as undeclared.
const resolveProfile = (testCase, templates) =>
  mergeFixture(templates[testCase.profileTemplate] ?? {}, testCase.profile);

const validateCase = (testCase, templates) => {
  const input = mergeFixture(
    templates[testCase.template] ?? {},
    testCase.input,
  );
  if (testCase.contract === "acceptance") {
    return validateAcceptance({
      acceptance: input,
      profile: resolveProfile(testCase, templates),
    });
  }
  if (testCase.contract === "authorization") {
    return validateAuthorization(input);
  }
  if (testCase.contract === "emergency-access") {
    return validateEmergencyAccess({
      access: input.access,
      request: input.request,
      profile: resolveProfile(testCase, templates),
    });
  }
  if (testCase.contract === "privileged-operation") {
    return validatePrivilegedOperation({
      operation: input,
      profile: resolveProfile(testCase, templates),
    });
  }
  if (testCase.contract === "transition") {
    return validateTransition({
      transition: input,
      profile: resolveProfile(testCase, templates),
    });
  }
  if (testCase.contract === "release") {
    return validateRelease({
      release: input,
      profile: resolveProfile(testCase, templates),
    });
  }
  if (testCase.contract === "observation") {
    return validateObservation(input);
  }
  if (testCase.contract === "incident") {
    return validateIncident(input);
  }
  if (testCase.contract === "corrective-action") {
    return validateCorrectiveAction(input);
  }
  if (testCase.contract === "retirement") {
    return validateRetirement({
      retirement: input,
      profile: resolveProfile(testCase, templates),
    });
  }
  if (testCase.contract === "agent-assistance") {
    return validateAgentAssistance(input);
  }
  return `unknown-contract:${testCase.contract}`;
};

// Acceptance and emergency-access independence are resolved from the profile,
// so these cases prove a record cannot elect the check applied to it. Deleting
// one would silently reopen the self-approval path; requiredCaseIds keeps them
// under test. The emergency-access set covers both kinds because the pair it
// separates — approving an exception and relying on it — does not vary by kind,
// and covers an operation that carries its own profile because the emergency
// access check is reached through validatePrivilegedOperation as well.
//
// `accountable` is spelled out separately at each site that records a decision
// actor, so one negative case per site keeps the sites from forking: a template
// carrying `accountable: true` passes whether or not the site still reads the
// field. The emergency-access cases cover both a false and an absent value, and
// both kinds, because the approver contract does not vary by kind.
//
// The emergency-access owner is a party reference, not a bare present value, so
// four cases keep the shape under test: a non-party string (no `<kind>:<id>`
// structure), an unknown-kind reference (`agent:` — an advisory executor kind
// the party set excludes, exercised on a plain exception to prove the guard
// does not vary by kind), a non-string object, and a non-person `team:` owner
// the set accepts (pinning that the party set is not narrowed to people).
// Deleting a rejection case would silently reopen an owner the self-grant
// comparison cannot read; deleting the acceptance case would let the set be
// narrowed away from the parties the record legitimately ranges over.
//
// `distinct-team` and `distinct-organization` cannot be satisfied by the
// identity comparison alone, so each independence-carrying contract keeps three
// cases proving the provider-evidence requirement: a distinct decider with the
// evidence reference recorded passes, the same record without the evidence fails
// closed as `independence-provider-evidence-required` rather than downgrading to
// identity-only, and a self-grant is still caught as the contract's
// `*-independence-failed` code even when the evidence is present — pinning that
// the identity inequality is evaluated before the evidence requirement. Deleting
// the without-evidence case would reopen the silent downgrade; deleting the
// with-evidence case would let the evidence reference stop threading into the
// check for that contract.
const requiredCaseIds = [
  "author-cannot-accept-own-subject",
  "record-cannot-waive-profile-independence",
  "acceptance-independence-requirement-undeclared",
  "acceptance-independence-unverifiable-without-author",
  "profile-may-waive-acceptance-independence",
  "acceptance-distinct-team-with-provider-evidence-is-valid",
  "acceptance-distinct-team-without-provider-evidence-fails-closed",
  "acceptance-distinct-team-self-grant-caught-with-provider-evidence",
  "emergency-exception-approver-cannot-rely-on-own-grant",
  "exception-approver-cannot-rely-on-own-grant",
  "emergency-access-record-cannot-waive-profile-independence",
  "emergency-access-independence-requirement-undeclared",
  "emergency-access-rejects-non-party-owner",
  "emergency-access-rejects-agent-owner",
  "emergency-access-rejects-object-owner",
  "emergency-access-accepts-non-person-owner",
  "emergency-access-accepts-service-owner",
  "emergency-access-accepts-role-owner",
  "incident-rejects-malformed-owner",
  "corrective-action-rejects-malformed-owner",
  "profile-may-waive-emergency-access-independence",
  "emergency-access-distinct-team-with-provider-evidence-is-valid",
  "emergency-access-distinct-team-without-provider-evidence-fails-closed",
  "emergency-access-distinct-team-self-grant-caught-with-provider-evidence",
  "privileged-operation-record-cannot-supply-its-own-profile",
  "unaccountable-human-cannot-record-acceptance",
  "unaccountable-human-cannot-authorize",
  "emergency-exception-requires-accountable-approver",
  "exception-approver-accountability-is-not-optional",
];

export const validateOperationalLifecycleFixtures = () => {
  const fixture = JSON.parse(readFileSync(fileURLToPath(fixtureUrl), "utf8"));
  const presentCaseIds = new Set(fixture.cases.map(({id}) => id));
  const missingCaseIds = requiredCaseIds
    .filter((id) => !presentCaseIds.has(id))
    .map((id) => `required fixture case is missing: ${id}`);
  const failures = fixture.cases.flatMap((testCase) => {
    const code = validateCase(testCase, fixture.templates);
    const valid = code === null;
    const expectedCodeMatches =
      testCase.expectedValid || code === testCase.expectedCode;
    return valid === testCase.expectedValid && expectedCodeMatches
      ? []
      : [
          `${testCase.id}: expected valid=${testCase.expectedValid} code=${testCase.expectedCode ?? "none"}, received valid=${valid} code=${code ?? "none"}`,
        ];
  });

  if (missingCaseIds.length > 0 || failures.length > 0) {
    throw new Error(
      `operational lifecycle fixture failures:\n${[...missingCaseIds, ...failures].join("\n")}`,
    );
  }

  return {cases: fixture.cases.length};
};
