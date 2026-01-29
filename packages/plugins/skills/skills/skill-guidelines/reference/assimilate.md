# assimilate: Augment Skill with Another Skill's Elements

**Guideline:** Extract elements from source skill and integrate into target skill while preserving target's structure, style, and voice.

**Rationale:** Skills evolve by importing relevant patterns from other skills while maintaining their established organization and voice.

**Example:**

```markdown
# Target skill: typescript-guidelines (existing)

- Type inference with `as const`
- Strict mode requirements
- ESM imports

# Source skill: zod-guidelines (has relevant patterns)

- Runtime validation
- Type inference from schemas
- Error handling patterns

# Extract from source: Type inference pattern

Source: "Use z.infer to extract types from schemas"
Target's context: TypeScript inference
Adapted: "Derive types from Zod schemas with z.infer for runtime-validated types"

# Insert into target's "Type Inference" section

- Type inference with `as const` for literal types
- Derive types from Zod schemas with z.infer for runtime-validated types
- Use `satisfies` operator to narrow complex unions

# Result: Target skill enhanced with validation pattern while keeping its voice
```

**Techniques:**

- Load target and source skills completely, including all detail files
- Analyze target structure: section order, bullet format, voice, tone
- Extract relevant patterns from source: guidelines, code examples, rationales
- Filter elements by aspect (validation, error-handling, testing, etc.)
- Filter by intensity percentage: 10-30% critical, 30-50% important, 50-70% comprehensive
- Rewrite source content in target's voice and style
- Insert into existing sections only, preserving structure
- Match formatting exactly: spacing, punctuation, capitalization, bullet style
- Adapt code examples to target's language and patterns
- Check for duplicates before adding to avoid redundancy
- Validate all changes preserve target's established organization and tone

## When to Apply

- Enhancing skill with patterns from related skill
- Merging complementary guidelines
- Adding missing aspects from another skill
- Updating skill with evolved patterns
