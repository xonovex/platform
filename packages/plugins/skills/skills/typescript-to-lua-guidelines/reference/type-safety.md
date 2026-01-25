# type-safety: Type Safety Patterns

**Guideline:** Use strict typing with readonly properties, discriminated unions, and complete annotations.

**Rationale:** TypeScript's type system catches errors at compile time and guides TSTL. Patterns like discriminated unions provide type-safe error handling.

**Example:**
```typescript
// Discriminated union for type-safe results
type Result<T, E> =
  | { kind: 'ok'; value: T }
  | { kind: 'err'; error: E }

namespace Result {
  export function ok<T, E>(value: T): Result<T, E> {
    return { kind: 'ok', value }
  }

  export function err<T, E>(error: E): Result<T, E> {
    return { kind: 'err', error }
  }

  export function map<T, U, E>(
    result: Result<T, E>,
    fn: (t: T) => U
  ): Result<U, E> {
    if (result.kind === 'ok') {
      return ok(fn(result.value))
    }
    return result
  }
}

// Readonly properties signal immutability
interface Position {
  readonly x: number
  readonly y: number
}

interface Entity {
  readonly id: number
  readonly position: Position
  readonly velocity: readonly [number, number]  // Readonly tuple
}

// Complete generic types
type Option<T> = { kind: 'some'; value: T } | { kind: 'none' }

namespace Option {
  export function some<T>(value: T): Option<T> {
    return { kind: 'some', value }
  }

  export function none<T>(): Option<T> {
    return { kind: 'none' }
  }

  export function getOrElse<T>(opt: Option<T>, def: T): T {
    if (opt.kind === 'some') return opt.value
    return def
  }
}

// Type-safe usage
const result: Result<number, string> = Result.ok(42)
if (result.kind === 'ok') {
  console.log(result.value)  // Type is number
}
```

**Techniques:**
- Use `readonly` modifiers for immutable properties to signal intent
- Define complete interfaces with all properties typed explicitly
- Implement discriminated unions for type-safe variant handling
- Create namespaces paired with interfaces for factories and utilities
- Use union types for optional or alternative values
- Implement generic types for reusable patterns (Result<T, E>, Option<T>)
- Use factory functions to enforce proper initialization
- Provide utility functions as namespace members for convenience
