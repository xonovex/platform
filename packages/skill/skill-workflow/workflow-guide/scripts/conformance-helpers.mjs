const clone = (value) => structuredClone(value);

const requiredResultProviderOperations = [
  "resolve",
  "read",
  "publish",
  "revise",
  "relate",
  "version",
  "capabilities",
];

export const validateResultContracts = (
  resultContracts,
  expectedKinds,
  inventoryOptionalComponents,
) => {
  const actualKinds = resultContracts.map(({kind}) => kind);
  const missingKinds = expectedKinds.filter(
    (kind) => !actualKinds.includes(kind),
  );
  const duplicateKinds = actualKinds.filter(
    (kind, index) => actualKinds.indexOf(kind) !== index,
  );
  const incompleteKinds = resultContracts
    .filter(
      ({requiredSemantics}) =>
        !Array.isArray(requiredSemantics) || requiredSemantics.length < 4,
    )
    .map(({kind}) => kind);
  const inventoryContract = resultContracts.find(
    ({kind}) => kind === "Inventory",
  );
  const missingInventoryComponents = inventoryOptionalComponents.filter(
    (component) => !inventoryContract?.optionalComponents?.includes(component),
  );

  if (
    missingKinds.length > 0 ||
    duplicateKinds.length > 0 ||
    incompleteKinds.length > 0 ||
    missingInventoryComponents.length > 0
  ) {
    throw new Error(
      `result contract failure: missing=${missingKinds.join(",")} duplicates=${duplicateKinds.join(",")} incomplete=${incompleteKinds.join(",")} inventory=${missingInventoryComponents.join(",")}`,
    );
  }
};

export const validateHandle = (handle) => {
  if (handle.requiredPersistedUniversalEnvelope) return "universal-envelope";
  if (handle.workflowIdentitySource === "runtime-trace-id")
    return "runtime-identity";
  if (!handle.kind || !handle.providerContext || !handle.nativeReference)
    return "incomplete-handle";
  if (
    !Array.isArray(handle.sourceReferences) ||
    !Array.isArray(handle.availableCapabilities)
  ) {
    return "incomplete-handle";
  }
  return null;
};

export const validateProfile = (profile) => {
  if (profile.explicitProvider && !profile.providerAvailable) {
    return "explicit-provider-unavailable";
  }
  if (
    profile.includedResults.some(
      (kind) => !profile.preservedResults.includes(kind),
    )
  ) {
    return "result-erased";
  }
  const unenforced = profile.mandatoryControls.some(
    ({enforcementPoints}) =>
      !enforcementPoints.some(
        ({supported, guaranteed}) => supported && guaranteed,
      ),
  );
  return unenforced ? "no-enforcement-guarantee" : null;
};

export const validateResultProvider = (provider) => {
  if (provider.requiresUniversalSerialization || provider.requiresFile) {
    return "storage-coupled-provider";
  }
  if (!provider.nativeReferencesOpaque) return "non-opaque-reference";
  if (!provider.reconstructableAfterRestart) return "not-reconstructable";
  const missingOperations = requiredResultProviderOperations.filter(
    (operation) => !provider.operations?.includes(operation),
  );
  return missingOperations.length > 0
    ? `missing-provider-operation:${missingOperations.join(",")}`
    : null;
};

export const validateWorkflowCase = (testCase) =>
  testCase.contract === "handle"
    ? validateHandle(testCase.handle)
    : validateProfile(testCase.profile);

export const createTaskSystemProvider = ({records = []} = {}) => {
  const state = new Map(
    records.map((record) => [record.reference, clone(record)]),
  );
  let nextId = records.length + 1;

  const requireRecord = (reference) => {
    const record = state.get(reference);
    if (!record) throw new Error(`unknown native reference: ${reference}`);
    return record;
  };

  const publish = (result) => {
    const reference = `task-system:work-item:${nextId}`;
    nextId += 1;
    const record = {
      reference,
      revision: 1,
      result: clone(result),
      relationships: [],
    };
    state.set(reference, record);
    return {nativeReference: reference, nativeRevision: "1"};
  };

  const revise = (reference, expectedRevision, changes) => {
    const record = requireRecord(reference);
    if (`${record.revision}` !== `${expectedRevision}`) {
      throw new Error(`stale native revision: ${expectedRevision}`);
    }
    record.result = {...record.result, ...clone(changes)};
    record.revision += 1;
    return {nativeReference: reference, nativeRevision: `${record.revision}`};
  };

  const relate = (reference, relatedReference, relationship) => {
    const record = requireRecord(reference);
    record.relationships.push({relatedReference, relationship});
    record.revision += 1;
    return {nativeReference: reference, nativeRevision: `${record.revision}`};
  };

  return {
    capabilities: () => [...requiredResultProviderOperations],
    publish,
    read: (reference) => clone(requireRecord(reference).result),
    relate,
    resolve: (reference) => clone(requireRecord(reference)),
    revise,
    snapshot: () => [...state.values()].map(clone),
    version: (reference) => `${requireRecord(reference).revision}`,
  };
};

export const exerciseTaskSystemProvider = () => {
  const provider = createTaskSystemProvider();
  const published = provider.publish({kind: "Planning", status: "draft"});
  const revised = provider.revise(
    published.nativeReference,
    published.nativeRevision,
    {status: "approved"},
  );
  provider.relate(
    revised.nativeReference,
    "task-system:work-item:source",
    "derived-from",
  );
  const snapshot = provider.snapshot();
  const restartedProvider = createTaskSystemProvider({records: snapshot});
  const reconstructed = restartedProvider.resolve(revised.nativeReference);

  if (
    restartedProvider.read(revised.nativeReference).status !== "approved" ||
    reconstructed.relationships.length !== 1 ||
    restartedProvider.version(revised.nativeReference) !== "3" ||
    !requiredResultProviderOperations.every((operation) =>
      restartedProvider.capabilities().includes(operation),
    )
  ) {
    throw new Error("task-system provider reconstruction failure");
  }
};
