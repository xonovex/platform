import {checkIndependence} from "./independence-helpers.mjs";

const hasValue = (value) =>
  value !== undefined && value !== null && value !== "";

const hasEveryValue = (value, fields) =>
  fields.every((field) => hasValue(value?.[field]));

const isNonEmptyArray = (value) => Array.isArray(value) && value.length > 0;

const timestamp = (value) => Date.parse(value ?? "");

const isValidTimeRange = (start, end) =>
  Number.isFinite(timestamp(start)) &&
  Number.isFinite(timestamp(end)) &&
  timestamp(start) < timestamp(end);

const sameStringSet = (left, right) =>
  Array.isArray(left) &&
  Array.isArray(right) &&
  left.length === right.length &&
  left.every((value) => right.includes(value));

const hasRequiredKinds = (values, kinds) =>
  Array.isArray(values) && kinds.every((kind) => values.includes(kind));

const evidenceMatches = (authorized, current) =>
  Array.isArray(authorized) &&
  Array.isArray(current) &&
  authorized.length > 0 &&
  authorized.every(({reference, revision}) =>
    current.some(
      (candidate) =>
        candidate.reference === reference && candidate.revision === revision,
    ),
  );

export const validateAcceptance = ({acceptance, profile}) => {
  if (
    !hasEveryValue(acceptance?.subject, ["reference", "revision"]) ||
    !hasEveryValue(acceptance?.target, ["reference", "revision"]) ||
    !hasEveryValue(acceptance?.policy, ["identity", "version"])
  ) {
    return "acceptance-not-exactly-bound";
  }

  if (acceptance.mode === "evidence") {
    if (
      !isNonEmptyArray(acceptance.evidence) ||
      acceptance.evidence.some(
        (item) =>
          !hasEveryValue(item, ["reference", "revision", "origin", "outcome"]),
      )
    ) {
      return "acceptance-evidence-incomplete";
    }
    if (hasValue(acceptance.decision) || hasValue(acceptance.actor)) {
      return "evidence-assembly-claims-acceptance";
    }
    if (
      acceptance.assistance &&
      validateAgentAssistance(acceptance.assistance) !== null
    ) {
      return "invalid-acceptance-assistance";
    }
    return null;
  }

  if (acceptance.mode === "decision") {
    if (
      !["accept", "accept-with-conditions", "reject"].includes(
        acceptance.decision,
      ) ||
      !hasEveryValue(acceptance.actor, ["identity", "role"]) ||
      acceptance.actor.type !== "human" ||
      acceptance.actor.accountable !== true ||
      !acceptance.evidencePackageReference ||
      !isValidTimeRange(acceptance.decidedAt, acceptance.expiresAt)
    ) {
      return "human-acceptance-required";
    }
    return checkIndependence({
      required: profile?.independence?.acceptance,
      decider: acceptance.actor.identity,
      author: acceptance.subjectAuthor,
      failureCode: "acceptance-independence-failed",
    });
  }

  return "unknown-acceptance-mode";
};

export const validateAuthorization = ({authorization, request}) => {
  if (
    !hasEveryValue(authorization?.subject, ["reference", "revision"]) ||
    !hasEveryValue(authorization?.target, ["reference", "revision"]) ||
    !hasEveryValue(authorization?.actor, ["identity", "role"]) ||
    !hasEveryValue(authorization?.policy, ["identity", "version"]) ||
    !hasValue(authorization.action) ||
    !isNonEmptyArray(authorization.evidence) ||
    !isValidTimeRange(authorization.issuedAt, authorization.expiresAt) ||
    authorization.decision !== "allow"
  ) {
    return "authorization-contract-incomplete";
  }
  if (
    authorization.actor.type !== "human" ||
    authorization.actor.accountable !== true
  ) {
    return "human-authorization-required";
  }
  if (
    authorization.subject.reference !== request?.subject?.reference ||
    authorization.subject.revision !== request?.subject?.revision
  ) {
    return "authorization-subject-drift";
  }
  if (
    authorization.target.reference !== request?.target?.reference ||
    authorization.target.revision !== request?.target?.revision
  ) {
    return "authorization-target-drift";
  }
  if (authorization.action !== request?.action) {
    return "authorization-action-mismatch";
  }
  if (
    authorization.policy.identity !== request?.policy?.identity ||
    authorization.policy.version !== request?.policy?.version
  ) {
    return "authorization-policy-drift";
  }
  if (!evidenceMatches(authorization.evidence, request?.evidence)) {
    return "authorization-stale-evidence";
  }
  if (
    !Number.isFinite(timestamp(request?.now)) ||
    timestamp(request.now) < timestamp(authorization.issuedAt) ||
    timestamp(request.now) >= timestamp(authorization.expiresAt) ||
    authorization.evidence.some(
      ({validUntil}) =>
        hasValue(validUntil) && timestamp(request.now) >= timestamp(validUntil),
    )
  ) {
    return "authorization-expired";
  }
  return null;
};

export const validateEmergencyAccess = ({access, request}) => {
  if (
    !["exception", "break-glass"].includes(access?.kind) ||
    !hasEveryValue(access, [
      "owner",
      "rationale",
      "evidenceReference",
      "start",
      "expiresAt",
    ]) ||
    !hasEveryValue(access.approver, ["identity", "role"]) ||
    access.approver.type !== "human" ||
    !isNonEmptyArray(access.scope?.actions) ||
    !isNonEmptyArray(access.scope?.targets) ||
    !isNonEmptyArray(access.compensatingControls) ||
    access.postUseReviewRequired !== true ||
    !isValidTimeRange(access.start, access.expiresAt)
  ) {
    return "emergency-access-contract-incomplete";
  }
  if (
    !access.scope.actions.includes(request?.action) ||
    !access.scope.targets.includes(request?.target?.reference)
  ) {
    return "emergency-access-out-of-scope";
  }
  if (
    !Number.isFinite(timestamp(request?.now)) ||
    timestamp(request.now) < timestamp(access.start) ||
    timestamp(request.now) >= timestamp(access.expiresAt)
  ) {
    return "emergency-access-expired";
  }
  if (
    access.kind === "break-glass" &&
    (!hasEveryValue(access, [
      "emergencyReason",
      "accessEvidenceReference",
      "notificationReference",
      "containment",
      "revocationPlan",
    ]) ||
      access.explicitInvocation !== true)
  ) {
    return "break-glass-contract-incomplete";
  }
  return null;
};

export const validatePrivilegedOperation = (operation) => {
  if (
    ![
      "integration",
      "transition",
      "release",
      "data-deletion",
      "retirement",
    ].includes(operation?.intent) ||
    !operation.request
  ) {
    return "unknown-privileged-intent";
  }
  if (operation.intent !== operation.request.action) {
    return "privileged-intent-mismatch";
  }
  if (
    operation.ordinaryToolMutation === true ||
    operation.capability !== "external-enforcement"
  ) {
    return "ordinary-tool-bypass";
  }
  if (
    !hasEveryValue(operation.enforcement, ["kind", "nativeReference"]) ||
    operation.enforcement.blocking !== true ||
    operation.enforcement.governedActorCanBypass !== false ||
    operation.enforcement.failurePolicy !== "fail-closed"
  ) {
    return "external-enforcement-inadequate";
  }

  const authorityCode = operation.emergencyAccess
    ? validateEmergencyAccess({
        access: operation.emergencyAccess,
        request: operation.request,
      })
    : validateAuthorization({
        authorization: operation.authorization,
        request: operation.request,
      });
  if (authorityCode !== null) return authorityCode;

  if (
    !isValidTimeRange(
      operation.credentials?.issuedAt,
      operation.credentials?.expiresAt,
    ) ||
    timestamp(operation.request.now) <
      timestamp(operation.credentials?.issuedAt) ||
    timestamp(operation.request.now) >=
      timestamp(operation.credentials?.expiresAt) ||
    !sameStringSet(
      operation.credentials?.requiredScopes,
      operation.credentials?.effectiveScopes,
    )
  ) {
    return "least-privilege-credentials-invalid";
  }
  if (
    !hasEveryValue(operation.outcome, [
      "status",
      "nativeReference",
      "verificationReference",
    ])
  ) {
    return "privileged-outcome-incomplete";
  }
  if (
    operation.irreversible !== true &&
    !operation.rollbackPlanReference &&
    !operation.recoveryPlanReference
  ) {
    return "rollback-or-recovery-missing";
  }
  if (operation.irreversible === true && !operation.irreversibilityHandling) {
    return "irreversibility-handling-missing";
  }
  return null;
};

export const validateTransition = (transition) => {
  const requiredAreas = [
    "data",
    "users",
    "providers",
    "flags",
    "training",
    "support",
    "resilience",
  ];
  if (
    !["plan", "execute", "verify", "rollback"].includes(transition?.mode) ||
    !hasEveryValue(transition?.source, ["reference", "revision"]) ||
    !hasEveryValue(transition?.target, ["reference", "revision"]) ||
    !hasRequiredKinds(transition?.plan?.areas, requiredAreas) ||
    !transition.plan.rollbackTriggers?.length
  ) {
    return "transition-contract-incomplete";
  }
  if (
    ["execute", "rollback"].includes(transition.mode) &&
    transition.operation?.intent !== "transition"
  ) {
    return "transition-intent-mismatch";
  }
  if (
    ["execute", "rollback"].includes(transition.mode) &&
    validatePrivilegedOperation(transition.operation) !== null
  ) {
    return validatePrivilegedOperation(transition.operation);
  }
  if (
    ["verify", "rollback"].includes(transition.mode) &&
    !isNonEmptyArray(transition.verification)
  ) {
    return "transition-verification-missing";
  }
  return null;
};

export const validateRelease = (release) => {
  if (
    !hasEveryValue(release?.artifact, ["reference", "digest"]) ||
    !hasEveryValue(release?.protectedEnvironment, ["reference", "revision"]) ||
    !hasValue(release.approvalReference) ||
    release.automation?.controlled !== true ||
    !hasValue(release.automation.nativeReference)
  ) {
    return "release-contract-incomplete";
  }
  if (release.operation?.intent !== "release") {
    return "release-intent-mismatch";
  }
  const operationCode = validatePrivilegedOperation(release.operation);
  if (operationCode !== null) return operationCode;
  if (
    release.operation.outcome.status === "failed" &&
    !release.rollbackResultReference &&
    !release.recoveryResultReference
  ) {
    return "release-failure-unrecovered";
  }
  if (
    release.operation.outcome.status === "failed" &&
    release.reportedStatus === "success"
  ) {
    return "failed-release-reported-success";
  }
  return null;
};

export const validateObservation = (observation) => {
  const requiredSignals = [
    "monitoring",
    "user",
    "security",
    "ai",
    "cost",
    "accessibility",
    "delivery-outcome",
  ];
  if (
    !hasEveryValue(observation?.subject, ["reference", "revision"]) ||
    !isValidTimeRange(observation?.window?.start, observation?.window?.end) ||
    !isNonEmptyArray(observation.baselines) ||
    !isNonEmptyArray(observation.signals) ||
    !hasRequiredKinds(
      observation.signals.map(({kind}) => kind),
      requiredSignals,
    )
  ) {
    return "observation-coverage-incomplete";
  }
  if (
    observation.signals.some(
      (signal) =>
        !hasEveryValue(signal, [
          "kind",
          "nativeReference",
          "observedAt",
          "freshness",
        ]),
    )
  ) {
    return "observation-evidence-incomplete";
  }
  if (
    observation.assistance &&
    validateAgentAssistance(observation.assistance) !== null
  ) {
    return "invalid-observation-assistance";
  }
  return null;
};

export const validateIncident = (incident) => {
  if (
    !hasEveryValue(incident, [
      "reference",
      "severity",
      "owner",
      "currentState",
    ]) ||
    !hasValue(incident.scope) ||
    !isNonEmptyArray(incident.timeline) ||
    !isNonEmptyArray(incident.actions) ||
    !hasValue(incident.reportingApplicability?.status) ||
    !isNonEmptyArray(incident.reportingApplicability?.sources)
  ) {
    return "incident-contract-incomplete";
  }
  if (
    ["applicable", "not-applicable"].includes(
      incident.reportingApplicability.status,
    ) &&
    (!hasEveryValue(incident.reportingApplicability.reviewer, [
      "identity",
      "qualification",
    ]) ||
      incident.reportingApplicability.reviewer.type !== "human" ||
      !isNonEmptyArray(incident.reportingApplicability.sources))
  ) {
    return "qualified-reporting-review-required";
  }
  if (incident.severityChanged === true && !incident.escalationReference) {
    return "incident-escalation-missing";
  }
  if (
    incident.assistance &&
    validateAgentAssistance(incident.assistance) !== null
  ) {
    return "invalid-incident-assistance";
  }
  if (
    incident.currentState === "closed" &&
    (!incident.containmentVerified ||
      !incident.recoveryVerified ||
      !incident.postIncidentReviewReference)
  ) {
    return "incident-closure-incomplete";
  }
  return null;
};

export const validateCorrectiveAction = (correctiveAction) => {
  if (
    !hasEveryValue(correctiveAction, [
      "sourceReference",
      "sourceRevision",
      "cause",
      "action",
      "owner",
      "dueState",
    ]) ||
    !hasValue(correctiveAction.verification?.nativeReference) ||
    !hasValue(correctiveAction.effectiveness?.status)
  ) {
    return "corrective-action-contract-incomplete";
  }
  if (
    correctiveAction.status === "closed" &&
    (correctiveAction.verification.outcome !== "pass" ||
      correctiveAction.effectiveness.status !== "effective" ||
      !correctiveAction.learningReference ||
      !correctiveAction.residualRiskDisposition)
  ) {
    return "corrective-action-closure-incomplete";
  }
  return null;
};

export const validateRetirement = (retirement) => {
  const supportedKinds = [
    "model",
    "data",
    "credential",
    "feature",
    "api",
    "infrastructure",
    "dependency",
    "provider-configuration",
  ];
  if (
    !isNonEmptyArray(retirement?.resources) ||
    retirement.resources.some(
      (resource) =>
        !supportedKinds.includes(resource.kind) ||
        !hasEveryValue(resource, [
          "kind",
          "reference",
          "revision",
          "action",
          "evidenceReference",
        ]),
    )
  ) {
    return "retirement-scope-incomplete";
  }
  if (retirement.operation?.intent !== "retirement") {
    return "retirement-intent-mismatch";
  }
  const operationCode = validatePrivilegedOperation(retirement.operation);
  if (operationCode !== null) return operationCode;
  if (
    retirement.resources.some(
      ({action, verified}) => action !== "retain" && verified !== true,
    )
  ) {
    return "retirement-verification-missing";
  }
  if (
    retirement.resources.some(
      ({kind, action, deletionEvidenceReference, holdDisposition}) =>
        kind === "data" &&
        action === "delete" &&
        (!deletionEvidenceReference || !holdDisposition),
    )
  ) {
    return "data-deletion-evidence-missing";
  }
  if (!hasValue(retirement.residualRiskDisposition)) {
    return "retirement-residual-risk-missing";
  }
  return null;
};

export function validateAgentAssistance(assistance) {
  if (
    !["agent", "model"].includes(assistance?.origin) ||
    assistance.authority !== "advisory" ||
    !isNonEmptyArray(assistance.sourceReferences)
  ) {
    return "assistance-contract-incomplete";
  }
  if (assistance.fabricatesHumanAuthorization === true) {
    return "fabricated-human-authorization";
  }
  if (assistance.regulatoryConclusion === true) {
    return "fabricated-regulatory-conclusion";
  }
  return null;
}
