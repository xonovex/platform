# plan-tdd-create: Create TDD Implementation Plan with Research

**Guideline:** Generate TDD implementation plan with RED-GREEN-COMMIT cycles as tracked tasks.

**Rationale:** Creates actionable TDD plan from research context. Each RED-GREEN-COMMIT cycle becomes a tracked task, ensuring test-first development with proper sequencing.

**Example:**

```
## TDD Plan: User Authentication

### Story 1: Basic login validation
[RED] Write test: login() validates email format
[GREEN] Implement: if (!email.includes('@')) throw
[COMMIT] "feat: validate email format in login"

### Story 2: Password requirements
[RED] Write test: password must be 8+ chars
[GREEN] Implement: if (password.length < 8) throw
[COMMIT] "feat: enforce 8-char password minimum"

### Story 3: Token generation
[RED] Write test: login() returns JWT token
[GREEN] Implement: return jwt.sign({email})
[COMMIT] "feat: generate JWT token on successful login"

### Story 4: Error handling
[RED] Write test: invalid credentials return error
[GREEN] Implement: check hash match, return Err
[COMMIT] "feat: return error for invalid credentials"

### Story 5: Refactor
[REFACTOR] Extract validation to schema, improve test coverage
```

**Techniques:**

- Read specification and extract research findings from prior analysis
- Decompose feature into discrete, testable stories
- Sequence stories: basic → properties → behavior → edge cases → validation
- Write RED test code matching project's test framework and conventions
- Define GREEN implementation guidance: minimal to pass test, hardcode acceptable
- Create sequential task dependencies: each cycle depends on previous completion
- Structure each task with test code, implementation guidance, and validation steps
- Plan for refactoring after GREEN phase for code quality
- Check granularity: independent commit, tests one thing, clear description
