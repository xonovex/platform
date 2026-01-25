# deprecations: React 19 Removed & Deprecated APIs

**Guideline:** Migrate away from removed APIs; avoid deprecated patterns that will be removed in future versions.

**Rationale:** React 19 removes long-deprecated APIs. Upgrade to React 18.3 first to see deprecation warnings, then migrate.

| Removed | Migration |
|---------|-----------|
| `ReactDOM.render()` | `createRoot().render()` |
| `ReactDOM.hydrate()` | `hydrateRoot()` |
| `unmountComponentAtNode()` | `root.unmount()` |
| `ReactDOM.findDOMNode()` | Use refs |
| `propTypes` | TypeScript |
| `defaultProps` (functions) | ES6 default parameters |
| String refs | Callback refs or `useRef` |
| Legacy Context | `createContext` |
| `react-dom/test-utils` | `act` from `'react'` |

| Deprecated | Migration |
|------------|-----------|
| `forwardRef` | `ref` as prop |
| `Context.Provider` | `<Context value={}>` |

**Example:**

```tsx
// Removed: ReactDOM.render
import { render } from 'react-dom'; // ❌ Removed
render(<App />, document.getElementById('root'));

// Migration: createRoot
import { createRoot } from 'react-dom/client';
const root = createRoot(document.getElementById('root')!);
root.render(<App />);

// Removed: string refs
class Bad extends Component {
  componentDidMount() { this.refs.input.focus(); }
  render() { return <input ref="input" />; } // ❌ Removed
}

// Migration: useRef
function Good() {
  const inputRef = useRef<HTMLInputElement>(null);
  useEffect(() => { inputRef.current?.focus(); }, []);
  return <input ref={inputRef} />;
}

// Removed: propTypes
Button.propTypes = { label: PropTypes.string }; // ❌ Ignored

// Migration: TypeScript
function Button({ label }: { label: string }) {
  return <button>{label}</button>;
}

// Removed: defaultProps (function components)
function Button({ size }) { return <button className={size}>Click</button>; }
Button.defaultProps = { size: 'medium' }; // ❌ Removed

// Migration: ES6 defaults
function Button({ size = 'medium' }: { size?: string }) {
  return <button className={size}>Click</button>;
}

// Deprecated: forwardRef
const Input = forwardRef((props, ref) => <input ref={ref} {...props} />);

// Migration: ref as prop
function Input({ ref, ...props }: { ref?: Ref<HTMLInputElement> }) {
  return <input ref={ref} {...props} />;
}

// Deprecated: Context.Provider
<ThemeContext.Provider value="dark">{children}</ThemeContext.Provider>

// Migration: Context directly
<ThemeContext value="dark">{children}</ThemeContext>

// Removed: react-dom/test-utils
import { act } from 'react-dom/test-utils'; // ❌ Removed

// Migration: import from react
import { act } from 'react';
```

**Migration codemods:**

```bash
# Run all React 19 codemods
npx codemod@latest react/19/migration-recipe

# Individual codemods
npx codemod@latest react/19/replace-reactdom-render
npx codemod@latest react/19/replace-string-ref
npx codemod@latest react/19/replace-forward-ref
npx codemod@latest react/19/replace-act-import
```

**Techniques:**
- Removed APIs: ReactDOM.render, ReactDOM.hydrate, findDOMNode, string refs, propTypes
- createRoot: Use for client-side rendering instead of deprecated ReactDOM.render
- useRef/Callback refs: Replace string refs like ref="fieldName"
- TypeScript: Replace propTypes validation with type-safe function parameters
- defaultProps: Use ES6 default parameters instead of .defaultProps
- forwardRef: Migrate to ref as prop for better type safety
- Context: Use Context directly instead of Context.Provider wrapper
- Act import: Import from 'react' not 'react-dom/test-utils'
- Codemods: Run react/19/migration-recipe to automate common migrations
