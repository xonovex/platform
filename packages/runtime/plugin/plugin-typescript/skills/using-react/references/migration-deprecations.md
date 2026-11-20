# migration-deprecations: React 19 Removed & Deprecated APIs

Migrate away from removed APIs. Upgrade to React 18.3 first to surface deprecation warnings, then run the codemods.

| Removed                    | Migration                 |
| -------------------------- | ------------------------- |
| `ReactDOM.render()`        | `createRoot().render()`   |
| `ReactDOM.hydrate()`       | `hydrateRoot()`           |
| `unmountComponentAtNode()` | `root.unmount()`          |
| `ReactDOM.findDOMNode()`   | Use refs                  |
| `propTypes`                | TypeScript                |
| `defaultProps` (functions) | ES6 default parameters    |
| String refs                | Callback refs or `useRef` |
| Legacy Context             | `createContext`           |
| `react-dom/test-utils`     | `act` from `'react'`      |

| Deprecated         | Migration            |
| ------------------ | -------------------- |
| `forwardRef`       | `ref` as prop        |
| `Context.Provider` | `<Context value={}>` |

## Migration codemods

```bash
# Run all React 19 codemods
npx codemod@latest react/19/migration-recipe

# Individual codemods
npx codemod@latest react/19/replace-reactdom-render
npx codemod@latest react/19/replace-string-ref
npx codemod@latest react/19/replace-forward-ref
npx codemod@latest react/19/replace-act-import
```
