# typescript-changes: React 19 TypeScript Changes

**Guideline:** Update TypeScript types for React 19's breaking changes to refs, JSX namespace, and removed types.

**Rationale:** React 19 changes ref semantics, removes deprecated types, and scopes JSX namespace. Run codemods to automate most migrations.

**Example:**

```tsx
// Ref callbacks must not return values (except cleanup)
<div ref={current => (instance = current)} /> // ❌ Error - implicit return
<div ref={current => { instance = current; }} /> // ✅ Block body

// Cleanup functions ARE allowed
<div ref={current => {
  // setup
  return () => { /* cleanup */ };
}} />

// useRef requires argument
const ref = useRef<HTMLDivElement>(); // ❌ Error
const ref = useRef<HTMLDivElement>(null); // ✅ Required

// ReactElement props are now unknown
type Props = ReactElement['props']; // unknown (was any)
type Props = ReactElement<{ id: string }>['props']; // { id: string }

// Global JSX namespace removed - import explicitly
declare global { namespace JSX { ... } } // ❌ Removed

declare module 'react' { // ✅ Module scoped
  namespace JSX {
    interface IntrinsicElements {
      'my-element': { myProp: string };
    }
  }
}

// Import JSX type explicitly
import type { JSX } from 'react';
const element: JSX.Element = <div />;

// Removed types - use alternatives
import { ReactChild, VFC } from 'react'; // ❌ Removed

// Alternatives
type Child = ReactElement | number | string; // was ReactChild
type Component = FC<Props>; // was VFC, SFC

// Typing ref as prop
interface InputProps {
  placeholder?: string;
  ref?: React.Ref<HTMLInputElement>;
}

function Input({ placeholder, ref }: InputProps) {
  return <input placeholder={placeholder} ref={ref} />;
}

// Typing useActionState
interface FormState { error: string | null; success: boolean }

const [state, formAction, isPending] = useActionState<FormState, FormData>(
  async (prevState, formData) => {
    return { error: null, success: true };
  },
  { error: null, success: false }
);

// Typing useOptimistic
const [optimistic, addOptimistic] = useOptimistic<Message[], Omit<Message, 'pending'>>(
  messages,
  (state, newMessage) => [...state, { ...newMessage, pending: true }]
);
```

**Migration codemods:**

```bash
# Run all TypeScript codemods
npx types-react-codemod@latest preset-19 ./src

# Individual codemods
npx types-react-codemod@latest no-implicit-ref-callback-return ./src
npx types-react-codemod@latest refobject-defaults ./src
npx types-react-codemod@latest scoped-jsx ./src
```

**Techniques:**

- Update `@types/react` and `@types/react-dom` to ^19.0.0
- Run `npx types-react-codemod@latest preset-19 ./src`
- Fix ref callback implicit returns: `=> (x = y)` → `=> { x = y; }`
- Add arguments to `useRef()` calls
- Replace removed types (`ReactChild`, `VFC`)
- Import `JSX` type from `'react'` if used
