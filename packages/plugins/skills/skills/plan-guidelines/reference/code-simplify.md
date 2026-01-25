# code-simplify: Analyze Code Complexity for Simplification

**Guideline:** Identify consolidation, dead code removal, and simplification opportunities.

**Rationale:** Complexity accumulates through duplication and feature additions. Simplification reduces cognitive load, improves maintainability, and helps new contributors.

**Example:**
```typescript
// Dead code: unused export
export interface LegacyUserType { name: string }

// Duplicate validation repeated 5 times
if (!email || !email.includes('@')) throw new Error()
if (!phone || !phone.match(/^\d{10}$/)) throw new Error()

// Over-engineered: single implementation
interface IUserValidation { validate(u: any): boolean }
class UserValidation implements IUserValidation { ... }
// Just use: function validateUser(u: User): boolean

// Large interface needing simplification
interface Config { host, port, ssl, timeout, retries, backoff, circuit, cache, ... }
// Split: { server: { host, port, ssl }, network: { timeout, retries } }
```

**Techniques:**
- Scan files and build signature index for duplicate detection
- Find copy-paste duplicates: identical functions, replicated logic (>10 lines)
- Identify dead code: uncalled functions, unreachable branches, unused exports
- Detect redundancy: repeated validation, duplicate utilities, hardcoded constants
- Find over-engineering: single-implementation interfaces, unnecessary wrapper classes
- Analyze large interfaces: >5 parameters, nested config, inconsistent signatures
- Group issues by severity, prioritize by impact, and plan phased removal
