import {readFileSync} from "node:fs";
import {fileURLToPath} from "node:url";

const registryUrl = new URL(
  "../assets/capability-registry.json",
  import.meta.url,
);

// The single classification record: every selectable composition capability
// mapped to shipped or adopter-supplied. Loaded once so the reference-profile
// validator (Phase 3) and the whole-composition completeness check (Phase 5)
// consult one source instead of duplicating the classification.
export const capabilityRegistry = JSON.parse(
  readFileSync(fileURLToPath(registryUrl), "utf8"),
);

// Every entry across every category, each tagged with its category, for
// iteration by the registry validator and downstream checks.
export const registryEntries = () =>
  Object.entries(capabilityRegistry.categories).flatMap(([category, spec]) =>
    spec.entries.map((entry) => ({category, ...entry})),
  );

// The declared capability names in a category, or an empty list when the
// category is unknown.
export const capabilityNames = (category) =>
  capabilityRegistry.categories[category]?.entries.map((entry) => entry.name) ??
  [];

// Classify a capability by name within a category. Fail-closed: an unknown
// category or name resolves to "dangling" so a profile that selects it is
// reported as a dangling reference rather than silently accepted. Returns
// "shipped" | "adopter-supplied" | "dangling".
export const resolveCapability = (category, name) => {
  const spec = capabilityRegistry.categories[category];
  if (spec === undefined) return "dangling";
  const entry = spec.entries.find((candidate) => candidate.name === name);
  return entry === undefined ? "dangling" : entry.status;
};

// A capability is resolvable when the registry classifies it as shipped or
// adopter-supplied; a dangling capability is not.
export const isResolvableCapability = (category, name) =>
  resolveCapability(category, name) !== "dangling";
