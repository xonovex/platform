module.exports = {
  target: (name) => {
    // Hold typescript on the 6.x line. typescript-eslint declares a
    // `typescript >=4.8.4 <6.1.0` peer range, and its ts-api-utils dependency
    // throws "Cannot read properties of undefined (reading 'Intrinsic')"
    // against TypeScript 7, which fails every ts-lint task.
    if (name === "typescript") return "minor";

    // Hold npm on the 11.x line. npm 12 resolves the optional
    // `typescript >=5.0.0` peer of @vitest/eslint-plugin to the newest release
    // instead of reusing the pinned typescript, so it demands a second
    // typescript and every moon task that installs dependencies fails with
    // ERESOLVE.
    if (name === "npm") return "minor";

    return "latest";
  },
};
