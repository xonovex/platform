---
name: react-guidelines
description: Trigger on `.tsx/.jsx` files with React imports, Vite config. Use when building React 19+ components. Apply for Server Components, Form Actions, new hooks, Suspense streaming, React Compiler. Keywords: React 19, RSC, Server Components, useActionState, useOptimistic, use(), Suspense, Form Actions, 'use client', 'use server', ref as prop, React Compiler.
---

# React Coding Guidelines

## Requirements

- React ≥ 19, Vite ≥ 6, Tailwind ≥ 4, Headless UI.

## React 19 Essentials

- **Server-first** - Components run on server by default; add `'use client'` only for interactivity
- **Compiler handles memoization** - Write clean code; React Compiler optimizes automatically
- **ref as prop** - Pass `ref` directly, no `forwardRef` wrapper needed
- **Context as provider** - Use `<Context value={}>` directly, not `.Provider`
- **Form Actions** - Use `useActionState` and `FormData` instead of controlled inputs
- **New hooks** - `useActionState`, `useOptimistic`, `use()`, `useFormStatus`
- **Native metadata** - Use `<title>`, `<meta>`, `<link>` anywhere, auto-hoisted to `<head>`

## Quick Reference

| Feature             | React 18                          | React 19+                         |
| ------------------- | --------------------------------- | --------------------------------- |
| Memoization         | Manual (`useMemo`, `useCallback`) | React Compiler (automatic)        |
| Forward refs        | `forwardRef()` wrapper            | `ref` as regular prop             |
| Context provider    | `<Context.Provider value={}>`     | `<Context value={}>`              |
| Form state          | Custom `useState`                 | `useActionState` hook             |
| Optimistic updates  | Manual state                      | `useOptimistic` hook              |
| Read promises       | Not possible                      | `use()` hook                      |
| Conditional context | Not possible                      | `use(Context)` after conditionals |
| Form pending        | Manual tracking                   | `useFormStatus` hook              |

## Example

```tsx
// React 19 Form with Actions
"use client";

import {useActionState} from "react";

function ContactForm() {
  const [state, formAction, isPending] = useActionState(
    async (prev, formData) => {
      const result = await submitForm(Object.fromEntries(formData));
      if (result.error) return {error: result.error};
      return {success: true};
    },
    null,
  );

  return (
    <form action={formAction}>
      <input name="email" type="email" disabled={isPending} />
      <button disabled={isPending}>
        {isPending ? "Submitting..." : "Submit"}
      </button>
      {state?.error && <p className="error">{state.error}</p>}
    </form>
  );
}

// ref as prop (no forwardRef needed)
function Input({ref, ...props}: {ref?: React.Ref<HTMLInputElement>}) {
  return <input ref={ref} {...props} />;
}

// Context as provider
const ThemeContext = createContext("light");

function App({children}) {
  return <ThemeContext value="dark">{children}</ThemeContext>;
}
```

## Essentials

- **Component design** - Small, composable; lift/minimize state; derive when possible, see [reference/component-design.md](reference/component-design.md)
- **Performance** - Let React Compiler handle memoization; manual `useMemo`/`useCallback` for effect deps only, see [reference/performance-optimization.md](reference/performance-optimization.md)
- **Rendering** - Prefer Server Components; use Suspense for streaming, see [reference/suspense-streaming.md](reference/suspense-streaming.md)
- **Accessibility** - Semantic HTML, ARIA, keyboard/focus management, see [reference/accessibility.md](reference/accessibility.md)
- **Custom hooks** - Extract reusable logic, see [reference/hooks.md](reference/hooks.md)

## Progressive Disclosure

### Guidelines

- Read [reference/component-design.md](reference/component-design.md) - When breaking down large components or managing state lifting
- Read [reference/state-management.md](reference/state-management.md) - When choosing between useState, useReducer, or Context
- Read [reference/performance-optimization.md](reference/performance-optimization.md) - When components re-render unnecessarily or performance lags
- Read [reference/hooks.md](reference/hooks.md) - When extracting reusable logic or creating custom hooks
- Read [reference/accessibility.md](reference/accessibility.md) - When adding keyboard navigation or screen reader support
- Read [reference/new-hooks.md](reference/new-hooks.md) - When using useActionState, useOptimistic, use(), or useFormStatus
- Read [reference/server-components.md](reference/server-components.md) - When building with RSC, Server Actions, or 'use server'/'use client' directives
- Read [reference/suspense-streaming.md](reference/suspense-streaming.md) - When using Suspense boundaries, streaming, or error handling
- Read [reference/react-compiler.md](reference/react-compiler.md) - When setting up React Compiler or understanding automatic memoization
- Read [reference/activity-effect-event.md](reference/activity-effect-event.md) - When using Activity component or useEffectEvent

### Migration from React 18

- Read [reference/migration-paradigm-shifts.md](reference/migration-paradigm-shifts.md) - When adapting mental model from React 18 to React 19
- Read [reference/migration-anti-patterns.md](reference/migration-anti-patterns.md) - When avoiding outdated patterns (useEffect for data, manual loading states)
- Read [reference/migration-deprecations.md](reference/migration-deprecations.md) - When migrating from React 18 or handling removed APIs
- Read [reference/migration-typescript.md](reference/migration-typescript.md) - When fixing TypeScript errors after React 19 upgrade
