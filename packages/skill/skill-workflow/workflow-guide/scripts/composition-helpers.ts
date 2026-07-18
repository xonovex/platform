import {resolveCapability} from "../../../skill-agent-governance/agent-governance-guide/scripts/capability-registry-helpers.mjs";
import {
  validateComposition,
  validateProfile as validateGovernanceProfile,
} from "../../../skill-agent-governance/agent-governance-guide/scripts/conformance-helpers.mjs";
import {validateProfile as validateWorkflowProfile} from "./conformance-helpers.mjs";

// A full assembled selection across both planes: the workflow and governance
// profiles, the capabilities/providers/modules/methods selected under them, and
// the mandatory cross-plane controls. Either profile may be null for a
// single-plane adoption mode.
export interface AssembledComposition {
  adoptionMode: string;
  absenceReport: string;
  workflowProfile: any;
  governanceProfile: any;
  capabilities: Record<string, string[]>;
  requiredCapabilities?: Record<string, string[]>;
  modules?: {id: string; conflictsWith: string[]}[];
  conflictResolution?: boolean;
  providers?: {port: string; available: boolean; compatible: boolean}[];
  mandatoryControls?: {
    id: string;
    enforcementPoints: {supported: boolean; guaranteed: boolean}[];
  }[];
}

// Validate a whole assembled composition as one artifact, proving a catalog
// cannot be assembled inconsistent or incomplete. This is a composition root: it
// wires the per-plane validators (validateProfile in each plane, validateComposition
// for module conflict) and the capability-registry port (resolveCapability) rather
// than re-deriving their logic, and adds only the cross-plane checks no single
// plane owns — the adoption-mode absence report, the integrated profile pairing,
// required-capability presence, provider compatibility, and mandatory cross-plane
// enforcement. It returns the first failure code, or null when the composition is
// complete and internally consistent across both planes.
export const validateAssembledComposition = (
  selection: AssembledComposition,
): string | null => {
  // The adoption mode's expected-absence report must be stated, so an
  // unselected plane is an explicit gap rather than a hidden default.
  if (
    typeof selection.absenceReport !== "string" ||
    selection.absenceReport.length === 0
  ) {
    return "absence-report-missing";
  }

  // Each plane's profile passes its own contract first.
  if (selection.workflowProfile !== null) {
    const code = validateWorkflowProfile(selection.workflowProfile);
    if (code !== null) return code;
  }
  if (selection.governanceProfile !== null) {
    const code = validateGovernanceProfile(selection.governanceProfile);
    if (code !== null) return code;
  }

  // An integrated workflow profile names its governance profile by identity; the
  // named profile must be the governance profile assembled here.
  const governanceRef = selection.workflowProfile?.governance?.profile;
  if (governanceRef !== undefined && governanceRef !== null) {
    if (selection.governanceProfile?.identity !== governanceRef) {
      return "dangling-governance-reference";
    }
  }

  // Selected governance modules compose without an unresolved conflict.
  const composition = validateComposition({
    modules: selection.modules ?? [],
    conflictResolution: selection.conflictResolution ?? false,
  });
  if (composition !== null) return composition;

  // Every selected capability resolves in the registry (shipped or
  // adopter-supplied); an unclassified reference is dangling.
  for (const [category, names] of Object.entries(
    selection.capabilities ?? {},
  )) {
    for (const name of names) {
      if (resolveCapability(category, name) === "dangling") {
        return "dangling-capability";
      }
    }
  }

  // Every capability a profile requires is actually selected.
  for (const [category, names] of Object.entries(
    selection.requiredCapabilities ?? {},
  )) {
    const selected = selection.capabilities?.[category] ?? [];
    for (const name of names) {
      if (!selected.includes(name)) return "missing-capability";
    }
  }

  // Every selected provider is available and compatible.
  for (const provider of selection.providers ?? []) {
    if (!provider.available || !provider.compatible) {
      return "incompatible-provider";
    }
  }

  // Every mandatory cross-plane control binds to a supported, guaranteed
  // enforcement point.
  for (const control of selection.mandatoryControls ?? []) {
    if (
      !control.enforcementPoints.some(
        ({supported, guaranteed}) => supported && guaranteed,
      )
    ) {
      return "unenforced-mandatory-control";
    }
  }

  return null;
};
