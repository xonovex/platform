# plan-create: Create Implementation Plan with Research

**Guideline:** Generate implementation plan from spec/requirements using research findings.

**Rationale:** Creates actionable plan from research context. Breaks work into tracked tasks with dependencies, enabling parallel or sequential execution as needed.

**Example:**
```
## Implementation Plan: Add OAuth2 Authentication

### Goals
- Support OAuth2 authentication
- Maintain backward compatibility with JWT
- Consolidate OAuth logic across services

### Key Decisions
- Upgrade @auth/core to v6.0
- Migrate session layer (breaks change)
- Extract OAuth to shared-auth package

### Tasks
1. [BLOCKED BY: none] Update @auth/core dependency
2. [BLOCKED BY: 1] Migrate session handling in auth-service
3. [PARALLEL TO: 2] Extract GitHub OAuth to shared-auth
4. [BLOCKED BY: 2,3] Update packages/web to use shared OAuth
5. [BLOCKED BY: 2,3,4] End-to-end testing and deployment
```

**Techniques:**
- Read specification and clarify requirements with stakeholders
- Synthesize research findings from conversation history and prior analysis
- Document key architecture decisions: technology choices, versions, rationale
- Create granular tasks with clear objectives in imperative form
- Define task dependencies: which tasks block or are blocked by others
- Identify independent tasks for parallel execution
- Structure plan sections: overview, goals, current state, findings, tasks
- Set success criteria and completion indicators for each task
