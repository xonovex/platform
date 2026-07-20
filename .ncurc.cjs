// Vite 8.1 cannot resolve package-name tsconfig `extends` values through its
// transform, and TypeScript-ESLint supports TypeScript versions below 6.1.
module.exports = {
  target: (name) =>
    name === "vite" || name === "typescript" ? "patch" : "latest",
};
