import {resolve} from "node:path";
import {describe, expect, it} from "vitest";
import {
  conflictingQueryOwners,
  missingValidationRoutingOwners,
  unresolvedOperationRationales,
} from "../../../src/routing-catalog.js";

// These read the live skill catalog rather than a fixture, so they state what the
// repository's own content must satisfy. The rules they apply are covered against
// built fixtures in the unit tier; what these add is the catalog itself passing.
const CATALOG_ROOT = resolve(import.meta.dirname, "../../../../../skill");

describe("the live skill catalog", () => {
  it("gives every catalog skill a validation routing scenario", () => {
    expect(missingValidationRoutingOwners(CATALOG_ROOT)).toEqual([]);
  });

  it("keeps the catalog free of claimed-twice queries and phantom operations", () => {
    expect(conflictingQueryOwners(CATALOG_ROOT)).toEqual([]);
    expect(unresolvedOperationRationales(CATALOG_ROOT)).toEqual([]);
  });
});
