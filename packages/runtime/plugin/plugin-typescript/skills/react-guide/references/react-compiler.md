# react-compiler: React Compiler Setup

Enable the React Compiler only for code that follows the Rules of React, with explicit opt-in or opt-out where needed.

## Techniques

- Enable: Vite `react({ babel: { plugins: ['babel-plugin-react-compiler'] } })` or Next.js `experimental.reactCompiler`
- Requires Rules of React: pure functions, immutable state, unconditional hooks (compiler enforces)
- Opt-in per component: `'use memo'`; opt-out: `'use no memo'`
