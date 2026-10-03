// isRecord narrows parsed JSON to a keyed object. Arrays are objects to `typeof`
// but carry no named fields, so callers walking a decoded payload by key must
// exclude them alongside null.
export const isRecord = (value: unknown): value is Record<string, unknown> =>
  typeof value === "object" && value !== null && !Array.isArray(value);
